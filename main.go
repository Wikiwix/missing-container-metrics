package main

import (
	"context"
	"net/http"

	"github.com/draganm/missing-container-metrics/docker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var Version string

func main() {
	a := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "bind-address",
				Value: ":3001",
				EnvVars: []string{
					"BIND_ADDRESS",
				},
			},
		},
		Action: func(c *cli.Context) error {
			logger, err := zap.NewProduction()
			if err != nil {
				return err
			}

			slogger := logger.Sugar().With("version", Version)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				err := docker.HandleDocker(ctx, slogger)
				if err != nil {
					slogger.With("error", err).Error("while handling docker")
					cancel()
				}
			}()

			slogger.Info("started")

			a := c.String("bind-address")

			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())

			server := &http.Server{
				Addr:    a,
				Handler: mux,
			}

			// Close server when the context gets cancelled
			go func() {
				<-ctx.Done()
				server.Close()
			}()

			slogger.Infof("Listening on %s", a)
			return server.ListenAndServe()

		},
	}

	a.RunAndExitOnError()

}
