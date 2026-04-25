package dto

type CreateClassInput struct {
	Name        string
	Description string
}

type CreateClassOutput struct {
	ID int32
}
