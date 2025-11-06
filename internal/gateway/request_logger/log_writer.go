package requestlogger

import (
	"log"
	"time"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
)

type LogWriter struct {
	service *RequestLogService
	done    chan any
	logChan chan sqlc.RequestLog
}

func NewLogWriter(service *RequestLogService) *LogWriter {
	writer := &LogWriter{
		service: service,
		done:    make(chan any),
		logChan: make(chan sqlc.RequestLog, 1000),
	}

	writer.start()

	return writer
}

func (lw *LogWriter) start() {
	go lw.worker()
}

func (lw *LogWriter) Stop() {
	close(lw.done)
	time.Sleep(100 * time.Millisecond)
}

func (lw *LogWriter) Write(entry sqlc.RequestLog) {
	select {
	case lw.logChan <- entry:
	default:
		log.Println("[WARN] Log channel is full, dropping entry.")
	}
}

func (lw *LogWriter) worker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	batch := make([]sqlc.RequestLog, 0, 100)

	for {
		select {
		case <-lw.done:
			if len(batch) > 0 {
				lw.service.CreateLogs(batch)
			}
			return
		case logEntry := <-lw.logChan:
			batch = append(batch, logEntry)

			if len(batch) >= 100 {
				lw.service.CreateLogs(batch)
				batch = make([]sqlc.RequestLog, 0, 100)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				lw.service.CreateLogs(batch)
				batch = make([]sqlc.RequestLog, 0, 100)
			}
		}
	}
}
