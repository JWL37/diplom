package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"classifier-orchestrator/internal/app/database"
	"classifier-orchestrator/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type goldenRow struct {
	Text      string
	Label     string
	ClassName string
}

type importStats struct {
	classes       int
	positiveAdded int64
	duplicates    int
}

func main() {
	datasetPath := flag.String("dataset", "../dataset.txt", "path to classifier dataset")
	limit := flag.Int("limit", 10, "max positive goldens per class")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rows, duplicates, err := readGoldens(*datasetPath, *limit)
	if err != nil {
		logger.Error("Failed to read dataset", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if len(rows) == 0 {
		logger.Warn("Dataset contains no rows")
		return
	}

	cfg := config.LoadConfig()
	db, err := database.NewPostgresDB(ctx, logger, cfg)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	stats, err := seedGoldens(ctx, db, rows, duplicates)
	if err != nil {
		logger.Error("Failed to seed database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info(
		"Goldens import completed",
		slog.Int("classes", stats.classes),
		slog.Int64("positive_added", stats.positiveAdded),
		slog.Int("duplicates", stats.duplicates),
	)
}

func readGoldens(path string, limit int) ([]goldenRow, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("open dataset: %w", err)
	}
	defer file.Close()

	if limit <= 0 {
		return nil, 0, fmt.Errorf("limit must be positive")
	}

	rows := make([]goldenRow, 0)
	counts := make(map[string]int)
	seen := make(map[string]struct{})
	duplicates := 0

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		labelsPart, text, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		label := extractFirstLabel(labelsPart)
		if label == "" {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		key := label + "\x00" + text
		if _, ok := seen[key]; ok {
			duplicates++
			continue
		}
		seen[key] = struct{}{}

		if counts[label] >= limit {
			continue
		}
		counts[label]++

		rows = append(rows, goldenRow{
			Text:      text,
			Label:     label,
			ClassName: translateLabel(label),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("scan dataset: %w", err)
	}

	return rows, duplicates, nil
}

func extractFirstLabel(labelsPart string) string {
	firstLabel := strings.Split(labelsPart, ",")[0]
	firstLabel = strings.TrimSpace(firstLabel)
	firstLabel = strings.TrimPrefix(firstLabel, "__label__")
	return strings.ToUpper(firstLabel)
}

func translateLabel(label string) string {
	switch label {
	case "NORMAL":
		return "Нормальный комментарий"
	case "INSULT":
		return "Оскорбление"
	case "THREAT":
		return "Угроза"
	case "OBSCENITY":
		return "Нецензурная лексика"
	default:
		return label
	}
}

func seedGoldens(ctx context.Context, db *pgxpool.Pool, rows []goldenRow, duplicates int) (importStats, error) {
	const upsertClass = `
		INSERT INTO prediction_classes (name, description)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE
		SET description = EXCLUDED.description,
			updated_at = NOW()
		RETURNING id
	`
	const insertPositiveGolden = `
		INSERT INTO positive_goldens (class_id, text_content)
		VALUES ($1, $2)
		ON CONFLICT (text_content) DO NOTHING
	`

	tx, err := db.Begin(ctx)
	if err != nil {
		return importStats{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	byClass := make(map[string][]goldenRow)
	for _, row := range rows {
		byClass[row.ClassName] = append(byClass[row.ClassName], row)
	}

	classNames := make([]string, 0, len(byClass))
	for className := range byClass {
		classNames = append(classNames, className)
	}
	sort.Strings(classNames)

	stats := importStats{
		classes:    len(classNames),
		duplicates: duplicates,
	}

	for _, className := range classNames {
		var classID int32
		description := fmt.Sprintf("Импортировано из dataset.txt, исходная метка: %s", byClass[className][0].Label)
		if err := tx.QueryRow(ctx, upsertClass, className, description).Scan(&classID); err != nil {
			return importStats{}, fmt.Errorf("upsert class %s: %w", className, err)
		}

		for _, row := range byClass[className] {
			tag, err := tx.Exec(ctx, insertPositiveGolden, classID, row.Text)
			if err != nil {
				return importStats{}, fmt.Errorf("insert positive golden for class %s: %w", className, err)
			}
			stats.positiveAdded += tag.RowsAffected()
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return importStats{}, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}
