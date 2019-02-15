package vpn

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error
	Wait(chan error)

	GenerateClientKey(string) ([]byte, error)
	DisconnectClient(string) error
	ClientList() ([]Client, error)
}

type Client struct {
	ID       string `json:"id"`
	Upload   int    `json:"upload"`
	Download int    `json:"download"`
}

func NewClient(id string, upload, download int) Client {
	return Client{
		ID:       id,
		Upload:   upload,
		Download: download,
	}
}
