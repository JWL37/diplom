package dto

type Class struct {
	ID          int32
	Name        string
	Description string
}

type ListClassesOutput []Class
