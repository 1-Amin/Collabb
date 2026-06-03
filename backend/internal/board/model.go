package board

import "time"

type Board struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Columns   []Column  `json:"columns"`
}

type Column struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	Cards     []Card    `json:"cards"`
}

type Card struct {
	ID          string    `json:"id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
}
