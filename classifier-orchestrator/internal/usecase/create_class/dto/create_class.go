package dto

type CreateClassRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateClassResponse struct {
	ID int32 `json:"id"`
}
