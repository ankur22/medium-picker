package main

import (
	"context"

	"github.com/ankur22/medium-picker/internal/logging"
	"go.uber.org/zap"
)

var Version string

func main() {
	ctx := context.Background()
	ctx, sync := logging.NewContext(ctx)
	defer sync()

	logging.Info(ctx, "Starting medium-picker", zap.String("version", Version))

	logging.Info(ctx, "Shutting down medium-picker")
}
