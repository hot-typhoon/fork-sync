package handler

import (
	"errors"
	"fmt"
	"hot-typhoon/sync/pkg/util"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Params struct {
	UpstreamRepo   string
	UpstreamBranch string
	ForkRepo       string
	ForkBranch     string
	Pat            string
}

func ReadParamsFromQuery(queryParams url.Values) (*Params, error) {
	params := &Params{}
	missing := make([]string, 0)
	val := reflect.ValueOf(params).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		paramName := util.CamelToSnake(field.Name)
		paramValue := queryParams.Get(paramName)
		if paramValue == "" {
			missing = append(missing, paramName)
		}
		val.Field(i).SetString(paramValue)
	}

	if len(missing) != 0 {
		return nil, fmt.Errorf("missing parameters: %v", missing)
	}

	return params, nil
}

func sync(params *Params) error {
	tempDir, err := os.MkdirTemp("", "repo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	upstreamUrl := fmt.Sprintf("https://github.com/%s.git", params.UpstreamRepo)
	forkUrl := fmt.Sprintf("https://github.com/%s.git", params.ForkRepo)

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           upstreamUrl,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + params.UpstreamBranch),
		SingleBranch:  true,
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: params.Pat,
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
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/heads/%s", params.UpstreamBranch, params.ForkBranch)),
		},
		Auth: &gitHttp.BasicAuth{
			Username: "fork-sync",
			Password: params.Pat,
		},
		Force: true,
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}

	return err
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.HttpResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	params, err := ReadParamsFromQuery(r.URL.Query())
	if err != nil {
		util.HttpResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err = sync(params)
	if err != nil {
		util.HttpResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.HttpResponse(w, http.StatusOK, "OK")
}
