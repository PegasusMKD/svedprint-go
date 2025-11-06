package requestlogger

import (
	"context"
	"time"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/rs/zerolog"
)

type RequestLogRepository struct {
	queries *sqlc.Queries
	logger  *zerolog.Logger
}

func NewRequestLogRepository(queries *sqlc.Queries) *RequestLogRepository {
	return &RequestLogRepository{
		queries: queries,
		logger:  logger.Get(),
	}
}

func (repository *RequestLogRepository) GetLogs() {
}

func (repository *RequestLogRepository) InsertBatch(logs []sqlc.BatchInsertRequestLogsParams) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := repository.queries.BatchInsertRequestLogs(ctx, logs)
	if err != nil {
		repository.logger.Printf("ERROR: Failed writing log batch due to '%v'", err)
		return
	}

	repository.logger.Printf("Successfully wrote %d log entries to the database.", len(logs))

}
