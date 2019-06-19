package types

type StdError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Info    interface{} `json:"info,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}
