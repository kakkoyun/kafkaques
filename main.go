package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/kakkoyun/kafkaques/consumer"
	"github.com/kakkoyun/kafkaques/kafkaques"
	"github.com/kakkoyun/kafkaques/producer"
	"github.com/metalmatze/signal/internalserver"

	"github.com/alecthomas/kong"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-kit/log/level"
	"github.com/metalmatze/signal/healthcheck"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type Flags struct {
	LogLevel string `default:"info" enum:"error,warn,info,debug" help:"log level."`
	Address     string `default:":8080" help:"Address string for internal server"`
	Produce  struct {
		Brokers []string `kong:"required,help='Brokers.'"`
		Topic   string   `kong:"required,arg,name='topic',help='Topic push messages to.'"`
	} `cmd:"" help:"Consumer messages"`
	Consume struct {
		Brokers []string `kong:"required,help='Brokers.'"`
		Topic   string   `kong:"required,arg,name='topic',help='Topic to receive messages from.'"`
		Group   string   `kong:"help='Group.'"`
	} `cmd:"" help:"Produce messages"`
}

func main() {
	flags := &Flags{}
	kongCtx := kong.Parse(flags)

	serverStr := figure.NewColorFigure("Kafkaques", "roman", "yellow", true)
	serverStr.Print()

	logger := kafkaques.NewLogger(flags.LogLevel, kafkaques.LogFormatLogfmt, "kafkaques")
	level.Debug(logger).Log("msg", "kafkaques initialized",
		"version", version,
		"commit", commit,
		"date", date,
		"builtBy", builtBy,
		"config", fmt.Sprint(flags),
	)

	registry := prometheus.NewRegistry()
	healthchecks := healthcheck.NewMetricsHandler(healthcheck.NewHandler(), registry)
	h := internalserver.NewHandler(
		internalserver.WithHealthchecks(healthchecks),
		internalserver.WithPrometheusRegistry(registry),
		internalserver.WithPProf(),
	)
	s := http.Server{
		Addr:    flags.Address,
		Handler: h,
	}

	var g run.Group

	ctx, cancel := context.WithCancel(context.Background())
	switch kongCtx.Command() {
	case "produce <topic>":
		g.Add(func() error {
			return producer.Run(ctx, logger, flags.Produce.Brokers, flags.Produce.Topic)
		}, func(error) {
			cancel()
		})
	case "consume <topic>":
		g.Add(func() error {
			return consumer.Run(ctx, logger, flags.Consume.Brokers, flags.Consume.Group, flags.Consume.Topic)
		}, func(error) {
			cancel()
		})

	default:
		level.Error(logger).Log("err", "unknown command", "cmd", kongCtx.Command())
		os.Exit(1)
	}

	g.Add(func() error {
		level.Info(logger).Log("msg", "starting internal HTTP server", "address", s.Addr)
		return s.ListenAndServe()
	}, func(err error) {
		_ = s.Shutdown(context.Background())
	})

	g.Add(run.SignalHandler(ctx, os.Interrupt, os.Kill))
	if err := g.Run(); err != nil {
		var e run.SignalError
		if errors.As(err, &e) {
			level.Error(logger).Log("msg", "program exited with signal", "err", err, "signal", e.Signal)
		} else {
			level.Error(logger).Log("msg", "program exited with error", "err", err)
		}
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "exited")
}
