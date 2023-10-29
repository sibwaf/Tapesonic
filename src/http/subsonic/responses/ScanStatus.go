package responses

type ScanStatus struct {
	Scanning bool `json:"scanning" xml:"scanning,attr"`
	Count    int  `json:"count" xml:"count,attr"`
}

func NewScanStatus(
	scanning bool,
	count int,
) *ScanStatus {
	return &ScanStatus{
		Scanning: scanning,
		Count:    count,
	}
}
