package dto

type UpdateClassRequest struct {
	ID              int32         `json:"-"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	PositiveGoldens []GoldenInput `json:"positiveGoldens"`
	NegativeGoldens []GoldenInput `json:"negativeGoldens"`
}

type GoldenInput struct {
	Text string `json:"text"`
}

type UpdateClassResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
