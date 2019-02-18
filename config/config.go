package config

type Config interface {
	LoadFromPath(string) error
	SaveToPath(string) error
}
