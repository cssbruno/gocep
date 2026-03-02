package handlers

import (
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, http.StatusNotFound, "not_found", "resource not found")
}
