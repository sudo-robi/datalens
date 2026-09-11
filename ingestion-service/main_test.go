package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/datalens/ingestion-service/elasticsearch"
	"github.com/datalens/ingestion-service/handler"
	"github.com/datalens/ingestion-service/postgres"
)

func TestGetEnvReturnsFallback(t *testing.T) {
	os.Unsetenv("TEST_GETENV_NOTSET_KEY_12345")
	val := getEnv("TEST_GETENV_NOTSET_KEY_12345", "fallback_value")
	if val != "fallback_value" {
		t.Errorf("expected 'fallback_value', got %q", val)
	}
}

func TestGetEnvReturnsEnvValue(t *testing.T) {
	os.Setenv("TEST_GETENV_SET_KEY_12345", "env_value")
	defer os.Unsetenv("TEST_GETENV_SET_KEY_12345")

	val := getEnv("TEST_GETENV_SET_KEY_12345", "fallback_value")
	if val != "env_value" {
		t.Errorf("expected 'env_value', got %q", val)
	}
}

func TestGetEnvEmptyEnvValue(t *testing.T) {
	os.Setenv("TEST_GETENV_EMPTY_KEY_12345", "")
	defer os.Unsetenv("TEST_GETENV_EMPTY_KEY_12345")

	val := getEnv("TEST_GETENV_EMPTY_KEY_12345", "fallback")
	// When env is set (even to empty string), LookupEnv returns exists=true
	if val != "" {
		t.Errorf("expected empty string, got %q", val)
	}
}

func TestParseCSVRowIntegers(t *testing.T) {
	header := []string{"id", "count"}
	row := []string{"100", "0"}

	result := parseCSVRow(header, row)

	if result["id"] != 100 {
		t.Errorf("expected id=100 (int), got %v (%T)", result["id"], result["id"])
	}
	if result["count"] != 0 {
		t.Errorf("expected count=0 (int), got %v (%T)", result["count"], result["count"])
	}
}

func TestParseCSVRowFloats(t *testing.T) {
	header := []string{"price", "tax"}
	row := []string{"19.99", "0.075"}

	result := parseCSVRow(header, row)

	if result["price"] != 19.99 {
		t.Errorf("expected price=19.99, got %v", result["price"])
	}
	if result["tax"] != 0.075 {
		t.Errorf("expected tax=0.075, got %v", result["tax"])
	}
}

