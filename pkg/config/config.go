package config

type Config struct {
	VaultConfiguration *VaultConfiguration `yaml:"vault" mapstructure:"vault" json:"vault" validate:"required"`
	CloudflareConfig   *CloudflareConfig   `yaml:"cloudflare" mapstructure:"cloudflare" json:"cloudflare" validate:"required"`

	Domains *DomainsConfig `yaml:"domains" mapstructure:"domains" json:"domains" validate:"required"`

	KubernetesConfig *KubernetesConfig `yaml:"kubernetes" mapstructure:"kubernetes" json:"kubernetes" validate:"required"`
}
