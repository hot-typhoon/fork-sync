package sync

import (
	"context"

	"github.com/google/go-github/v66/github"
)

func SyncByAPI(owner, repo, branch, pat string) error {
	_, _, err := github.NewClientWithEnvProxy().WithAuthToken(pat).Repositories.MergeUpstream(context.Background(), owner, repo, &github.RepoMergeUpstreamRequest{
		Branch: &branch,
	})
	return err
}
