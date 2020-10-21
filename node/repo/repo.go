package repo

import (
	"golang.org/x/xerrors"
)

// New instantiates a lotus Repo
func New(repoType, repoID string) (Repo, error) {
	switch repoType {
	case "fs":
		return NewFS(repoID)
	default:
		return nil, xerrors.Errorf("Unknown repo type: %s", repoType)
	}
}
