package main

import (
	"context"
	"fmt"
	"os"

	"kafkaques/consumer"
	"kafkaques/kafkaques"
	"kafkaques/producer"

	"github.com/alecthomas/kong"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type flags struct {
	LogLevel string `default:"info" enum:"error,warn,info,debug" help:"log level."`

	Produce struct {
	} `cmd:"" help:"Produce messages"`

	Consume struct {
	} `cmd:"" help:"Consumer messages"`
}

func main() {
	flags := &flags{}
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

	var (
		g run.Group
	)

	ctx, cancel := context.WithCancel(context.Background())
	switch kongCtx.Command() {
	case "produce":
		g.Add(func() error {
			return producer.Run(ctx)
		}, func(error) {
			cancel()
		})
	case "consumer":
		g.Add(func() error {
			return consumer.Run(ctx)
		}, func(error) {
			cancel()
		})

	default:
		level.Error(logger).Log("err", "unknown command", "cmd", kongCtx.Command())
		os.Exit(1)
	}

	g.Add(run.SignalHandler(ctx, os.Interrupt, os.Kill))
	if err := g.Run(); err != nil {
		level.Error(logger).Log("msg", "program exited with error", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "exited")
}
