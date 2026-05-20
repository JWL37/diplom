package dto

import "time"

type UpdateClassStatusRequest struct {
	ID     int32  `json:"-"`
	Status string `json:"status"`
}

type UpdateClassStatusResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
