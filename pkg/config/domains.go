package config

type DomainsConfig struct {
	RootDomain string `json:"root" mapstructure:"root" yaml:"root" validate:"required"`
	Subdomain  string `json:"sub" mapstructure:"sub" yaml:"sub" validate:"required"`
}
