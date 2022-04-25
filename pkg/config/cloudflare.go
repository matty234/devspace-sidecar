package config

type CloudflareConfig struct {
	Email string `yaml:"email" mapstructure:"email" json:"email" validate:"required"`
}
