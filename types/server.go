package types

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Info    interface{} `json:"info,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   Error       `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}
