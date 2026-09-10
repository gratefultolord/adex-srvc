package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	auctionHandler "github.com/gratefultolord/adex-srvc/internal/api/handlers/auction"
	dspClient "github.com/gratefultolord/adex-srvc/internal/client/dsp"
	"github.com/gratefultolord/adex-srvc/internal/storage"
	auctionUsecase "github.com/gratefultolord/adex-srvc/internal/usecases/auction"
)

const (
	defaultHTTPAddr        = ":8080"
	defaultDSPTimeout      = 200 * time.Millisecond
	defaultShutdownTimeout = 10 * time.Second
)

func main() {
	_ = godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	logger, err := newLogger(cfg.logLevel)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := sqlx.ConnectContext(
		ctx,
		"postgres",
		cfg.databaseURL,
	)
	if err != nil {
		logger.Error(
			"sqlx.ConnectContext",
			zap.Error(err),
		)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(
				"db.Close",
				zap.Error(err),
			)
		}
	}()

	partnerStorage := storage.NewStorage(db)

	httpClient := &http.Client{}

	dsp := dspClient.NewClient(httpClient)

	auctionUC := auctionUsecase.NewUsecase(
		partnerStorage,
		dsp,
		cfg.dspTimeout,
	)

	auctionHTTPHandler := auctionHandler.NewHandler(
		logger,
		auctionUC,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		}); err != nil {
			logger.Error("json.NewEncoder.Encode", zap.Error(err))
		}
	})

	mux.HandleFunc(
		"POST /auction",
		auctionHTTPHandler.Handle,
	)

	server := &http.Server{
		Addr:              cfg.httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info(
		"server starting",
		zap.String("addr", cfg.httpAddr),
		zap.Duration("dsp_timeout", cfg.dspTimeout),
		zap.Duration("shutdown_timeout", cfg.shutdownTimeout),
		zap.String("log_level", cfg.logLevel),
	)

	if err := serve(
		ctx,
		server,
		logger,
		cfg.shutdownTimeout,
	); err != nil {
		logger.Error(
			"serve",
			zap.Error(err),
		)
	}
}

type config struct {
	databaseURL     string
	httpAddr        string
	logLevel        string
	dspTimeout      time.Duration
	shutdownTimeout time.Duration
}

func loadConfig() (config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	dspTimeout, err := getDurationEnv(
		"DSP_TIMEOUT",
		defaultDSPTimeout,
	)
	if err != nil {
		return config{}, fmt.Errorf(
			"getDurationEnv DSP_TIMEOUT: %w",
			err,
		)
	}

	shutdownTimeout, err := getDurationEnv(
		"SHUTDOWN_TIMEOUT",
		defaultShutdownTimeout,
	)
	if err != nil {
		return config{}, fmt.Errorf(
			"getDurationEnv SHUTDOWN_TIMEOUT: %w",
			err,
		)
	}

	httpAddr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	logLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	return config{
		databaseURL:     databaseURL,
		httpAddr:        httpAddr,
		logLevel:        logLevel,
		dspTimeout:      dspTimeout,
		shutdownTimeout: shutdownTimeout,
	}, nil
}

func getDurationEnv(
	key string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("time.ParseDuration: %w", err)
	}

	return duration, nil
}

func newLogger(levelValue string) (*zap.Logger, error) {
	level := zapcore.InfoLevel

	if err := level.Set(strings.ToLower(levelValue)); err != nil {
		return nil, fmt.Errorf("level.Set: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("cfg.Build: %w", err)
	}

	return logger, nil
}

func serve(
	ctx context.Context,
	server *http.Server,
	logger *zap.Logger,
	shutdownTimeout time.Duration,
) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("server.ListenAndServe: %w", err)

	case <-ctx.Done():
		logger.Info("shutdown started")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				logger.Error(
					"server.Close",
					zap.Error(closeErr),
				)
			}

			return fmt.Errorf("server.Shutdown: %w", err)
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server.ListenAndServe: %w", err)
		}

		logger.Info("shutdown completed")

		return nil
	}
}
