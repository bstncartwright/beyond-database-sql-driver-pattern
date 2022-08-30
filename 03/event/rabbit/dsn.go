package rabbit

type config struct {
	Scheme   string
	Username string
	Password string
	Host     string
	Port     string
	Name     string
}

func NewConfig() config {
	return config{
		Scheme:   "amqp",
		Username: "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     "5672",
		Name:     "demo",
	}
}
