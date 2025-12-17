package server

type Config struct {
	Pass          string
	StoreFileName string
}

func NewConfig(pass string, storeFileName string) Config {
	return Config{
		Pass:          pass,
		StoreFileName: storeFileName,
	}
}
