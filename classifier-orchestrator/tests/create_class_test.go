package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultBaseURL     = "http://localhost:8080"
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/classifier_orchestrator?sslmode=disable"
)

type goldenRequest struct {
	Text string `json:"text"`
}

type createClassRequest struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	PositiveGoldens []goldenRequest `json:"positiveGoldens"`
	NegativeGoldens []goldenRequest `json:"negativeGoldens"`
}

type createClassResponse struct {
	ID int32 `json:"id"`
}

type updateClassStatusRequest struct {
	Status string `json:"status"`
}

type updateClassStatusResponse struct {
	ID     int32  `json:"id"`
	Status string `json:"status"`
}

func TestCreateClassSavesData(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("set INTEGRATION_TESTS=1 to run integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suffix := time.Now().UnixNano()
	reqBody := createClassRequest{
		Name:        fmt.Sprintf("integration_class_%d", suffix),
		Description: fmt.Sprintf("integration class description %d", suffix),
		PositiveGoldens: []goldenRequest{
			{Text: fmt.Sprintf("positive integration golden %d", suffix)},
		},
		NegativeGoldens: []goldenRequest{
			{Text: fmt.Sprintf("negative integration golden %d", suffix)},
		},
	}

	createdClass := createClass(t, ctx, reqBody)
	if createdClass.ID <= 0 {
		t.Fatalf("expected positive class id, got %d", createdClass.ID)
	}

	db := connectDB(t, ctx)
	defer db.Close()

	assertClassSaved(t, ctx, db, createdClass.ID, reqBody)
	assertPositiveGoldenSaved(t, ctx, db, createdClass.ID, reqBody.PositiveGoldens[0].Text)
	negativeGoldenID := assertNegativeGoldenSaved(t, ctx, db, reqBody.NegativeGoldens[0].Text)
	assertNegativeGoldenLinked(t, ctx, db, createdClass.ID, negativeGoldenID)
}

func TestUpdateClassStatusSavesData(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("set INTEGRATION_TESTS=1 to run integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suffix := time.Now().UnixNano()
	reqBody := createClassRequest{
		Name:        fmt.Sprintf("integration_status_class_%d", suffix),
		Description: fmt.Sprintf("integration status class description %d", suffix),
		PositiveGoldens: []goldenRequest{
			{Text: fmt.Sprintf("positive status integration golden %d", suffix)},
		},
		NegativeGoldens: []goldenRequest{
			{Text: fmt.Sprintf("negative status integration golden %d", suffix)},
		},
	}

	createdClass := createClass(t, ctx, reqBody)
	if createdClass.ID <= 0 {
		t.Fatalf("expected positive class id, got %d", createdClass.ID)
	}

	db := connectDB(t, ctx)
	defer db.Close()

	activeResp := updateClassStatus(t, ctx, createdClass.ID, "ACTIVE")
	if activeResp.ID != createdClass.ID {
		t.Fatalf("expected class id %d, got %d", createdClass.ID, activeResp.ID)
	}
	if activeResp.Status != "ACTIVE" {
		t.Fatalf("expected response status ACTIVE, got %q", activeResp.Status)
	}
	assertClassStatus(t, ctx, db, createdClass.ID, "ACTIVE")

	draftResp := updateClassStatus(t, ctx, createdClass.ID, "DRAFT")
	if draftResp.ID != createdClass.ID {
		t.Fatalf("expected class id %d, got %d", createdClass.ID, draftResp.ID)
	}
	if draftResp.Status != "DRAFT" {
		t.Fatalf("expected response status DRAFT, got %q", draftResp.Status)
	}
	assertClassStatus(t, ctx, db, createdClass.ID, "DRAFT")
}

func createClass(t *testing.T, ctx context.Context, reqBody createClassRequest) createClassResponse {
	t.Helper()

	payload, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	baseURL := strings.TrimRight(getEnv("CLASSIFIER_ORCHESTRATOR_BASE_URL", defaultBaseURL), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/v1/classes", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to call create class API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, resp.StatusCode, string(body))
	}

	var created createClassResponse
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return created
}

func updateClassStatus(t *testing.T, ctx context.Context, classID int32, status string) updateClassStatusResponse {
	t.Helper()

	payload, err := json.Marshal(updateClassStatusRequest{Status: status})
	if err != nil {
		t.Fatalf("failed to marshal update status request body: %v", err)
	}

	baseURL := strings.TrimRight(getEnv("CLASSIFIER_ORCHESTRATOR_BASE_URL", defaultBaseURL), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/api/v1/classes/%d/status", baseURL, classID), bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to build update status request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to call update class status API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read update status response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, resp.StatusCode, string(body))
	}

	var updated updateClassStatusResponse
	if err := json.Unmarshal(body, &updated); err != nil {
		t.Fatalf("failed to decode update status response body: %v", err)
	}

	return updated
}

func connectDB(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	db, err := pgxpool.New(ctx, getEnv("TEST_DATABASE_URL", defaultDatabaseURL))
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping database: %v", err)
	}

	return db
}

func assertClassSaved(t *testing.T, ctx context.Context, db *pgxpool.Pool, classID int32, reqBody createClassRequest) {
	t.Helper()

	var name string
	var description string
	var status string
	err := db.QueryRow(ctx, `
		SELECT name, COALESCE(description, ''), status::text
		FROM prediction_classes
		WHERE id = $1
	`, classID).Scan(&name, &description, &status)
	if err != nil {
		t.Fatalf("failed to query created class: %v", err)
	}
	if name != reqBody.Name {
		t.Fatalf("expected class name %q, got %q", reqBody.Name, name)
	}
	if description != reqBody.Description {
		t.Fatalf("expected class description %q, got %q", reqBody.Description, description)
	}
	if status != "DRAFT" {
		t.Fatalf("expected class status DRAFT, got %q", status)
	}
}

func assertPositiveGoldenSaved(t *testing.T, ctx context.Context, db *pgxpool.Pool, classID int32, text string) {
	t.Helper()

	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM positive_goldens
		WHERE class_id = $1 AND text_content = $2
	`, classID, text).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query positive golden: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 positive golden, got %d", count)
	}
}

func assertNegativeGoldenSaved(t *testing.T, ctx context.Context, db *pgxpool.Pool, text string) int32 {
	t.Helper()

	var id int32
	err := db.QueryRow(ctx, `
		SELECT id
		FROM negative_goldens
		WHERE text_content = $1
	`, text).Scan(&id)
	if err != nil {
		t.Fatalf("failed to query negative golden: %v", err)
	}

	return id
}

func assertNegativeGoldenLinked(t *testing.T, ctx context.Context, db *pgxpool.Pool, classID int32, negativeGoldenID int32) {
	t.Helper()

	var count int
	err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM class_negative_goldens
		WHERE class_id = $1 AND negative_golden_id = $2
	`, classID, negativeGoldenID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query negative golden link: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 negative golden link, got %d", count)
	}
}

func assertClassStatus(t *testing.T, ctx context.Context, db *pgxpool.Pool, classID int32, expectedStatus string) {
	t.Helper()

	var status string
	err := db.QueryRow(ctx, `
		SELECT status::text
		FROM prediction_classes
		WHERE id = $1
	`, classID).Scan(&status)
	if err != nil {
		t.Fatalf("failed to query class status: %v", err)
	}
	if status != expectedStatus {
		t.Fatalf("expected class status %q, got %q", expectedStatus, status)
	}
}

func getEnv(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	return value
}
