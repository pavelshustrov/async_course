package services

type AuditRecord struct {
	ID           int
	AccountID    int
	PublicID     string
	EventID      string
	Debit        int
	Credit       int
	TaskID       string
	BillingCycle string
}

type Account struct {
	ID       int
	PublicID string
	Amount   int
	UserID   string
}

type Task struct {
	PublicID   string `json:"public_id"`
	PriceFee   int    `json:"price_fee"`
	PriceAward int    `json:"price_award"`
	Status     status `json:"status"`
	AssigneeID string `json:"assignee_id"`
}

type TaskClosed struct {
	PublicID   string `json:"public_id"`
	AssigneeID string `json:"assignee_id"`
	EventID    string `json:"event_id"`
}

type User struct {
	PublicID string `json:"public_id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type status string

var (
	StatusOpen   status = "OPEN"
	StatusClosed status = "CLOSED"
)
