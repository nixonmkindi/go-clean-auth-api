package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/go-clean-auth-api/internal/app"
)

func main() {
	application, err := app.New(context.Background())
	if err != nil {
		// Bootstrap must succeed before the service can start.
		_, _ = os.Stderr.WriteString("application bootstrap error: " + err.Error() + "\n")
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() {
		addr := ":" + application.Config.AppPort
		application.Logger.Info("server_starting", "address", addr)
		serverErr <- application.Echo.Start(addr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			application.Logger.Error("server_error", "error", err.Error())
			os.Exit(1)
		}
	case <-quit:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			application.Logger.Error("server_shutdown_error", "error", err.Error())
			os.Exit(1)
		}
	}
}
