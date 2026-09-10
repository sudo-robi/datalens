package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/datalens/ingestion-service/elasticsearch"
	"github.com/datalens/ingestion-service/handler"
	"github.com/datalens/ingestion-service/postgres"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://datalens:datalens_dev_2024@localhost:5433/datalens?sslmode=disable")
	esURL := getEnv("ELASTICSEARCH_URL", "http://localhost:9200")
	port := getEnv("PORT", "8081")

	db, err := postgres.NewClient(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	es, err := elasticsearch.NewClient(esURL)
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	h := handler.NewHandler(db, es)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", h.IngestCSV)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /search", h.Search)

	corsHandler := corsMiddleware(mux)

	log.Printf("Ingestion service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// parseCSVRow parses a CSV row into a map using the header
func parseCSVRow(header []string, row []string) map[string]interface{} {
	result := make(map[string]interface{})
	for i, field := range header {
		if i < len(row) {
			val := strings.TrimSpace(row[i])
			if intVal, err := strconv.Atoi(val); err == nil {
				result[field] = intVal
			} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				result[field] = floatVal
			} else {
				result[field] = val
			}
		}
	}
	return result
}

// processCSV is a helper that reads CSV and returns rows as maps
func processCSV(reader io.Reader) ([]map[string]interface{}, []string, error) {
	csvReader := csv.NewReader(reader)

	header, err := csvReader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var rows []map[string]interface{}
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rows = append(rows, parseCSVRow(header, row))
	}

	return rows, header, nil
}

// ensureIndex creates the Elasticsearch index if it doesn't exist
func ensureIndex(ctx context.Context, es *elasticsearch.Client, indexName string) error {
	exists, err := es.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}
	if !exists {
		return es.CreateIndex(ctx, indexName)
	}
	return nil
}

// batchIndex indexes documents in batches for better performance
func batchIndex(ctx context.Context, es *elasticsearch.Client, indexName string, docs []map[string]interface{}) error {
	const batchSize = 100
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batch := docs[i:end]

		for j, doc := range batch {
			doc["index"] = i + j
			doc["ingested_at"] = time.Now().UTC()
		}

		if err := es.BulkIndex(ctx, indexName, batch); err != nil {
			return fmt.Errorf("failed to index batch at offset %d: %w", i, err)
		}
	}
	return nil
}
