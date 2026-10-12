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

	shouldRun, startErr := h.api.MarkStarted(buildID)
	if startErr != nil {
		err = startErr
		return err
	}
	if !shouldRun {
		log.Printf("build %s is already terminal; acknowledging duplicate delivery", buildID)
		return nil
	}

	start := time.Now()

	err = h.executor.Execute(ctx, build)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		buildErr := err
		if statusErr := h.api.MarkFailed(buildID, duration, buildErr.Error()); statusErr != nil {
			return statusErr
		}
		log.Printf("build %s failed: %v", buildID, buildErr)
		return nil
	}
	if err = h.api.MarkCompleted(buildID, duration); err != nil {
		return err
	}
	return nil
}
