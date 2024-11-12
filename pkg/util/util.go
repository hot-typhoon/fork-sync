package util

import (
	"encoding/json"
	"net/http"
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

func HttpResponse(w http.ResponseWriter, status int, message string) {
	h := w.Header()

	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(status)
	jsonResponse, _ := json.Marshal(map[string]string{"message": message})
	w.Write([]byte(jsonResponse))
}
