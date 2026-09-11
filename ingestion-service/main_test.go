package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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
