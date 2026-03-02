package models

import (
	"encoding/xml"
)

// CEPAddress is the normalized address response used by this project.
type CEPAddress struct {
	City         string `json:"cidade"`
	StateCode    string `json:"uf"`
	Street       string `json:"logradouro"`
	Neighborhood string `json:"bairro"`
}

type CdnApiCep struct {
	Status   int    `json:"status"`
	Code     string `json:"code"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

type GithubJeffotoni struct {
	PostalCode   string `json:"cep"`
	Street       string `json:"logradouro"`
	Neighborhood string `json:"bairro"`
	StateCode    string `json:"uf"`
	State        string `json:"estado"`
	City         string `json:"cidade"`
	IBGE         int    `json:"ibge"`
}

// ViaCep provider response.
type ViaCep struct {
	PostalCode   string `json:"cep"`
	Street       string `json:"logradouro"`
	Complement   string `json:"complemento"`
	Neighborhood string `json:"bairro"`
	City         string `json:"localidade"`
	StateCode    string `json:"uf"`
	Unit         string `json:"unidade"`
	IBGE         string `json:"ibge"`
	GIA          string `json:"gia"`
}

// Postmon provider response.
type PostMon struct {
	Neighborhood string `json:"bairro"`
	City         string `json:"cidade"`
	Street       string `json:"logradouro"`
	StateInfo    struct {
		AreaKM2  string `json:"area_km2"`
		IBGECode string `json:"codigo_ibge"`
		Name     string `json:"nome"`
	} `json:"estado_info"`
	PostalCode string `json:"cep"`
	CityInfo   struct {
		AreaKM2  string `json:"area_km2"`
		IBGECode string `json:"codigo_ibge"`
	} `json:"cidade_info"`
	State string `json:"estado"`
}

// RepublicaVirtual provider response.
type RepublicaVirtual struct {
	Result       string `json:"resultado"`
	ResultText   string `json:"resultado_txt"`
	StateCode    string `json:"uf"`
	City         string `json:"cidade"`
	Neighborhood string `json:"bairro"`
	StreetType   string `json:"tipo_logradouro"`
	Street       string `json:"logradouro"`
}

type Correio struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	Soap    string   `xml:"soap,attr"`
	Body    struct {
		Text              string `xml:",chardata"`
		LookupCEPResponse struct {
			Text   string `xml:",chardata"`
			Ns2    string `xml:"ns2,attr"`
			Return struct {
				Text         string `xml:",chardata"`
				Neighborhood string `xml:"bairro"`
				PostalCode   string `xml:"cep"`
				City         string `xml:"cidade"`
				Complement2  string `xml:"complemento2"`
				Address      string `xml:"end"`
				StateCode    string `xml:"uf"`
			} `xml:"return"`
		} `xml:"consultaCEPResponse"`
	} `xml:"Body"`
}

type BrasilAPI struct {
	PostalCode   string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

// OpenCEP provider response.
type OpenCEP struct {
	PostalCode   string `json:"cep"`
	Street       string `json:"logradouro"`
	Complement   string `json:"complemento"`
	Neighborhood string `json:"bairro"`
	City         string `json:"localidade"`
	StateCode    string `json:"uf"`
	State        string `json:"estado"`
	IBGE         string `json:"ibge"`
}

// AwesomeAPI provider response.
type AwesomeAPI struct {
	PostalCode   string `json:"cep"`
	AddressType  string `json:"address_type"`
	AddressName  string `json:"address_name"`
	Street       string `json:"address"`
	StateCode    string `json:"state"`
	Neighborhood string `json:"district"`
	City         string `json:"city"`
	CityIBGE     string `json:"city_ibge"`
	DDD          string `json:"ddd"`
}
