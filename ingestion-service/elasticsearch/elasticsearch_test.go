package elasticsearch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_cluster/health" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "green"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClientUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := NewClient(server.URL)
	if err == nil {
		t.Fatal("expected error for unhealthy server, got nil")
	}
	if !strings.Contains(err.Error(), "health check failed") {
		t.Errorf("expected 'health check failed' in error, got: %v", err)
	}
}

func TestNewClientUnreachable(t *testing.T) {
	_, err := NewClient("http://localhost:19999")
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestIndexExistsTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" && r.URL.Path == "/test_index" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	exists, err := client.IndexExists(context.Background(), "test_index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected index to exist")
	}
}

func TestIndexExistsFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	exists, err := client.IndexExists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected index to not exist")
	}
}

func TestCreateIndex(t *testing.T) {
	var receivedBody []byte
	var receivedMethod string
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	err := client.CreateIndex(context.Background(), "my_index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != "PUT" {
		t.Errorf("expected PUT, got %s", receivedMethod)
	}
	if receivedPath != "/my_index" {
		t.Errorf("expected /my_index, got %s", receivedPath)
	}

	// Verify the mapping body
	var mapping map[string]interface{}
	if err := json.Unmarshal(receivedBody, &mapping); err != nil {
		t.Fatalf("failed to unmarshal mapping: %v", err)
	}

	mappings, ok := mapping["mappings"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'mappings' key in body")
	}
	props, ok := mappings["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'properties' key in mappings")
	}
	if _, ok := props["dataset_id"]; !ok {
		t.Error("expected 'dataset_id' in properties")
	}
	if _, ok := props["ingested_at"]; !ok {
		t.Error("expected 'ingested_at' in properties")
	}
	if _, ok := props["index"]; !ok {
		t.Error("expected 'index' in properties")
	}
}

func TestCreateIndexConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict) // 409 - index already exists
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	err := client.CreateIndex(context.Background(), "existing_index")
	if err == nil {
		t.Fatal("expected error for 409 status, got nil")
	}
	if !strings.Contains(err.Error(), "status 409") {
		t.Errorf("expected 'status 409' in error, got: %v", err)
	}
}

func TestBulkIndex(t *testing.T) {
	var receivedBody []byte
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": false})
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}

	docs := []map[string]interface{}{
		{"dataset_id": 1, "name": "Alice"},
		{"dataset_id": 1, "name": "Bob"},
	}

	err := client.BulkIndex(context.Background(), "test_index", docs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedContentType != "application/x-ndjson" {
		t.Errorf("expected Content-Type 'application/x-ndjson', got %s", receivedContentType)
	}

	// Verify NDJSON format (each line is a JSON object)
	lines := strings.Split(strings.TrimSpace(string(receivedBody)), "\n")
	// 2 docs = 4 lines (2 action + 2 doc)
	if len(lines) != 4 {
		t.Errorf("expected 4 NDJSON lines, got %d: %s", len(lines), string(receivedBody))
	}

	// Verify first line is an index action
	var action map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &action); err != nil {
		t.Fatalf("failed to parse action line: %v", err)
	}
	indexAction, ok := action["index"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'index' key in action")
	}
	if indexAction["_index"] != "test_index" {
		t.Errorf("expected _index='test_index', got %v", indexAction["_index"])
	}
	if indexAction["_id"] != "1_0" {
		t.Errorf("expected _id='1_0', got %v", indexAction["_id"])
	}

	// Verify second line is a document
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &doc); err != nil {
		t.Fatalf("failed to parse doc line: %v", err)
	}
	if doc["name"] != "Alice" {
		t.Errorf("expected name='Alice', got %v", doc["name"])
	}
}

func TestBulkIndexErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": true})
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}

	docs := []map[string]interface{}{
		{"dataset_id": 1, "name": "test"},
	}

	err := client.BulkIndex(context.Background(), "test_index", docs)
	if err == nil {
		t.Fatal("expected error when bulk indexing has errors, got nil")
	}
	if !strings.Contains(err.Error(), "bulk indexing completed with errors") {
		t.Errorf("expected 'bulk indexing completed with errors', got: %v", err)
	}
}

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/_search") {
			t.Errorf("expected path ending with /_search, got %s", r.URL.Path)
		}

		// Read and verify the search body
		body, _ := io.ReadAll(r.Body)
		var searchReq map[string]interface{}
		json.Unmarshal(body, &searchReq)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": 2},
				"hits": []interface{}{
					map[string]interface{}{
						"_source": map[string]interface{}{
							"name": "Alice",
							"age":  30,
						},
					},
					map[string]interface{}{
						"_source": map[string]interface{}{
							"name": "Bob",
							"age":  25,
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	result, err := client.Search(context.Background(), "test_index", "Alice", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total=2, got %d", result.Total)
	}
	if len(result.Hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(result.Hits))
	}
	if result.Hits[0]["name"] != "Alice" {
		t.Errorf("expected first hit name='Alice', got %v", result.Hits[0]["name"])
	}
	if result.Page != 1 {
		t.Errorf("expected page=1, got %d", result.Page)
	}
	if result.PerPage != 20 {
		t.Errorf("expected perPage=20, got %d", result.PerPage)
	}
}

func TestSearchPagination(t *testing.T) {
	var fromValue, sizeValue float64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var searchReq map[string]interface{}
		json.Unmarshal(body, &searchReq)
		fromValue = searchReq["from"].(float64)
		sizeValue = searchReq["size"].(float64)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": 0},
				"hits":  []interface{}{},
			},
		})
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	_, err := client.Search(context.Background(), "idx", "q", 3, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// page=3, perPage=10 => from=(3-1)*10=20
	if fromValue != 20 {
		t.Errorf("expected from=20, got %v", fromValue)
	}
	if sizeValue != 10 {
		t.Errorf("expected size=10, got %v", sizeValue)
	}
}

func TestSearchEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": 0},
				"hits":  []interface{}{},
			},
		})
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	result, err := client.Search(context.Background(), "idx", "nothing", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected total=0, got %d", result.Total)
	}
	if len(result.Hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(result.Hits))
	}
}

func TestSearchInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, httpClient: &http.Client{}}
	_, err := client.Search(context.Background(), "idx", "q", 1, 20)
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
}
