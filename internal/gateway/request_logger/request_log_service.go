package requestlogger

import "github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"

type RequestLogService struct {
	repository *RequestLogRepository
}

func (service *RequestLogService) CreateLogs(logs []sqlc.RequestLog) {
	data := make([]sqlc.BatchInsertRequestLogsParams, 100)

	for idx, log := range logs {
		data[idx] = sqlc.BatchInsertRequestLogsParams{
			Timestamp:      log.Timestamp,
			Method:         log.Method,
			IncomingPath:   log.IncomingPath,
			RedirectedPath: log.RedirectedPath,
			// TODO: ....
		}
	}

	service.repository.InsertBatch(data)
}
