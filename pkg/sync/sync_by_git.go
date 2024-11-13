package sync

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v66/github"
)

func SyncByGit(owner, repo, branch, pat string) error {
	forkRepo, _, err := github.NewClientWithEnvProxy().WithAuthToken(pat).Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return err
	}

	parentRepo := forkRepo.GetParent()
	if parentRepo == nil {
		return errors.New("Parent repository not found")
	}

	tempDir, err := os.MkdirTemp("", "repo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           parentRepo.GetCloneURL(),
		ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
		SingleBranch:  true,
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: pat,
		},
	})
	if err != nil {
		return err
	}

	repository, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: "fork",
		URLs: []string{forkRepo.GetCloneURL()},
	})
	if err != nil {
		return err
	}

	err = repository.Push(&git.PushOptions{
		RemoteName: "fork",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/heads/%s", branch, branch)),
		},
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: pat,
		},
		Force: true,
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}
