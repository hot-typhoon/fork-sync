package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode"
)

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

func HttpResponse(w http.ResponseWriter, status int, message any) {
	h := w.Header()

	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(status)
	jsonResponse, _ := json.Marshal(map[string]any{"message": message})
	w.Write([]byte(jsonResponse))
}

func ReadParamsFromQuery[T any](queryParams url.Values) (*T, error) {
	params := new(T)
	missing := make([]string, 0)
	val := reflect.ValueOf(params).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		paramName := CamelToSnake(field.Name)
		paramValue := queryParams.Get(paramName)
		if paramValue == "" {
			if field.Tag.Get("sync") != "" {
				paramValue = field.Tag.Get("sync")
			} else {
				missing = append(missing, paramName)
			}
		}
		val.Field(i).SetString(paramValue)
	}

	if len(missing) != 0 {
		return nil, fmt.Errorf("missing parameters: %v", missing)
	}

	return params, nil
}
