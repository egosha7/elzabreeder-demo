package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/egosha7/site-go/internal/initial"
	"github.com/egosha7/site-go/internal/metrics"

	log "github.com/egosha7/site-go/internal/logger"
	routes "github.com/egosha7/site-go/internal/router"
	"go.uber.org/zap"
)

// Глобальные переменные
var (
	// Version - это версия сборки приложения.
	Version string
	// BuildTime - это временная метка времени сборки приложения.
	BuildTime string
	// Commit - это хеш коммита приложения.
	Commit string
)

// Main - это основная точка входа.
func main() {
	fmt.Printf("Версия сборки: %s\n", Version)
	fmt.Printf("Дата сборки: %s\n", BuildTime)
	fmt.Printf("Коммит: %s\n", Commit)

	// Настройка логгера.
	logger, err := log.SetupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания логгера: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	cfg, h := initial.Initial(logger)

	// Настройка маршрутов для приложения.
	r := routes.SetupRoutes(h, logger)

	// Настройка обработки сигналов для грациозного завершения.
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	var wg sync.WaitGroup

	// Запуск горутины для обработки сигналов.
	go func() {
		sig := <-signalCh
		fmt.Printf("Получен сигнал %v. Завершение работы...\n", sig)

		// Дождемся завершения оставшихся запросов.
		wg.Wait()

		// Завершаем программу.
		os.Exit(0)
	}()

	// Сервер для метрик
	go metrics.StartMetricsServer("1721")

	fmt.Printf("Starting server on %s with cert: %s and key: %s\n", cfg.Host.Addr, cfg.Host.CertFile, cfg.Host.KeyFile)
	if err := http.ListenAndServeTLS(cfg.Host.Addr, cfg.Host.CertFile, cfg.Host.KeyFile, log.LogMiddleware(logger, r)); err != nil {
		logger.Error("Ошибка запуска HTTPS сервера", zap.Error(err))
		os.Exit(1)
	}
}
