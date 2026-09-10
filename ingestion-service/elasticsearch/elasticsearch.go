package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/_cluster/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Elasticsearch health check failed with status: %d", resp.StatusCode)
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: client,
	}, nil
}

func (c *Client) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", c.baseURL+"/"+indexName, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) CreateIndex(ctx context.Context, indexName string) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"dataset_id":  map[string]string{"type": "long"},
				"ingested_at": map[string]string{"type": "date"},
				"index":       map[string]string{"type": "integer"},
			},
		},
	}

	body, _ := json.Marshal(mapping)
	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/"+indexName, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create index: status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) BulkIndex(ctx context.Context, indexName string, docs []map[string]interface{}) error {
	var buf bytes.Buffer

	for i, doc := range docs {
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    fmt.Sprintf("%d_%d", doc["dataset_id"], i),
			},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		docJSON, _ := json.Marshal(doc)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/_bulk", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if errors, ok := result["errors"].(bool); ok && errors {
		return fmt.Errorf("bulk indexing completed with errors")
	}

	return nil
}

type SearchResult struct {
	Total   int64                    `json:"total"`
	Hits    []map[string]interface{} `json:"hits"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
}

func (c *Client) Search(ctx context.Context, indexName, query string, page, perPage int) (*SearchResult, error) {
	from := (page - 1) * perPage

	searchBody := map[string]interface{}{
		"from": from,
		"size": perPage,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"*"},
			},
		},
	}

	body, _ := json.Marshal(searchBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/"+indexName+"/_search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var esResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &esResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	hits := esResp["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	hitList := hits["hits"].([]interface{})

	var results []map[string]interface{}
	for _, hit := range hitList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		results = append(results, source)
	}

	return &SearchResult{
		Total:   int64(total),
		Hits:    results,
		Page:    page,
		PerPage: perPage,
	}, nil
}
