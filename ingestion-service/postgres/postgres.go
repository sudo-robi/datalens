package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Client struct {
	db *sql.DB
}

func NewClient(connString string) (*Client, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) UpdateDatasetRowCount(ctx context.Context, datasetID int64, rowCount int64) error {
	query := `UPDATE datasets SET row_count = $1, updated_at = NOW() WHERE id = $2`
	_, err := c.db.ExecContext(ctx, query, rowCount, datasetID)
	return err
}

func (c *Client) GetDataset(ctx context.Context, datasetID int64) (map[string]interface{}, error) {
	query := `SELECT id, name, description, filename, mime_type, row_count, size_bytes, created_at 
	           FROM datasets WHERE id = $1`
	row := c.db.QueryRowContext(ctx, query, datasetID)

	var result map[string]interface{}
	var id int64
	var name, filename, mimeType sql.NullString
	var description sql.NullString
	var rowCount, sizeBytes sql.NullInt64
	var createdAt time.Time

	err := row.Scan(&id, &name, &description, &filename, &mimeType, &rowCount, &sizeBytes, &createdAt)
	if err != nil {
		return nil, err
	}

	result = map[string]interface{}{
		"id":         id,
		"name":       name.String,
		"description": description.String,
		"filename":   filename.String,
		"mime_type":  mimeType.String,
		"row_count":  rowCount.Int64,
		"size_bytes": sizeBytes.Int64,
		"created_at": createdAt,
	}

	return result, nil
}
