package responses

type InternetRadioStations struct {
	InternetRadioStation []InternetRadioStation `json:"internetRadioStation" xml:"internetRadioStation"`
}

func NewInternetRadioStations(internetRadioStations []InternetRadioStation) *InternetRadioStations {
	return &InternetRadioStations{
		InternetRadioStation: internetRadioStations,
	}
}
