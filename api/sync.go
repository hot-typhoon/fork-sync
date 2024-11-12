package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"unicode"

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

func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func ReadParamsFromQuery(queryParams url.Values) (*Params, error) {
	params := &Params{}
	missing := make([]string, 0)
	val := reflect.ValueOf(params).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		paramName := CamelToSnake(field.Name)
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

func response(w http.ResponseWriter, status int, message string) {
	h := w.Header()

	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(status)
	jsonResponse, _ := json.Marshal(map[string]string{"message": message})
	w.Write([]byte(jsonResponse))
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
		response(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	params, err := ReadParamsFromQuery(r.URL.Query())
	if err != nil {
		response(w, http.StatusBadRequest, err.Error())
		return
	}

	err = sync(params)
	if err != nil {
		response(w, http.StatusInternalServerError, err.Error())
		return
	}

	response(w, http.StatusOK, "OK")
}
