package executor

import (
	"context"

	"github.com/danlju/tulip-worker/internal/model"
)

type BuildExecutor interface {
	Execute(ctx context.Context, build model.BuildRequest) (Result, error)
}

type Result struct {
	DurationMs int64
	Success    bool
	Logs       string
}
