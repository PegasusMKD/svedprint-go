package requestlogger

import (
	"context"
	"log"
	"time"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
)

type RequestLogRepository struct {
	queries *sqlc.Queries
}

func NewRequestLogRepository(queries *sqlc.Queries) *RequestLogRepository {
	return &RequestLogRepository{
		queries: queries,
	}
}

func (repository *RequestLogRepository) GetLogs() {
}

func (repository *RequestLogRepository) InsertBatch(logs []sqlc.BatchInsertRequestLogsParams) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := repository.queries.BatchInsertRequestLogs(ctx, logs)
	if err != nil {
		log.Printf("ERROR: Failed writing log batch due to '%v'", err)
		return
	}

	log.Printf("Successfully wrote %d log entries to the database.", len(logs))

}
