package responses

type InternetRadioStation struct {
	Id          string `json:"id" xml:"id,attr"`
	Name        string `json:"name" xml:"name,attr"`
	StreamUrl   string `json:"streamUrl" xml:"streamUrl,attr"`
	HomePageUrl string `json:"homePageUrl,omitempty" xml:"homePageUrl,attr,omitempty"`
}

func NewInternetRadioStation(
	id string,
	name string,
	streamUrl string,
) *InternetRadioStation {
	return &InternetRadioStation{
		Id:        id,
		Name:      name,
		StreamUrl: streamUrl,
	}
}
