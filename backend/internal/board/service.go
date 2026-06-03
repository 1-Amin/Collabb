package board

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct{ db *pgxpool.Pool }

func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

// ── Boards ──────────────────────────────────────────────────────────────────

func (s *Service) CreateBoard(ctx context.Context, ownerID, title string) (*Board, error) {
	b := &Board{OwnerID: ownerID, Title: title}
	err := s.db.QueryRow(ctx,
		`INSERT INTO boards (owner_id, title) VALUES ($1, $2)
		 RETURNING id, owner_id, title, created_at`,
		ownerID, title,
	).Scan(&b.ID, &b.OwnerID, &b.Title, &b.CreatedAt)
	return b, err
}

func (s *Service) ListBoards(ctx context.Context, ownerID string) ([]Board, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, owner_id, title, created_at FROM boards WHERE owner_id = $1 ORDER BY created_at`,
		ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	boards := []Board{} // never return JSON null
	for rows.Next() {
		var b Board
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Title, &b.CreatedAt); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

func (s *Service) GetBoard(ctx context.Context, boardID string) (*Board, error) {
	b := &Board{}
	err := s.db.QueryRow(ctx,
		`SELECT id, owner_id, title, created_at FROM boards WHERE id = $1`, boardID,
	).Scan(&b.ID, &b.OwnerID, &b.Title, &b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("board not found: %w", err)
	}
	cols, err := s.getColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	b.Columns = cols
	return b, nil
}

// ── Columns ─────────────────────────────────────────────────────────────────

func (s *Service) getColumns(ctx context.Context, boardID string) ([]Column, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, board_id, title, position, created_at
		 FROM columns WHERE board_id = $1 ORDER BY position`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := []Column{}
	for rows.Next() {
		var c Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Title, &c.Position, &c.CreatedAt); err != nil {
			return nil, err
		}
		cards, err := s.getCards(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Cards = cards
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (s *Service) CreateColumn(ctx context.Context, boardID, title string) (*Column, error) {
	c := &Column{BoardID: boardID, Title: title}
	err := s.db.QueryRow(ctx,
		`INSERT INTO columns (board_id, title, position)
		 VALUES ($1, $2, COALESCE((SELECT MAX(position)+1 FROM columns WHERE board_id=$1), 0))
		 RETURNING id, board_id, title, position, created_at`,
		boardID, title,
	).Scan(&c.ID, &c.BoardID, &c.Title, &c.Position, &c.CreatedAt)
	return c, err
}

func (s *Service) UpdateColumn(ctx context.Context, colID, title string) (*Column, error) {
	c := &Column{}
	err := s.db.QueryRow(ctx,
		`UPDATE columns SET title=$1 WHERE id=$2
		 RETURNING id, board_id, title, position, created_at`,
		title, colID,
	).Scan(&c.ID, &c.BoardID, &c.Title, &c.Position, &c.CreatedAt)
	return c, err
}

func (s *Service) DeleteColumn(ctx context.Context, colID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM columns WHERE id=$1`, colID)
	return err
}

// ── Cards ────────────────────────────────────────────────────────────────────

func (s *Service) getCards(ctx context.Context, colID string) ([]Card, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, column_id, title, description, position, created_at
		 FROM cards WHERE column_id = $1 ORDER BY position`, colID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := []Card{}
	for rows.Next() {
		var c Card
		if err := rows.Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &c.CreatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (s *Service) CreateCard(ctx context.Context, colID, title, desc string) (*Card, error) {
	c := &Card{ColumnID: colID, Title: title, Description: desc}
	err := s.db.QueryRow(ctx,
		`INSERT INTO cards (column_id, title, description, position)
		 VALUES ($1, $2, $3, COALESCE((SELECT MAX(position)+1 FROM cards WHERE column_id=$1), 0))
		 RETURNING id, column_id, title, description, position, created_at`,
		colID, title, desc,
	).Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &c.CreatedAt)
	return c, err
}

func (s *Service) UpdateCard(ctx context.Context, cardID, title, desc string) (*Card, error) {
	c := &Card{}
	err := s.db.QueryRow(ctx,
		`UPDATE cards SET title=$1, description=$2 WHERE id=$3
		 RETURNING id, column_id, title, description, position, created_at`,
		title, desc, cardID,
	).Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &c.CreatedAt)
	return c, err
}

// MoveCard moves a card to a new column and/or position, re-ordering siblings.
func (s *Service) MoveCard(ctx context.Context, cardID, newColID string, newPos int) (*Card, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Shift cards in destination column to make room
	if _, err := tx.Exec(ctx,
		`UPDATE cards SET position = position + 1
		 WHERE column_id = $1 AND position >= $2 AND id != $3`,
		newColID, newPos, cardID,
	); err != nil {
		return nil, err
	}

	var c Card
	if err := tx.QueryRow(ctx,
		`UPDATE cards SET column_id=$1, position=$2 WHERE id=$3
		 RETURNING id, column_id, title, description, position, created_at`,
		newColID, newPos, cardID,
	).Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &c.CreatedAt); err != nil {
		return nil, err
	}

	return &c, tx.Commit(ctx)
}

func (s *Service) DeleteCard(ctx context.Context, cardID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM cards WHERE id=$1`, cardID)
	return err
}
