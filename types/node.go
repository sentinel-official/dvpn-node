package types

type Node struct {
	ID      string `json:"id" `
	Address string `json:"address" `
	Deposit Coin   `json:"deposit"`
	
	IP   string `json:"ip"`
	Port string `json:"port" `
	
	Type          string    `json:"type"`
	Version       string    `json:"version"`
	Moniker       string    `json:"moniker" `
	PricesPerGB   []Coin    `json:"prices_per_gb" `
	InternetSpeed Bandwidth `json:"internet_speed" `
	Encryption    string    `json:"encryption"`
	
	Status string `json:"status" bson:"status"`
}

type Coin struct {
	Denom string `json:"denom,omitempty" bson:"denom,omitempty"`
	Value int64  `json:"value,omitempty" bson:"value,omitempty"`
}

type Signature struct {
	PubKey    string `json:"pub_key,omitempty" bson:"pub_key,omitempty"`
	Signature string `json:"signature,omitempty" bson:"signature,omitempty"`
}

type Bandwidth struct {
	Upload   int64 `json:"upload" bson:"upload"`
	Download int64 `json:"download" bson:"download"`
}
