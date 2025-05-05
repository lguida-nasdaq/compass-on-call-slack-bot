package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/app"
	"github.com/metriodev/pompiers/internal/server"
)

type Cli struct {
	Run     RunCMD `cmd:"" description:"Run the server"`
	Timeout int    `default:"5000" help:"Timeout for shutdown in seconds"`
}

type RunCMD struct {
	Port               int    `default:"8080" help:"Port to run the server on"`
	Host               string `help:"Host to run the server on"`
	AtlassianApiKey    string `required:"" help:"Atlassian API key"`
	AtlassianApiUser   string `required:"" help:"Atlassian API user"`
	AtlassianCloudId   string `required:"" help:"Atlassian cloud ID"`
	SlackSigningSecret string `required:"" help:"Slack signing secret"`
}

func (r RunCMD) Run(cli *Cli) error {
	app := app.NewApp(
		api.NewCompassClient(r.AtlassianApiUser, r.AtlassianApiKey, r.AtlassianCloudId),
		api.NewJiraClient(r.AtlassianApiUser, r.AtlassianApiKey),
	)

	srv := server.NewServer(app, r.Host, r.Port, r.SlackSigningSecret)
	if err := srv.Start(); err != nil {
		slog.Error("Error starting server", slog.String("error", err.Error()))
		return fmt.Errorf("Error starting server: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-signalChan:
		slog.Info("Termination signal received, shutting down...", slog.String("signal", s.String()))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cli.Timeout)*time.Millisecond)
	defer shutdownCancel()

	err := srv.Stop(shutdownCtx)
	if err != nil {
		return fmt.Errorf("s.Stop: %v", err)
	}

	slog.Info("Shut down complete")
	return nil
}

func main() {
	slog.Info("Starting CLI")
	var cli Cli
	kctx := kong.Parse(&cli, kong.DefaultEnvars(""))
	err := kctx.Run()
	kctx.FatalIfErrorf(err)
}
