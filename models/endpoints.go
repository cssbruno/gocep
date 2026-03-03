package models

import "sync"

// Endpoint configures a single CEP provider request.
type Endpoint struct {
	// Method is the HTTP method used by the provider.
	Method string
	// Source identifies the parser used for the provider response.
	Source string
	// URL is the endpoint template; CEP must be injected with "%s".
	URL string
	// Body is an optional request payload template, used by SOAP providers.
	Body string
}

// HTTP method constants used by endpoint definitions.
const (
	MethodGet  = "GET"
	MethodPost = "POST"
)

// Source constants identify provider parsers.
const (
	SourceCdnApiCep        = "cdnapicep"
	SourceGitHubJeffotoni  = "githubjeffotoni"
	SourceViaCep           = "viacep"
	SourcePostmon          = "postmon"
	SourceRepublicaVirtual = "republicavirtual"
	SourceCorreio          = "correio"
	SourceBrasilAPI        = "brasilapi"
	SourceOpenCEP          = "opencep"
	SourceAwesomeAPI       = "awesomeapi"
)

// Provider endpoint templates and static payloads.
const (
	URLCdnApiCep        = "https://cdn.apicep.com/file/apicep/%s.json"
	URLGitHubJeffotoni  = "https://raw.githubusercontent.com/jeffotoni/api.cep/master/v1/cep/%s"
	URLViaCep           = "https://viacep.com.br/ws/%s/json/"
	URLPostmon          = "https://api.postmon.com.br/v1/cep/%s"
	URLRepublicaVirtual = "https://republicavirtual.com.br/web_cep.php?cep=%s&formato=json"
	URLCorreiosService  = "https://apps.correios.com.br/SigepMasterJPA/AtendeClienteService/AtendeCliente"
	URLBrasilAPI        = "https://brasilapi.com.br/api/cep/v1/%s"
	URLOpenCEP          = "https://opencep.com/v1/%s.json"
	URLAwesomeAPI       = "https://cep.awesomeapi.com.br/json/%s"
	PayloadCorreio      = `<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
    <x:Body>
        <cli:consultaCEP>
            <cep>%s</cep>
        </cli:consultaCEP>
    </x:Body>
</x:Envelope>`
)

var (
	endpointsMu sync.RWMutex

	// Endpoints contains provider configurations used by CEP search.
	// Prefer SetEndpoints/GetEndpoints for concurrent-safe updates.
	Endpoints = []Endpoint{
		{Method: MethodGet, Source: SourceCdnApiCep, URL: URLCdnApiCep},
		{Method: MethodGet, Source: SourceGitHubJeffotoni, URL: URLGitHubJeffotoni},
		{Method: MethodGet, Source: SourceViaCep, URL: URLViaCep},
		{Method: MethodGet, Source: SourcePostmon, URL: URLPostmon},
		{Method: MethodGet, Source: SourceRepublicaVirtual, URL: URLRepublicaVirtual},
		{Method: MethodPost, Source: SourceCorreio, URL: URLCorreiosService, Body: PayloadCorreio},
		{Method: MethodGet, Source: SourceBrasilAPI, URL: URLBrasilAPI},
		{Method: MethodGet, Source: SourceOpenCEP, URL: URLOpenCEP},
		{Method: MethodGet, Source: SourceAwesomeAPI, URL: URLAwesomeAPI},
	}
)

// GetEndpoints returns a copy of the configured provider endpoints.
func GetEndpoints() []Endpoint {
	endpointsMu.RLock()
	defer endpointsMu.RUnlock()
	return cloneEndpoints(Endpoints)
}

// SetEndpoints replaces provider endpoints using an internal copy.
func SetEndpoints(next []Endpoint) {
	endpointsMu.Lock()
	Endpoints = cloneEndpoints(next)
	endpointsMu.Unlock()
}

func cloneEndpoints(in []Endpoint) []Endpoint {
	if len(in) == 0 {
		return nil
	}
	out := make([]Endpoint, len(in))
	copy(out, in)
	return out
}
