package requestlogger

import "github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"

type LogWriter struct {
	service *RequestLogService
	done    chan any
	logChan chan sqlc.RequestLog
}
