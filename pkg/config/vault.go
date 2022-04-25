package config

type VaultConfiguration struct {
	Address   string `yaml:"address" mapstructure:"address" json:"address" validate:"required"`
	Namespace string `yaml:"namespace" mapstructure:"namespace" json:"namespace"`
	Token     string `yaml:"token" mapstructure:"token" json:"token"`

	Vaults struct {
		KeyValueVault string `yaml:"kv" mapstructure:"kv" json:"kv" validate:"required"`
		PkiVault      string `yaml:"pki" mapstructure:"pki" json:"pki" validate:"required"`
	} `yaml:"vaults" mapstructure:"vaults" json:"vaults" validate:"required"`
}
