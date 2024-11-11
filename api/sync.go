package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type RequestPayload struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Ref string `json:"ref"`
}

func getMessage(msg string) string {
	response := map[string]string{"message": msg}
	jsonResponse, _ := json.Marshal(response)
	return string(jsonResponse)
}

func sync(upstreamRepo, upstreamBranch, forkRepo, forkBranch, pat string) error {
	tempDir, err := os.MkdirTemp("", "repo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	upstreamUrl := fmt.Sprintf("https://github.com/%s.git", upstreamRepo)
	forkUrl := fmt.Sprintf("https://github.com/%s.git", forkRepo)

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           upstreamUrl,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + upstreamBranch),
		SingleBranch:  true,
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: pat,
		},
	})
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(tempDir)
	if err != nil {
		return err
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "fork",
		URLs: []string{forkUrl},
	})
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "fork",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/heads/%s", upstreamBranch, forkBranch)),
		},
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: pat,
		},
		Force: true,
	})
	return err
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, getMessage("Method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	params := r.URL.Query()
	forkRepo := params.Get("fork_repo")
	forkBranch := params.Get("fork_branch")
	upstreamRepo := params.Get("upstream_repo")
	upstreamBranch := params.Get("upstream_branch")
	pat := params.Get("pat")

	msg := make([]string, 0)

	switch {
	case forkRepo == "":
		msg = append(msg, "fork_repo")
	case forkBranch == "":
		msg = append(msg, "fork_branch")
	case upstreamRepo == "":
		msg = append(msg, "upstream_repo")
	case upstreamBranch == "":
		msg = append(msg, "upstream_branch")
	}

	if len(msg) != 0 {
		http.Error(w, getMessage(fmt.Sprintf("Please provide the following parameters: %v", msg)), http.StatusBadRequest)
		return
	}

	err := sync(upstreamRepo, upstreamBranch, forkRepo, forkBranch, pat)
	if err != nil {
		http.Error(w, getMessage(err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(getMessage("OK")))
}
