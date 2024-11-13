package sync

import (
	"hot-typhoon/sync/pkg/program"

	"github.com/google/go-github/v66/github"
)

func SyncByAPI(owner, repo, branch, pat string) error {
	_, _, err := github.NewClientWithEnvProxy().WithAuthToken(pat).Repositories.MergeUpstream(program.Ctx, owner, repo, &github.RepoMergeUpstreamRequest{
		Branch: &branch,
	})
	return err
}
