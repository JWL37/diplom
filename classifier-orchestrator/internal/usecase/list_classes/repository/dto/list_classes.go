package dto

import "time"

type Class struct {
	ID              int32
	Name            string
	Description     string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PositiveGoldens []Golden
	NegativeGoldens []Golden
}

type Golden struct {
	ID   int32  `json:"id"`
	Text string `json:"text"`
}

type ListClassesOutput []Class
