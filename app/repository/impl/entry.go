package impl

import (
	"context"
	"github.com/belito3/go-web-api/app/model"
)


type Entry = model.Entry

type CreateEntryParams struct {
	AccountID	int64	`json:"account_id"`
	Amount		int64	`json:"amount"`
}

type ListEntriesParams struct {
	AccountID 	int64	`json:"account_id"`
	Limit		int32	`json:"limit"`
	Offset		int32	`json:"offset"`
}

type IEntry interface {
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	GetEntry(ctx context.Context, id int64) (Entry, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
}


func(a *Queries) CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error) {
	query := `INSERT INTO entries (account_id, amount) VALUES ($1, $2) RETURNING *`
	var i Entry

	row := a.db.QueryRowContext(ctx, query, arg.AccountID, arg.Amount)
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.CreatedAt,
	)

	return i, err
}

func(a *Queries) GetEntry(ctx context.Context, id int64) (Entry, error) {
	query := `SELECT id, account_id, amount, created_at FROM entries WHERE id = $1 LIMIT 1`
	var i Entry

	row := a.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&i.ID,
		&i.AccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

func(a *Queries) ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error) {
	query := `SELECT id, account_id, amount, created_at FROM entries WHERE account_id = $1 ORDER BY id LIMIT $2 OFFSET $3`

	rows, err := a.db.QueryContext(ctx, query, arg.AccountID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Entry
	for rows.Next() {
		var i Entry
		if err := rows.Scan(
			&i.ID,
			&i.AccountID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}