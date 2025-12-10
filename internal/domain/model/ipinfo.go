package model

// IPInfo IP地理位置信息
type IPInfo struct {
	IP              string  `json:"ip"`
	Country         string  `json:"country"`
	CountryCode     string  `json:"country_code"`
	ASN             int     `json:"asn"`
	ASNOrganization string  `json:"asn_organization"`
	Organization    string  `json:"organization"`
	ISP             string  `json:"isp"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Timezone        string  `json:"timezone"`
	ContinentCode   string  `json:"continent_code"`
	Offset          int     `json:"offset"`
}
