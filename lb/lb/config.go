package lb

func DefaultConfig() *Config {
	return &Config{
		Bind: "0.0.0.0",
		Port: 9888,
	}
}

type Config struct {
	Bind string
	Port int
	Store map[string]interface{}
}
