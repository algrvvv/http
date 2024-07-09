package main

import (
	"github.com/algrvvv/http/internal"
	"github.com/algrvvv/http/internal/logger"
)

func main() {
	// получаем реквест
	req, err := internal.ParseUtilOptions()
	if err != nil {
		logger.Logger(err, logger.ExitLogType)
	}

	// проверяем на его заполненость
	// он будет пустой, если во время во время запуска
	// был использован флаг -help / -version
	if req.URL == "" {
		return
	}

	logger.Logger(req, logger.InfoLogType)
}
