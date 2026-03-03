package cep

import (
	"encoding/json"
	"errors"

	"github.com/cssbruno/gocep/models"
	"github.com/cssbruno/gocep/pkg/util"
)

var errUnknownParserSource = errors.New("unknown parser source")

type sourceParser func(body []byte) (models.CEPAddress, error)

var parserBySource = map[string]sourceParser{
	models.SourceCdnApiCep:        parseCDNApiCEP,
	models.SourceGitHubJeffotoni:  parseGitHubJeffotoni,
	models.SourceViaCep:           parseViaCEP,
	models.SourcePostmon:          parsePostmon,
	models.SourceRepublicaVirtual: parseRepublicaVirtual,
	models.SourceBrasilAPI:        parseBrasilAPI,
	models.SourceOpenCEP:          parseOpenCEP,
	models.SourceAwesomeAPI:       parseAwesomeAPI,
}

// ParseCEPAddress parses a provider response payload into the normalized address model.
func ParseCEPAddress(source string, body []byte) (models.CEPAddress, error) {
	parser, ok := parserBySource[source]
	if !ok {
		return models.CEPAddress{}, errUnknownParserSource
	}
	return parser(body)
}

func parseCDNApiCEP(body []byte) (models.CEPAddress, error) {
	var payload models.CdnApiCep
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.State, payload.Address, payload.District), nil
}

func parseGitHubJeffotoni(body []byte) (models.CEPAddress, error) {
	var payload models.GithubJeffotoni
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.StateCode, payload.Street, payload.Neighborhood), nil
}

func parseViaCEP(body []byte) (models.CEPAddress, error) {
	var payload models.ViaCep
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.StateCode, payload.Street, payload.Neighborhood), nil
}

func parsePostmon(body []byte) (models.CEPAddress, error) {
	var payload models.PostMon
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.State, payload.Street, payload.Neighborhood), nil
}

func parseRepublicaVirtual(body []byte) (models.CEPAddress, error) {
	var payload models.RepublicaVirtual
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.StateCode, payload.Street, payload.Neighborhood), nil
}

func parseBrasilAPI(body []byte) (models.CEPAddress, error) {
	var payload models.BrasilAPI
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.State, payload.Street, payload.Neighborhood), nil
}

func parseOpenCEP(body []byte) (models.CEPAddress, error) {
	var payload models.OpenCEP
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.StateCode, payload.Street, payload.Neighborhood), nil
}

func parseAwesomeAPI(body []byte) (models.CEPAddress, error) {
	var payload models.AwesomeAPI
	if err := decodeProviderPayload(body, &payload); err != nil {
		return models.CEPAddress{}, err
	}
	return buildAddress(payload.City, payload.StateCode, payload.Street, payload.Neighborhood), nil
}

func decodeProviderPayload(body []byte, dst any) error {
	return json.Unmarshal(body, dst)
}

func buildAddress(city, stateCode, street, neighborhood string) models.CEPAddress {
	return util.NormalizeAddress(models.CEPAddress{
		City:         city,
		StateCode:    stateCode,
		Street:       street,
		Neighborhood: neighborhood,
	})
}
