package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/pkg/cep"
	"github.com/cssbruno/gocep/pkg/util"
)

func SearchCep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	cepstr := r.PathValue("cep")
	if strings.ContainsRune(cepstr, '/') {
		writeAPIError(w, http.StatusFound, "invalid_endpoint", "invalid endpoint")
		return
	}

	if err := util.CheckCEP(cepstr); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_cep", "cep must contain exactly 8 digits")
		return
	}

	result, wecep, err := cep.Search(cepstr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "search_error", "failed to search cep")
		return
	}

	if !cep.ValidCEP(wecep) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, result)
}
