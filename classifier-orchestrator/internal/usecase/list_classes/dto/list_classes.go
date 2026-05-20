package dto

import "time"

type Class struct {
	ID              int32     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	PositiveGoldens []Golden  `json:"positiveGoldens"`
	NegativeGoldens []Golden  `json:"negativeGoldens"`
}

type Golden struct {
	ID   int32  `json:"id"`
	Text string `json:"text"`
}

type ListClassesResponse []Class
