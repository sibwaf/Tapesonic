package responses

import "time"

type License struct {
	Valid          bool       `json:"valid" xml:"valid,attr"`
	Email          string     `json:"email,omitempty" xml:"email,attr,omitempty"`
	LicenseExpires *time.Time `json:"licenseExpires" xml:"licenseExpires,attr,omitempty"`
	TrialExpires   *time.Time `json:"trialExpires" xml:"trialExpires,attr,omitempty"`
}

func NewLicense(valid bool) *License {
	return &License{Valid: valid}
}
