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

func NewBuildHandler(apiClient *api.Client) *BuildHandler {
	return &BuildHandler{api: apiClient}
}

func (h *BuildHandler) Handle(ctx context.Context, msg queue.Message) error {
	build := msg.Payload
	buildID := build.Build.ID

	log.Printf("processing build %s", buildID)

	if err := h.api.MarkStarted(buildID); err != nil {
		return err
	}

	start := time.Now()

	var handlerErr error

	defer func() {
		duration := time.Since(start).Milliseconds()

		if handlerErr != nil {
			if err := h.api.MarkFailed(buildID, duration); err != nil {
				log.Printf("failed to mark build %s as FAILED: %v", buildID, err)
			}
			return
		}

		if err := h.api.MarkCompleted(buildID, duration); err != nil {
			log.Printf("failed to mark build %s as COMPLETED: %v", buildID, err)
		}
	}()

	time.Sleep(3 * time.Second)

	return nil
}
