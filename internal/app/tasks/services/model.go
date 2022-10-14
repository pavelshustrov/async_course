package services

type Task struct {
	PublicID    string `json:"public_id"`
	Description string `json:"description"`
	Status      status
	Price       Price
	AssigneeID  string `json:"assignee_id"`
}

type Price struct {
	Fee   int
	Award int
}

type User struct {
	PublicID string `json:"public_id"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

type status string

var (
	StatusOpen   status = "OPEN"
	StatusClosed status = "CLOSED"
)
