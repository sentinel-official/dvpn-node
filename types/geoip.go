package types

type GeoIPLocation struct {
	City      string  `json:"city"`
	Country   string  `json:"country"`
	IP        string  `json:"ip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
