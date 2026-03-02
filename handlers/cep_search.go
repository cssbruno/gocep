package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/cssbruno/gocep/pkg/cep"
	"github.com/cssbruno/gocep/pkg/util"
)

var searchCEP = cep.Search

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

	normalizedCEP, err := util.NormalizeCEP(cepstr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_cep", "cep must be in 00000000 or 00000-000 format")
		return
	}

	result, address, err := searchCEP(normalizedCEP)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "search_error", "failed to search cep")
		return
	}

	if !cep.ValidCEP(address) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, result)
}
