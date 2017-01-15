package constants

// Database settings
type Database struct {
	Host             string
	ManagementSystem string `toml:"management_system"`
	Name             string
	User             string
	Password         string
	Offset           int
	IsContainer      bool `toml:"is_container"`
}

// SSH settings
type SSH struct {
	Host string
	Port string
	User string
	Key  string
}
