package executor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/danlju/tulip-worker/internal/model"
)

type BuildExecutor interface {
	Execute(ctx context.Context, build model.BuildRequest) error
}

type Executor struct {
	workspaceRoot string
}

func NewExecutor(root string) *Executor {
	return &Executor{
		workspaceRoot: root,
	}
}

type Result struct {
	DurationMs int64
	Success    bool
	Logs       string
}

func (e *Executor) Execute(ctx context.Context, req model.BuildRequest) error {
	workspace := filepath.Join(e.workspaceRoot, req.Build.ID)

	if err := os.RemoveAll(workspace); err != nil {
		return err
	}
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return err
	}

	log.Printf("[build:%s] workspace: %s", req.Build.ID, workspace)

	if err := e.clone(ctx, req.Repository.CloneURL, workspace); err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	if err := e.checkout(ctx, workspace, req.Source.CommitSHA, req.Source.Branch); err != nil {
		return fmt.Errorf("checkout failed: %w", err)
	}

	if err := e.runContainer(ctx, req, workspace); err != nil {
		return fmt.Errorf("container run failed: %w", err)
	}

	return nil
}

func (e *Executor) clone(ctx context.Context, repoURL, workspace string) error {
	log.Printf("cloning repo: %s", repoURL)

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, workspace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e *Executor) checkout(ctx context.Context, workspace, sha, branch string) error {
	if sha == "" && branch == "" {
		return nil
	}

	if sha != "" {
		log.Printf("checking out commit: %s", sha)

		cmd := exec.CommandContext(ctx, "git", "-C", workspace, "checkout", sha)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	log.Printf("checking out branch: %s", branch)

	cmd := exec.CommandContext(ctx, "git", "-C", workspace, "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()

}

func (e *Executor) runContainer(ctx context.Context, req model.BuildRequest, workspace string) error {
	log.Printf("[build:%s] starting container", req.Build.ID)

	args := []string{
		"run",
		"--rm",

		"-v", fmt.Sprintf("%s:/workspace", workspace),
		"-w", "/workspace",

		"alpine:3.19",

		"sh", "-c", "echo 'container started successfully' && ls -la",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
