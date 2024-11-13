package handler

import (
	"hot-typhoon/sync/pkg/sync"
	"hot-typhoon/sync/pkg/util"
	"net/http"
)

type QueryParams struct {
	Owner  string
	Repo   string
	Branch string
	Pat    string
	// IsShallow string `sync:"false"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.HttpResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	params, err := util.ReadParamsFromQuery[QueryParams](r.URL.Query())
	if err != nil {
		util.HttpResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	errByAPI := sync.SyncByAPI(params.Owner, params.Repo, params.Branch, params.Pat)
	if errByAPI != nil {
		errByGit := sync.SyncByGit(params.Owner, params.Repo, params.Branch, params.Pat)
		if errByGit != nil {
			util.HttpResponse(w, http.StatusInternalServerError, []string{errByAPI.Error(), errByGit.Error()})
			return
		}
		util.HttpResponse(w, http.StatusOK, errByAPI.Error())
		return
	}

	util.HttpResponse(w, http.StatusOK, "OK")
}
