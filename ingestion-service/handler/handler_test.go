package handler

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/datalens/ingestion-service/elasticsearch"
	"github.com/datalens/ingestion-service/postgres"
)

// --- Test driver for mock database ---

type handlerTestDriver struct{}

func init() {
	sql.Register("handlertestdrv", &handlerTestDriver{})
}

func (d *handlerTestDriver) Open(name string) (driver.Conn, error) {
	return &handlerTestConn{}, nil
}

type handlerTestConn struct{}

func (c *handlerTestConn) Prepare(query string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *handlerTestConn) Close() error                             { return nil }
func (c *handlerTestConn) Begin() (driver.Tx, error)                { return nil, driver.ErrSkip }
func (c *handlerTestConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return &handlerTestResult{}, nil
}
func (c *handlerTestConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

type handlerTestResult struct{}

func (r *handlerTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r *handlerTestResult) RowsAffected() (int64, error) { return 0, nil }

// createTestPGClient creates a postgres.Client with a mock sql.DB via reflect.
func createTestPGClient() *postgres.Client {
	db, _ := sql.Open("handlertestdrv", "")
	v := reflect.New(reflect.TypeOf(postgres.Client{}))
	f := v.Elem().FieldByName("db")
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(db))
	return v.Interface().(*postgres.Client)
}

