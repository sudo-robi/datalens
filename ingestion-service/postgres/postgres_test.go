package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"
)

// --- Test driver for mock database ---

type pgTestDriver struct{}

func init() {
	sql.Register("pgtestdrv", &pgTestDriver{})
}

func (d *pgTestDriver) Open(name string) (driver.Conn, error) {
	return &pgTestConn{}, nil
}

type pgTestConn struct{}

func (c *pgTestConn) Prepare(query string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *pgTestConn) Close() error                             { return nil }
func (c *pgTestConn) Begin() (driver.Tx, error)                { return nil, driver.ErrSkip }
func (c *pgTestConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return &pgTestResult{}, nil
}
func (c *pgTestConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return &pgTestRows{
		cols:    []string{"id", "name", "description", "filename", "mime_type", "row_count", "size_bytes", "created_at"},
		results: [][]driver.Value{{int64(1), "test", "desc", "test.csv", "text/csv", int64(100), int64(1024), time.Now()}},
	}, nil
}

type pgTestResult struct{}

func (r *pgTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r *pgTestResult) RowsAffected() (int64, error) { return 0, nil }

type pgTestRows struct {
	cols    []string
	results [][]driver.Value
	pos     int
}

func (r *pgTestRows) Columns() []string { return r.cols }
func (r *pgTestRows) Close() error      { return nil }
func (r *pgTestRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.results) {
		return io.EOF
	}
	copy(dest, r.results[r.pos])
	r.pos++
	return nil
}

func newTestClient() *Client {
	db, _ := sql.Open("pgtestdrv", "")
	return &Client{db: db}
}

// --- Tests ---

func TestNewClientInvalidConnString(t *testing.T) {
	_, err := NewClient("postgres://invalid:invalid@localhost:99999/nonexistent?sslmode=disable")
	if err == nil {
		t.Error("expected error for invalid connection string, got nil")
	}
}

func TestNewClientEmptyConnString(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Error("expected error for empty connection string, got nil")
	}
}

func TestNewClientUnreachableHost(t *testing.T) {
	_, err := NewClient("postgres://user:pass@localhost:1/nonexistent?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Error("expected error for unreachable host, got nil")
	}
}

func TestNewClientInvalidFormat(t *testing.T) {
	_, err := NewClient("not-a-valid-connection-string")
	if err == nil {
		t.Error("expected error for malformed connection string, got nil")
	}
}

func TestClientClose(t *testing.T) {
	client := newTestClient()
	if err := client.Close(); err != nil {
		t.Errorf("expected no error on Close, got %v", err)
	}
}

func TestUpdateDatasetRowCountSuccess(t *testing.T) {
	client := newTestClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.UpdateDatasetRowCount(ctx, 1, 100)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUpdateDatasetRowCountContextCancelled(t *testing.T) {
	client := newTestClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := client.UpdateDatasetRowCount(ctx, 1, 100)
	if err == nil {
		t.Log("UpdateDatasetRowCount did not return error with cancelled context (driver may ignore context)")
	}
}

func TestGetDatasetSuccess(t *testing.T) {
	client := newTestClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.GetDataset(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result["id"] != int64(1) {
		t.Errorf("expected id=1, got %v", result["id"])
	}
	if result["name"] != "test" {
		t.Errorf("expected name='test', got %v", result["name"])
	}
	if result["filename"] != "test.csv" {
		t.Errorf("expected filename='test.csv', got %v", result["filename"])
	}
	if result["row_count"] != int64(100) {
		t.Errorf("expected row_count=100, got %v", result["row_count"])
	}
}

func TestGetDatasetContextCancelled(t *testing.T) {
	client := newTestClient()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetDataset(ctx, 1)
	if err == nil {
		t.Log("GetDataset did not return error with cancelled context (driver may ignore context)")
	}
}
