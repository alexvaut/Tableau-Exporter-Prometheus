package config

// Configurations exported
type Configurations struct {
	Database              DatabaseConfigurations
	ScrapeIntervalSeconds int64
	Port                  int
}

// DatabaseConfigurations exported
type DatabaseConfigurations struct {
	Host     string //`mapstructure:"HOST"`
	Port     int
	Name     string
	User     string
	Password string
}
