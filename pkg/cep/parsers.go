package cep

import (
	"encoding/json"
	"errors"

	"github.com/cssbruno/gocep/models"
)

var errUnknownParserSource = errors.New("unknown parser source")

func ParseWeCep(source string, body []byte) (models.WeCep, error) {
	wecep := models.WeCep{}
	switch source {
	case models.SourceCdnApiCep:
		var cdnapi models.CdnApiCep
		if err := json.Unmarshal(body, &cdnapi); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = cdnapi.City
		wecep.StateCode = cdnapi.State
		wecep.Street = cdnapi.Address
		wecep.Neighborhood = cdnapi.District
	case models.SourceGitHubJeffotoni:
		var githubjeff models.GithubJeffotoni
		if err := json.Unmarshal(body, &githubjeff); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = githubjeff.City
		wecep.StateCode = githubjeff.StateCode
		wecep.Street = githubjeff.Street
		wecep.Neighborhood = githubjeff.Neighborhood
	case models.SourceViaCep:
		var viacep models.ViaCep
		if err := json.Unmarshal(body, &viacep); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = viacep.City
		wecep.StateCode = viacep.StateCode
		wecep.Street = viacep.Street
		wecep.Neighborhood = viacep.Neighborhood
	case models.SourcePostmon:
		var postmon models.PostMon
		if err := json.Unmarshal(body, &postmon); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = postmon.City
		wecep.StateCode = postmon.State
		wecep.Street = postmon.Street
		wecep.Neighborhood = postmon.Neighborhood
	case models.SourceRepublicaVirtual:
		var repub models.RepublicaVirtual
		if err := json.Unmarshal(body, &repub); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = repub.City
		wecep.StateCode = repub.StateCode
		wecep.Street = repub.Street
		wecep.Neighborhood = repub.Neighborhood
	case models.SourceBrasilAPI:
		var brasilapi models.BrasilAPI
		if err := json.Unmarshal(body, &brasilapi); err != nil {
			return models.WeCep{}, err
		}
		wecep.City = brasilapi.City
		wecep.StateCode = brasilapi.State
		wecep.Street = brasilapi.Street
		wecep.Neighborhood = brasilapi.Neighborhood
	default:
		return models.WeCep{}, errUnknownParserSource
	}
	return wecep, nil
}
