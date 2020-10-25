package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/ankur22/medium-picker/internal/logging"
)

var Version string

func main() {
	ctx := context.Background()
	ctx, sync := logging.NewContext(ctx)
	defer func() {
		if err := sync(); err != nil {
			logging.Error(ctx, "Can't sync logs", zap.Error(err))
		}
	}()

	logging.Info(ctx, "Starting medium-picker", zap.String("version", Version))

	logging.Info(ctx, "Shutting down medium-picker")
}
