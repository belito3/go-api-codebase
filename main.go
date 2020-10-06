package main

import (
	"context"
	"flag"
	"github.com/belito3/go-api-codebase/app"
	"github.com/belito3/go-api-codebase/pkg/logger"
)

// VERSION: go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "1.1.0"
func main() {
	var fileConf string
	flag.StringVar(&fileConf, "config", `./app/config/config.yaml`, "Absolute path of configuration file")
	flag.Parse()

	logger.SetVersion(VERSION)
	// Attach TraceID to context
	ctx := logger.NewTraceIDContext(context.Background(), "main")
	err := app.Run(ctx,
		app.SetConfigFile(fileConf),
		app.SetVersion(VERSION))
	if err != nil {
		logger.Errorf(ctx, err.Error())
	}
}
