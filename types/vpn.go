package types

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error
	Wait(chan error)

	GenerateClientKey(id string) ([]byte, error)
	DisconnectClient(id string) error
	ClientList() ([]VPNClient, error)
}

type VPNClient struct {
	ID       string `json:"id"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func NewVPNClient(id string, upload, download int64) VPNClient {
	return VPNClient{
		ID:       id,
		Upload:   upload,
		Download: download,
	}
}
