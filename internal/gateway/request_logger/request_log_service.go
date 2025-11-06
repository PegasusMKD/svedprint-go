package requestlogger

import (
	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/pkg/logger"
	"github.com/rs/zerolog"
)

type RequestLogService struct {
	repository *RequestLogRepository
	logger     *zerolog.Logger
}

func NewRequestLogService(repository *RequestLogRepository) *RequestLogService {
	return &RequestLogService{
		repository: repository,
		logger:     logger.Get(),
	}
}

func (service *RequestLogService) CreateLogs(logs []sqlc.RequestLog) {
	data := make([]sqlc.BatchInsertRequestLogsParams, len(logs))

	for idx, log := range logs {
		data[idx] = sqlc.BatchInsertRequestLogsParams{
			Timestamp:       log.Timestamp,
			Method:          log.Method,
			IncomingPath:    log.IncomingPath,
			RedirectedPath:  log.RedirectedPath,
			OrganizationID:  log.OrganizationID,
			UserID:          log.UserID,
			StatusCode:      log.StatusCode,
			ResponseTimeMs:  log.ResponseTimeMs,
			UpstreamService: log.UpstreamService,
			ErrorMessage:    log.ErrorMessage,
		}
	}

	service.repository.InsertBatch(data)
}
