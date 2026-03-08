package application

import (
	"context"
	"database/sql"
	httpadapter "marketflow/internal/adapters/http"
	"marketflow/internal/adapters/sqlite"
	"marketflow/internal/service"
	"marketflow/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Application struct {
	service *service.Service
	server  *http.Server
	logger  *logger.Logger
}

func GetApp() (*Application, error) {
	appLogger := logger.NewLogger("INFO")
	db, err := sql.Open("sqlite3", "marketflow.db")
	if err != nil {
		return nil, err
	}

	// Создание таблиц
	db.Exec(`CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		exchange TEXT NOT NULL,
		price REAL NOT NULL,
		timestamp INTEGER NOT NULL
	);`)
	db.Exec(`CREATE TABLE IF NOT EXISTS aggr_price_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pair_name TEXT NOT NULL,
		exchange TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		average_price REAL NOT NULL,
		min_price REAL NOT NULL,
		max_price REAL NOT NULL
	);`)

	pricerepo := sqlite.GetPriceRepositorySqlite(db)
	aggrrepo := sqlite.GetAggrPriceDataRepositorySqlite(db)
	svc := service.GetService(pricerepo, aggrrepo, []string{"40101", "40102", "40103"})

	app := &Application{
		service: svc,
		logger:  appLogger,
	}
	app.server = app.GetServer()
	return app, nil
}

func (a *Application) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Перехват Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервиса
	go a.service.Start(ctx)

	// Запуск HTTP сервера
	go func() {
		a.logger.Info("HTTP server starting", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("HTTP server failed", "error", err)
		}
	}()

	// Ждём сигнал или отмену контекста
	select {
	case sig := <-sigCh:
		a.logger.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		a.logger.Info("Application context canceled")
	}

	// Отмена контекста (остановка сервиса и воркеров)
	cancel()

	// Graceful shutdown HTTP сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("HTTP server shutdown error", "error", err)
	} else {
		a.logger.Info("HTTP server gracefully stopped")
	}

	// Остановка сервиса
	a.service.Stop()
	a.logger.Info("Application gracefully finished")
}

func (a *Application) GetServer() *http.Server {
	mux := http.NewServeMux()
	handler := httpadapter.GetHandler(a.service)
	mux.HandleFunc("/prices/highest/", handler.GetHighestPriceHandler)

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}
