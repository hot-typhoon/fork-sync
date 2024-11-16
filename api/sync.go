package handler

import (
	"hot-typhoon/sync/pkg/sync"
	"hot-typhoon/sync/pkg/util"
	"net/http"
)

type QueryParams struct {
	Owner   string
	Repo    string
	Branch  string
	Pat     string
	Only    string `sync:"none"`
	Shallow string `sync:"false"`
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

	errs := make([]string, 0)

	switch params.Only {
	case "git":
		errByGit := sync.SyncByGit(params.Owner, params.Repo, params.Branch, params.Pat, params.Shallow == "true")
		if errByGit != nil {
			errs = append(errs, errByGit.Error())
		}
	case "api":
		errByAPI := sync.SyncByAPI(params.Owner, params.Repo, params.Branch, params.Pat)
		if errByAPI != nil {
			errs = append(errs, errByAPI.Error())
		}
	default:
		errByAPI := sync.SyncByAPI(params.Owner, params.Repo, params.Branch, params.Pat)
		if errByAPI != nil {
			errs = append(errs, errByAPI.Error())
			errByGit := sync.SyncByGit(params.Owner, params.Repo, params.Branch, params.Pat, params.Shallow == "true")
			if errByGit != nil {
				errs = append(errs, errByGit.Error())
			}
		}
	}

	if len(errs) > 0 {
		util.HttpResponse(w, http.StatusInternalServerError, errs)
	} else {
		util.HttpResponse(w, http.StatusOK, "OK")
	}
}
