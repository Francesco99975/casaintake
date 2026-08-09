package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/config"
	"github.com/labstack/echo/v4"

	"github.com/Francesco99975/casaintake/internal/helpers"
)

func main() {
	err := boot.LoadEnvVariables()
	if err != nil {
		panic(err)
	}

	slog.SetDefault(boot.NewLogger())

	if err := config.LoadManifest("./static"); err != nil {
		log.Fatalf("Failed to load Vite manifest: %v", err)
	}

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	port := boot.Environment.Port

	e := createRouter()

	go func() {
		slog.Info("Starting Server",
			slog.String("framework", "echo"),
			slog.String("version", echo.Version),
			slog.String("port", port),
		)
		slog.Info("Local Access", slog.String("url", fmt.Sprintf("http://localhost:%s", port)))
		slog.Info("Internet Access", slog.String("url", boot.Environment.URL))
		slog.Info("Press Ctrl+C to stop the server and exit.")
		err := e.Start(":" + port)
		boot.FatalLog("Server exited, could not start", err)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	helpers.Notify("casaintake", "Server is shutting down")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		helpers.Notify("casaintake", fmt.Sprintf("Server forced to shutdown: %v", err))
		boot.FatalLog("Server forced to shutdown", err)
	}
}
