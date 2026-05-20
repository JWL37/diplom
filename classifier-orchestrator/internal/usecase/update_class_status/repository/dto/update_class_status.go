package dto

import "time"

type UpdateClassStatusInput struct {
	ID     int32
	Status string
}

type UpdateClassStatusOutput struct {
	ID          int32
	Name        string
	Description string
	Status      string
	UpdatedAt   time.Time
}
