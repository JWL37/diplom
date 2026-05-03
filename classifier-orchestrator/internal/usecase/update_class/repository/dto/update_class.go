package dto

type UpdateClassInput struct {
	ID              int32
	Name            string
	Description     string
	PositiveGoldens []GoldenInput
	NegativeGoldens []GoldenInput
}

type GoldenInput struct {
	Text string
}

type UpdateClassOutput struct {
	ID          int32
	Name        string
	Description string
}
