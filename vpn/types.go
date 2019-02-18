package vpn

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error
	Wait(chan error)

	GenerateClientKey(string) (interface{}, error)
	DisconnectClient(string) error
	ClientList() ([]Client, error)
}

type Client struct {
	ID       string `json:"id"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func NewClient(id string, upload, download int64) Client {
	return Client{
		ID:       id,
		Upload:   upload,
		Download: download,
	}
}
