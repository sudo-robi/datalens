package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/datalens/ingestion-service/elasticsearch"
	"github.com/datalens/ingestion-service/postgres"
)

type Handler struct {
	db  *postgres.Client
	es  *elasticsearch.Client
}

func NewHandler(db *postgres.Client, es *elasticsearch.Client) *Handler {
	return &Handler{db: db, es: es}
}

type IngestRequest struct {
	DatasetID int64  `json:"dataset_id"`
	Filename  string `json:"filename"`
}

type SearchRequest struct {
	Query   string `json:"query"`
	Index   string `json:"index"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"status":  "healthy",
			"service": "ingestion-service",
		},
	})
}

func (h *Handler) IngestCSV(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	datasetIDStr := r.FormValue("dataset_id")
	datasetID, err := strconv.ParseInt(datasetIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid dataset_id")
		return
	}

	csvReader := csv.NewReader(file)
	csvHeader, err := csvReader.Read()
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to read CSV header")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	indexName := fmt.Sprintf("dataset_%d", datasetID)
	exists, err := h.es.IndexExists(ctx, indexName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check index")
		return
	}
	if !exists {
		if err := h.es.CreateIndex(ctx, indexName); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create index")
			return
		}
	}

	var rowCount int64
	const batchSize = 100
	var batch []map[string]interface{}

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		doc := parseCSVRow(csvHeader, row)
		doc["dataset_id"] = datasetID
		doc["ingested_at"] = time.Now().UTC()
		batch = append(batch, doc)

		if len(batch) >= batchSize {
			if err := h.es.BulkIndex(ctx, indexName, batch); err != nil {
				log.Printf("Failed to index batch: %v", err)
			}
			rowCount += int64(len(batch))
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := h.es.BulkIndex(ctx, indexName, batch); err != nil {
			log.Printf("Failed to index final batch: %v", err)
		}
		rowCount += int64(len(batch))
	}

	if err := h.db.UpdateDatasetRowCount(ctx, datasetID, rowCount); err != nil {
		log.Printf("Failed to update row count: %v", err)
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"rows_ingested": rowCount,
			"filename":      header.Filename,
			"index":         indexName,
		},
	})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Index == "" {
		req.Index = "datasets"
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 20
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results, err := h.es.Search(ctx, req.Index, req.Query, req.Page, req.PerPage)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    results,
	})
}

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

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, APIResponse{
		Success: false,
		Error:   message,
	})
}
