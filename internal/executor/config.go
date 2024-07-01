package executor

type Config struct {
	Version string
	Commit  string
	Date    string
}

var AppConfig Config

func SetConfig(config Config) {
	AppConfig = config
}