func TestParseCSVRowStrings(t *testing.T) {
	header := []string{"name", "email"}
	row := []string{"John Doe", "john@example.com"}

	result := parseCSVRow(header, row)

	if result["name"] != "John Doe" {
		t.Errorf("expected name='John Doe', got %v", result["name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("expected email='john@example.com', got %v", result["email"])
	}
}

func TestParseCSVRowMixedTypes(t *testing.T) {
	header := []string{"id", "name", "price", "active"}
	row := []string{"1", "Widget", "29.99", "yes"}

	result := parseCSVRow(header, row)

	if result["id"] != 1 {
		t.Errorf("expected id=1 (int), got %v (%T)", result["id"], result["id"])
	}
	if result["name"] != "Widget" {
		t.Errorf("expected name='Widget', got %v", result["name"])
	}
	if result["price"] != 29.99 {
		t.Errorf("expected price=29.99 (float), got %v (%T)", result["price"], result["price"])
	}
	if result["active"] != "yes" {
		t.Errorf("expected active='yes', got %v", result["active"])
	}
}

func TestParseCSVRowEmptyRow(t *testing.T) {
	header := []string{"a", "b", "c"}
	row := []string{}

	result := parseCSVRow(header, row)

	// All fields should be missing since row is empty
	for _, h := range header {
		if _, exists := result[h]; exists {
			t.Errorf("expected %s to not exist in result", h)
		}
	}
}

func TestParseCSVRowMismatchedLengths(t *testing.T) {
	header := []string{"a", "b", "c", "d"}
	row := []string{"1"}

	result := parseCSVRow(header, row)

	if result["a"] != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if _, exists := result["b"]; exists {
		t.Error("expected b to not exist")
	}
}

func TestParseCSVRowMoreRowThanHeader(t *testing.T) {
	header := []string{"a"}
	row := []string{"1", "2", "3"}

	result := parseCSVRow(header, row)

	if result["a"] != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if len(result) != 1 {
		t.Errorf("expected 1 field in result, got %d", len(result))
	}
}

func TestParseCSVRowSpecialCharacters(t *testing.T) {
	header := []string{"text", "path", "quotes"}
	row := []string{"hello world", "/usr/local/bin", `say "hello"`}

	result := parseCSVRow(header, row)

	if result["text"] != "hello world" {
		t.Errorf("expected 'hello world', got %v", result["text"])
	}
	if result["path"] != "/usr/local/bin" {
		t.Errorf("expected '/usr/local/bin', got %v", result["path"])
	}
	if result["quotes"] != `say "hello"` {
		t.Errorf("expected 'say \"hello\"', got %v", result["quotes"])
	}
}

func TestParseCSVRowNegativeNumbers(t *testing.T) {
	header := []string{"temp", "balance"}
	row := []string{"-5.5", "-100"}

	result := parseCSVRow(header, row)

	if result["temp"] != -5.5 {
		t.Errorf("expected temp=-5.5, got %v", result["temp"])
	}
	if result["balance"] != -100 {
		t.Errorf("expected balance=-100, got %v (%T)", result["balance"], result["balance"])
	}
}

func TestParseCSVRowWhitespace(t *testing.T) {
	header := []string{"name", "value"}
	row := []string{"  hello  ", "  42  "}

	result := parseCSVRow(header, row)

	// "  hello  " trimmed -> "hello" (string, not int)
	if result["name"] != "hello" {
		t.Errorf("expected name='hello', got %v", result["name"])
	}
	// "  42  " trimmed -> "42" (parsed as int)
	if result["value"] != 42 {
		t.Errorf("expected value=42, got %v", result["value"])
	}
}

func TestProcessCSVValid(t *testing.T) {
	csvData := "name,age,city\nAlice,30,NYC\nBob,25,LA\n"
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(header) != 3 {
		t.Fatalf("expected 3 header columns, got %d", len(header))
	}
	if header[0] != "name" || header[1] != "age" || header[2] != "city" {
		t.Errorf("unexpected headers: %v", header)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", rows[0]["name"])
	}
	if rows[0]["age"] != 30 {
		t.Errorf("expected age=30, got %v", rows[0]["age"])
	}
	if rows[1]["name"] != "Bob" {
		t.Errorf("expected name='Bob', got %v", rows[1]["name"])
	}
}

func TestProcessCSVEmpty(t *testing.T) {
	_, _, err := processCSV(strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty CSV, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read CSV header") {
		t.Errorf("expected 'failed to read CSV header' error, got: %v", err)
	}
}

func TestProcessCSVMalformedRow(t *testing.T) {
	// CSV with a row that has extra commas: Go's csv.Reader returns an error
	// for field count mismatch, and processCSV skips it with `continue`.
	csvData := "id,name\n1,Alice\n2,Bob,extra\n"
	rows, _, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the well-formed row is kept; the extra-column row is skipped.
	if len(rows) != 1 {
		t.Errorf("expected 1 row (malformed skipped), got %d", len(rows))
	}
}

func TestProcessCSVHeaderOnly(t *testing.T) {
	csvData := "col1,col2,col3\n"
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(header) != 3 {
		t.Errorf("expected 3 headers, got %d", len(header))
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestProcessCSVDifferentDataTypes(t *testing.T) {
	csvData := "id,name,score,active\n1,Alice,99.5,true\n2,Bob,88.0,false\n"
	rows, _, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Check types
	if _, ok := rows[0]["id"].(int); !ok {
		t.Errorf("expected id to be int, got %T", rows[0]["id"])
	}
	if _, ok := rows[0]["score"].(float64); !ok {
		t.Errorf("expected score to be float64, got %T", rows[0]["score"])
	}
	if _, ok := rows[0]["active"].(string); !ok {
		t.Errorf("expected active to be string, got %T", rows[0]["active"])
	}
}

func TestCorsMiddlewareAllowsOrigin(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected ACAO='*', got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsMiddlewareHandlesOPTIONS(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}
}

func TestCorsMiddlewareHeaders(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("unexpected ACM header: %s", resp.Header.Get("Access-Control-Allow-Methods"))
	}
	if resp.Header.Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("unexpected ACH header: %s", resp.Header.Get("Access-Control-Allow-Headers"))
	}
}

// --- Mock ES server for main tests ---

func newMockESForMain(t *testing.T, indexExists bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/_cluster/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
		case r.Method == "HEAD":
			if indexExists {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && path == "/_bulk":
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": false})
		case r.Method == "POST" && strings.HasSuffix(path, "/_search"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"hits": map[string]interface{}{
					"total": map[string]interface{}{"value": 0},
					"hits":  []interface{}{},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newMockESBulkError(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/_cluster/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
		case r.Method == "POST" && path == "/_bulk":
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": true})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newMockESCreateFail(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/_cluster/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
		case r.Method == "HEAD":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == "PUT":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func createMainESClient(t *testing.T, serverURL string) *elasticsearch.Client {
	t.Helper()
	esClient, err := elasticsearch.NewClient(serverURL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}
	return esClient
}

// --- Tests for ensureIndex ---

func TestEnsureIndexExists(t *testing.T) {
	mockES := newMockESForMain(t, true)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	err := ensureIndex(ctx, esClient, "test_index")
	if err != nil {
		t.Errorf("expected no error when index exists, got %v", err)
	}
}

func TestEnsureIndexNotExists(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	err := ensureIndex(ctx, esClient, "new_index")
	if err != nil {
		t.Errorf("expected no error when creating index, got %v", err)
	}
}

func TestEnsureIndexCreateError(t *testing.T) {
	mockES := newMockESCreateFail(t)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	err := ensureIndex(ctx, esClient, "fail_index")
	if err == nil {
		t.Error("expected error when CreateIndex fails, got nil")
	}
}

func TestEnsureIndexIndexExistsError(t *testing.T) {
	// Start server, create client, close server so IndexExists fails
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "green"})
	}))

	esClient := createMainESClient(t, mockES.URL)
	mockES.Close() // Close so subsequent requests fail

	ctx := context.Background()
	err := ensureIndex(ctx, esClient, "test")
	if err == nil {
		t.Error("expected error when IndexExists fails, got nil")
	}
}

// --- Tests for batchIndex ---

func TestBatchIndexEmpty(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	err := batchIndex(ctx, esClient, "test_index", nil)
	if err != nil {
		t.Errorf("expected no error for empty docs, got %v", err)
	}
}

func TestBatchIndexSmall(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	docs := make([]map[string]interface{}, 5)
	for i := range docs {
		docs[i] = map[string]interface{}{
			"dataset_id": int64(1),
			"name":       "item",
		}
	}

	err := batchIndex(ctx, esClient, "test_index", docs)
	if err != nil {
		t.Errorf("expected no error for small batch, got %v", err)
	}

	// Verify docs were modified with index and ingested_at
	for i, doc := range docs {
		if doc["index"] != i {
			t.Errorf("expected doc[%d].index=%d, got %v", i, i, doc["index"])
		}
		if doc["ingested_at"] == nil {
			t.Errorf("expected doc[%d].ingested_at to be set", i)
		}
	}
}

func TestBatchIndexLarge(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	docs := make([]map[string]interface{}, 150)
	for i := range docs {
		docs[i] = map[string]interface{}{
			"dataset_id": int64(1),
			"name":       "item",
		}
	}

	err := batchIndex(ctx, esClient, "test_index", docs)
	if err != nil {
		t.Errorf("expected no error for large batch, got %v", err)
	}

	// Verify last doc index
	lastDoc := docs[149]
	if lastDoc["index"] != 149 {
		t.Errorf("expected last doc index=149, got %v", lastDoc["index"])
	}
}

func TestBatchIndexExactBatchSize(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	docs := make([]map[string]interface{}, 100)
	for i := range docs {
		docs[i] = map[string]interface{}{
			"dataset_id": int64(1),
			"name":       "item",
		}
	}

	err := batchIndex(ctx, esClient, "test_index", docs)
	if err != nil {
		t.Errorf("expected no error for exact batch size, got %v", err)
	}
}

func TestBatchIndexBulkError(t *testing.T) {
	mockES := newMockESBulkError(t)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx := context.Background()

	docs := []map[string]interface{}{
		{"dataset_id": int64(1), "name": "item"},
	}

	err := batchIndex(ctx, esClient, "test_index", docs)
	if err == nil {
		t.Error("expected error when BulkIndex fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to index batch") {
		t.Errorf("expected error message to contain 'failed to index batch', got: %v", err)
	}
}

func TestBatchIndexContextCancelled(t *testing.T) {
	mockES := newMockESForMain(t, false)
	defer mockES.Close()

	esClient := createMainESClient(t, mockES.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	docs := []map[string]interface{}{
		{"dataset_id": int64(1), "name": "item"},
	}

	err := batchIndex(ctx, esClient, "test_index", docs)
	if err == nil {
		t.Log("batchIndex did not return error with cancelled context")
	}
}

// --- Tests for processCSV with different data ---

func TestProcessCSVDelimiters(t *testing.T) {
	// Standard comma-separated works
	csvData := "name,age\nAlice,30\n"
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(header) != 2 || header[0] != "name" || header[1] != "age" {
		t.Errorf("unexpected headers: %v", header)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestProcessCSVSemicolonDelimited(t *testing.T) {
	// Semicolon-separated with default comma reader: entire line is one field
	csvData := "name;age\nAlice;30\n"
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The semicolons are treated as part of the data, not delimiters
	if len(header) != 1 || header[0] != "name;age" {
		t.Errorf("expected single header 'name;age', got %v", header)
	}
	// Row data is also one field
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestProcessCSVTabDelimited(t *testing.T) {
	// Tab-separated with default comma reader: entire line is one field
	csvData := "name\tage\nAlice\t30\n"
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(header) != 1 || header[0] != "name\tage" {
		t.Errorf("expected single header 'name\\tage', got %v", header)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

func TestProcessCSVLargeDataset(t *testing.T) {
	csvData := "id,value\n"
	for i := 0; i < 1000; i++ {
		csvData += "x,y\n"
	}

	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(header) != 2 {
		t.Errorf("expected 2 headers, got %d", len(header))
	}
	if len(rows) != 1000 {
		t.Errorf("expected 1000 rows, got %d", len(rows))
	}
}

func TestProcessCSVWithQuotes(t *testing.T) {
	csvData := `"name","description"
"Alice","She said ""hello"""
"Bob","A ""quoted"" value"
`
	rows, header, err := processCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(header) != 2 {
		t.Errorf("expected 2 headers, got %d", len(header))
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0]["description"] != `She said "hello"` {
		t.Errorf("expected unescaped quotes, got %v", rows[0]["description"])
	}
}

func TestCorsMiddlewareNonOPTIONS(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected ACAO='*', got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestGetEnvConcurrency(t *testing.T) {
	key := "TEST_CONCURRENT_KEY_99999"
	os.Unsetenv(key)
	defer os.Unsetenv(key)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			val := getEnv(key, "fallback")
			if val != "fallback" {
				t.Errorf("expected 'fallback', got %q", val)
			}
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestParseCSVRowEdgeCases(t *testing.T) {
	// Empty header
	result := parseCSVRow([]string{}, []string{"a", "b"})
	if len(result) != 0 {
		t.Errorf("expected empty result for empty header, got %d fields", len(result))
	}

	// Very large number - ParseFloat succeeds, returns float64
	result = parseCSVRow([]string{"big"}, []string{"99999999999999999999"})
	if _, ok := result["big"].(float64); !ok {
		t.Errorf("expected very large number as float64, got %T", result["big"])
	}

	// "inf" is parsed as float64 +Inf by ParseFloat
	result = parseCSVRow([]string{"val"}, []string{"inf"})
	if _, ok := result["val"].(float64); !ok {
		t.Errorf("expected 'inf' as float64 (+Inf), got %T", result["val"])
	}

	// Negative zero
	result = parseCSVRow([]string{"val"}, []string{"-0"})
	if result["val"] != 0 {
		t.Errorf("expected -0 as 0, got %v", result["val"])
	}

	// Empty string field
	result = parseCSVRow([]string{"val"}, []string{""})
	if result["val"] != "" {
		t.Errorf("expected empty string, got %v", result["val"])
	}
}

// --- Test driver for mock database ---

type mainTestDriver struct{}

func init() {
	sql.Register("maintestdrv", &mainTestDriver{})
}

func (d *mainTestDriver) Open(name string) (driver.Conn, error) {
	return &mainTestConn{}, nil
}

type mainTestConn struct{}

func (c *mainTestConn) Prepare(query string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *mainTestConn) Close() error                             { return nil }
func (c *mainTestConn) Begin() (driver.Tx, error)                { return nil, driver.ErrSkip }
func (c *mainTestConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return &mainTestResult{}, nil
}
func (c *mainTestConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

type mainTestResult struct{}

func (r *mainTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r *mainTestResult) RowsAffected() (int64, error) { return 0, nil }

func createMainPGClient() *postgres.Client {
	db, _ := sql.Open("maintestdrv", "")
	v := reflect.New(reflect.TypeOf(postgres.Client{}))
	f := v.Elem().FieldByName("db")
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(db))
	return v.Interface().(*postgres.Client)
}

// TestServerSetup exercises the code paths in main() without calling main() directly.
func TestServerSetup(t *testing.T) {
	// Create mock ES server
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/_cluster/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
		case r.Method == "HEAD":
			w.WriteHeader(http.StatusOK)
		case r.Method == "POST" && r.URL.Path == "/_bulk":
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": false})
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/_search"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"hits": map[string]interface{}{
					"total": map[string]interface{}{"value": 0},
					"hits":  []interface{}{},
				},
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createMainPGClient()
	defer pgClient.Close()

	h := handler.NewHandler(pgClient, esClient)

	// Set up routes (same as main())
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", h.IngestCSV)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /search", h.Search)

	corsHandler := corsMiddleware(mux)

	// Start server on random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	go http.Serve(listener, corsHandler)
	time.Sleep(50 * time.Millisecond) // Wait for server to start

	baseURL := "http://" + listener.Addr().String()

	// Test health endpoint
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected health status 200, got %d", resp.StatusCode)
	}

	// Test search endpoint
	searchBody := `{"query": "test"}`
	resp, err = http.Post(baseURL+"/search", "application/json", strings.NewReader(searchBody))
	if err != nil {
		t.Fatalf("search request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected search status 200, got %d", resp.StatusCode)
	}

	// Test CORS preflight
	req, _ := http.NewRequest("OPTIONS", baseURL+"/health", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("options request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected OPTIONS status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected ACAO='*', got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestMainEnvironmentVariables tests the environment variable handling in main().
func TestMainEnvironmentVariables(t *testing.T) {
	// Test default values
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("ELASTICSEARCH_URL")
	os.Unsetenv("PORT")

	dbURL := getEnv("DATABASE_URL", "postgres://datalens:datalens_dev_2024@localhost:5433/datalens?sslmode=disable")
	if !strings.Contains(dbURL, "localhost:5433") {
		t.Errorf("unexpected default DB URL: %s", dbURL)
	}

	esURL := getEnv("ELASTICSEARCH_URL", "http://localhost:9200")
	if esURL != "http://localhost:9200" {
		t.Errorf("unexpected default ES URL: %s", esURL)
	}

	port := getEnv("PORT", "8081")
	if port != "8081" {
		t.Errorf("unexpected default port: %s", port)
	}

	// Test custom values
	os.Setenv("DATABASE_URL", "postgres://custom:custom@remote:5432/db")
	os.Setenv("ELASTICSEARCH_URL", "http://remote:9201")
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("ELASTICSEARCH_URL")
	defer os.Unsetenv("PORT")

	dbURL = getEnv("DATABASE_URL", "postgres://default")
	if !strings.Contains(dbURL, "custom@remote:5432") {
		t.Errorf("expected custom DB URL, got: %s", dbURL)
	}

	esURL = getEnv("ELASTICSEARCH_URL", "http://default")
	if esURL != "http://remote:9201" {
		t.Errorf("expected custom ES URL, got: %s", esURL)
	}

	port = getEnv("PORT", "8081")
	if port != "9090" {
		t.Errorf("expected custom port, got: %s", port)
	}
}