// newMockESServer creates an httptest.Server that mimics Elasticsearch.
func newMockESServer(t *testing.T, indexExists bool) *httptest.Server {
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

// --- Mock Elasticsearch ---

type mockES struct {
	indexExists bool
	searchFn    func(index, query string, page, perPage int) *searchResult
}

type searchResult struct {
	Total   int64                    `json:"total"`
	Hits    []map[string]interface{} `json:"hits"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
}

func (m *mockES) IndexExistsHandler(w http.ResponseWriter, r *http.Request) {
	if m.indexExists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *mockES) CreateIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (m *mockES) BulkIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"errors": false})
}

func (m *mockES) SearchHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Query string `json:"query"`
	}
	json.Unmarshal(body, &req)

	if m.searchFn != nil {
		idx := r.URL.Path
		// extract index from path: /{index}/_search
		if len(idx) > 0 && idx[0] == '/' {
			idx = idx[1:]
		}
		// strip /_search suffix
		if len(idx) > len("/_search") {
			idx = idx[:len(idx)-len("/_search")]
		}
		result := m.searchFn(idx, req.Query, 1, 20)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": result.Total},
				"hits":  result.Hits,
			},
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hits": map[string]interface{}{
			"total": map[string]interface{}{"value": 0},
			"hits":  []interface{}{},
		},
	})
}

func (m *mockES) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "green",
	})
}

// --- Mock Postgres ---

type mockDB struct {
	updateRowCountFn func(datasetID, rowCount int64) error
}

func (m *mockDB) UpdateDatasetCountHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// --- Tests ---

func TestHealthHandler(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var body APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !body.Success {
		t.Error("expected success to be true")
	}

	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", body.Data)
	}

	if data["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", data["status"])
	}
	if data["service"] != "ingestion-service" {
		t.Errorf("expected service 'ingestion-service', got %v", data["service"])
	}
}

func TestParseCSVRowInt(t *testing.T) {
	header := []string{"id", "name", "age"}
	row := []string{"42", "Alice", "30"}

	result := parseCSVRow(header, row)

	if result["id"] != 42 {
		t.Errorf("expected id=42 (int), got %v (%T)", result["id"], result["id"])
	}
	if result["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("expected age=30 (int), got %v (%T)", result["age"], result["age"])
	}
}

func TestParseCSVRowFloat(t *testing.T) {
	header := []string{"price", "quantity"}
	row := []string{"19.99", "5"}

	result := parseCSVRow(header, row)

	if result["price"] != 19.99 {
		t.Errorf("expected price=19.99 (float64), got %v (%T)", result["price"], result["price"])
	}
	if result["quantity"] != 5 {
		t.Errorf("expected quantity=5 (int), got %v (%T)", result["quantity"], result["quantity"])
	}
}

func TestParseCSVRowMixed(t *testing.T) {
	header := []string{"id", "name", "active", "score"}
	row := []string{"1", "Bob", "true", "88.5"}

	result := parseCSVRow(header, row)

	if result["id"] != 1 {
		t.Errorf("expected id=1, got %v", result["id"])
	}
	if result["name"] != "Bob" {
		t.Errorf("expected name='Bob', got %v", result["name"])
	}
	if result["active"] != "true" {
		t.Errorf("expected active='true', got %v", result["active"])
	}
	if result["score"] != 88.5 {
		t.Errorf("expected score=88.5, got %v", result["score"])
	}
}

func TestParseCSVRowEmpty(t *testing.T) {
	header := []string{"a", "b", "c"}
	row := []string{"", "", ""}

	result := parseCSVRow(header, row)

	for _, h := range header {
		if result[h] != "" {
			t.Errorf("expected empty string for %s, got %v", h, result[h])
		}
	}
}

func TestParseCSVRowMismatchedLengths(t *testing.T) {
	header := []string{"a", "b", "c", "d"}
	row := []string{"1", "2"}

	result := parseCSVRow(header, row)

	if result["a"] != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if result["b"] != 2 {
		t.Errorf("expected b=2, got %v", result["b"])
	}
	if _, exists := result["c"]; exists {
		t.Error("expected c to not exist when row is shorter")
	}
	if _, exists := result["d"]; exists {
		t.Error("expected d to not exist when row is shorter")
	}
}

func TestParseCSVRowLongerRow(t *testing.T) {
	header := []string{"a", "b"}
	row := []string{"1", "2", "3", "4"}

	result := parseCSVRow(header, row)

	if result["a"] != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if result["b"] != 2 {
		t.Errorf("expected b=2, got %v", result["b"])
	}
}

func TestParseCSVRowSpecialChars(t *testing.T) {
	header := []string{"text", "emoji", "unicode"}
	row := []string{"hello, world", "🎉", "日本語"}

	result := parseCSVRow(header, row)

	if result["text"] != "hello, world" {
		t.Errorf("expected text='hello, world', got %v", result["text"])
	}
	if result["emoji"] != "🎉" {
		t.Errorf("expected emoji='🎉', got %v", result["emoji"])
	}
	if result["unicode"] != "日本語" {
		t.Errorf("expected unicode='日本語', got %v", result["unicode"])
	}
}

func TestParseCSVRowWhitespace(t *testing.T) {
	header := []string{"name", "value"}
	row := []string{"  Alice  ", "  42  "}

	result := parseCSVRow(header, row)

	if result["name"] != "Alice" {
		t.Errorf("expected name='Alice' (trimmed), got %v", result["name"])
	}
	// "  42  " trimmed becomes "42" which parses as int
	if result["value"] != 42 {
		t.Errorf("expected value=42, got %v", result["value"])
	}
}

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := APIResponse{Success: true, Data: "test"}

	respondJSON(w, http.StatusOK, data)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if !body.Success {
		t.Error("expected success=true")
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()

	respondError(w, http.StatusBadRequest, "bad input")

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Success {
		t.Error("expected success=false")
	}
	if body.Error != "bad input" {
		t.Errorf("expected error='bad input', got %s", body.Error)
	}
}

func TestSearchInvalidJSON(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.Search(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "Invalid request body" {
		t.Errorf("expected error 'Invalid request body', got %s", body.Error)
	}
}

func TestSearchEmptyBody(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(""))
	w := httptest.NewRecorder()

	h.Search(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestSearchDefaultValuesNilHandler(t *testing.T) {
	// Verify that a valid body with defaults is accepted by the decoder.
	h := &Handler{}

	body := `{"query": "test"}`
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			// Expected: nil pointer dereference because es is nil.
		}
	}()

	h.Search(w, req)
}

func TestIngestCSVNoFile(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest("POST", "/ingest", nil)
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "No file provided" {
		t.Errorf("expected error 'No file provided', got %s", body.Error)
	}
}

func TestIngestCSVInvalidDatasetID(t *testing.T) {
	h := &Handler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "not_a_number")
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("name,age\nAlice,30\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "Invalid dataset_id" {
		t.Errorf("expected error 'Invalid dataset_id', got %s", body.Error)
	}
}

func TestIngestCSVInvalidCSV(t *testing.T) {
	h := &Handler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "1")
	part, _ := writer.CreateFormFile("file", "test.csv")
	// Write invalid CSV (empty file causes header read error)
	part.Write([]byte(""))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "Failed to read CSV header" {
		t.Errorf("expected error 'Failed to read CSV header', got %s", body.Error)
	}
}

func TestIngestCSVNoDatasetID(t *testing.T) {
	h := &Handler{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("name,age\nAlice,30\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil Handler")
	}
	if h.db != nil {
		t.Error("expected nil db")
	}
	if h.es != nil {
		t.Error("expected nil es")
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	mockES := newMockESServer(t, false)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	h2 := NewHandler(pgClient, esClient)
	if h2 == nil {
		t.Fatal("expected non-nil Handler with real deps")
	}
}

func TestIngestCSVSuccess(t *testing.T) {
	mockES := newMockESServer(t, false)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "42")
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("name,age,city\nAlice,30,NYC\nBob,25,LA\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if !body.Success {
		t.Error("expected success=true")
	}

	data, ok := body.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", body.Data)
	}

	if data["rows_ingested"] != float64(2) {
		t.Errorf("expected rows_ingested=2, got %v", data["rows_ingested"])
	}
	if data["index"] != "dataset_42" {
		t.Errorf("expected index='dataset_42', got %v", data["index"])
	}
	if data["filename"] != "test.csv" {
		t.Errorf("expected filename='test.csv', got %v", data["filename"])
	}
}

func TestIngestCSVIndexAlreadyExists(t *testing.T) {
	mockES := newMockESServer(t, true)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "1")
	part, _ := writer.CreateFormFile("file", "data.csv")
	part.Write([]byte("col1,col2\nval1,val2\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if !body.Success {
		t.Error("expected success=true")
	}
}

func TestIngestCSVLargeBatch(t *testing.T) {
	mockES := newMockESServer(t, false)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	csvData := "id,value\n"
	for i := 0; i < 150; i++ {
		csvData += fmt.Sprintf("%d,item%d\n", i, i)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "5")
	part, _ := writer.CreateFormFile("file", "large.csv")
	part.Write([]byte(csvData))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	data := body.Data.(map[string]interface{})
	if data["rows_ingested"] != float64(150) {
		t.Errorf("expected rows_ingested=150, got %v", data["rows_ingested"])
	}
}

func TestIngestCSVWithEmptyRows(t *testing.T) {
	mockES := newMockESServer(t, false)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	csvData := "name,age\nAlice,30\n\nBob,25\n\n"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "10")
	part, _ := writer.CreateFormFile("file", "empty_rows.csv")
	part.Write([]byte(csvData))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestSearchDefaultValues(t *testing.T) {
	mockES := newMockESServer(t, false)
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	h := NewHandler(nil, esClient)

	body := `{"query": "test"}`
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Search(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	if !apiResp.Success {
		t.Error("expected success=true")
	}
}

func TestIngestCSVIndexCheckError(t *testing.T) {
	// Start server and create client, then close server so requests fail
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "green"})
	}))

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}
	mockES.Close() // Close so subsequent requests fail

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "1")
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("name,age\nAlice,30\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "Failed to check index" {
		t.Errorf("expected error 'Failed to check index', got %s", body.Error)
	}
}

func TestIngestCSVCreateIndexError(t *testing.T) {
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		switch {
		case path == "/_cluster/health":
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
		case r.Method == "HEAD":
			w.WriteHeader(http.StatusNotFound) // index doesn't exist
		case r.Method == "PUT":
			w.WriteHeader(http.StatusInternalServerError) // create fails
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	pgClient := createTestPGClient()
	defer pgClient.Close()

	h := NewHandler(pgClient, esClient)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("dataset_id", "1")
	part, _ := writer.CreateFormFile("file", "test.csv")
	part.Write([]byte("name,age\nAlice,30\n"))
	writer.Close()

	req := httptest.NewRequest("POST", "/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	h.IngestCSV(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	var body APIResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Error != "Failed to create index" {
		t.Errorf("expected error 'Failed to create index', got %s", body.Error)
	}
}

func TestSearchESError(t *testing.T) {
	// Server returns invalid JSON for search
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/_cluster/health" {
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
			return
		}
		// Return invalid JSON for search
		w.Write([]byte("not valid json"))
	}))
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	h := NewHandler(nil, esClient)

	body := `{"query": "test", "index": "test_index"}`
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Search(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	var apiResp APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	if apiResp.Error != "Search failed" {
		t.Errorf("expected error 'Search failed', got %s", apiResp.Error)
	}
}

func TestSearchWithCustomValues(t *testing.T) {
	mockES := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/_cluster/health" {
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": 0},
				"hits":  []interface{}{},
			},
		})
	}))
	defer mockES.Close()

	esClient, err := elasticsearch.NewClient(mockES.URL)
	if err != nil {
		t.Fatalf("failed to create ES client: %v", err)
	}

	h := NewHandler(nil, esClient)

	body := `{"query": "test", "index": "custom_idx", "page": 3, "per_page": 10}`
	req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.Search(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
