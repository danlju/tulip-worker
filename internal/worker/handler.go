package worker

import (
	"context"
	"log"
	"time"

	"github.com/danlju/tulip-worker/internal/api"
	"github.com/danlju/tulip-worker/internal/executor"
	"github.com/danlju/tulip-worker/internal/queue"
)

type BuildHandler struct {
	api      *api.Client
	executor executor.BuildExecutor
}

func NewBuildHandler(apiClient *api.Client, exec executor.BuildExecutor) *BuildHandler {
	return &BuildHandler{
		api:      apiClient,
		executor: exec,
	}
}

func (h *BuildHandler) Handle(ctx context.Context, msg queue.Message) (err error) {
	build := msg.Payload
	buildID := build.Build.ID

	log.Printf("processing build %s", buildID)

	if err = h.api.MarkStarted(buildID); err != nil {
		return err
	}

	start := time.Now()

	defer func() {
		duration := time.Since(start).Milliseconds()

		if err != nil {
			if e := h.api.MarkFailed(buildID, duration); e != nil {
				log.Printf("failed to mark build %s as FAILED: %v", buildID, e)
			}
			return
		}

		if e := h.api.MarkCompleted(buildID, duration); e != nil {
			log.Printf("failed to mark build %s as COMPLETED: %v", buildID, e)
		}
	}()

	err = h.executor.Execute(ctx, build)
	return err
}
