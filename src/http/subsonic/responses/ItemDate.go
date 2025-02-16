package responses

type ItemDate struct {
	Year  int `json:"year" xml:"year,attr"`
	Month int `json:"month" xml:"month,attr"`
	Day   int `json:"day" xml:"day,attr"`
}

func NewItemDate(year int, month int, day int) *ItemDate {
	return &ItemDate{
		Year:  year,
		Month: month,
		Day:   day,
	}
}
