package model

import "time"

type BuildRequest struct {
	Version    int               `json:"version"`
	Build      Build             `json:"build"`
	Repository Repository        `json:"repository"`
	Source     Source            `json:"source"`
	Execution  Execution         `json:"execution"`
	Env        map[string]string `json:"env"`
	SentAt     time.Time         `json:"sentAt"`
}

type Build struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
}

type SourceProvider string

const (
	ProviderGeneric SourceProvider = "GENERIC"
	ProviderGithub  SourceProvider = "GITHUB"
	ProviderGitlab  SourceProvider = "GITLAB"
)

type Repository struct {
	CloneURL string         `json:"cloneUrl"`
	Provider SourceProvider `json:"provider"`
}

type Source struct {
	Branch        string `json:"branch"`
	CommitSHA     string `json:"commitSha"`
	CommitMessage string `json:"commitMessage"`
}

type Execution struct {
	ConfigPath string `json:"configPath"`
}
