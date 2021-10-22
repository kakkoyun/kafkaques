package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-kit/log/level"

	"kafkaques/kafkaques"
	"kafkaques/producer"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type flags struct {
	LogLevel string `default:"info" enum:"error,warn,info,debug" help:"log level."`
}

func main() {
	ctx := context.Background()
	flags := &flags{}
	kong.Parse(flags)

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

	err := producer.Run(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "Program exited with error", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "exited")
}
