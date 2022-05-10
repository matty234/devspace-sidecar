package config

import (
	"context"
	"os"

	vault "github.com/hashicorp/vault/api"
	kubeauth "github.com/hashicorp/vault/api/auth/kubernetes"
)

const (
	serviceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type VaultConfiguration struct {
	Address   string `yaml:"address" mapstructure:"address" json:"address" validate:"required"`
	Namespace string `yaml:"namespace" mapstructure:"namespace" json:"namespace"`
	Token     string `yaml:"token" mapstructure:"token" json:"token"`

	UseKubernetes  bool   `yaml:"use_kubernetes" mapstructure:"use_kubernetes" json:"use_kubernetes"`
	KubernetesRole string `yaml:"kubernetes_role" mapstructure:"kubernetes_role" json:"kubernetes_role"`
	Vaults         struct {
		KeyValueVault string `yaml:"kv" mapstructure:"kv" json:"kv" validate:"required"`
		PkiVault      string `yaml:"pki" mapstructure:"pki" json:"pki" validate:"required"`
	} `yaml:"vaults" mapstructure:"vaults" json:"vaults" validate:"required"`
}

func (v *VaultConfiguration) GetClient() (*vault.Client, error) {
	config := vault.DefaultConfig()

	config.Address = v.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	if v.Namespace != "" {
		client.SetNamespace(v.Namespace)
	}

	if v.Token != "" && !v.UseKubernetes {
		client.SetToken(v.Token)
		return client, nil
	}

	if v.UseKubernetes {
		if _, err := os.Stat(serviceAccountFile); os.IsNotExist(err) {
			return nil, err
		}

		auth, err := kubeauth.NewKubernetesAuth(v.KubernetesRole)
		if err != nil {
			return nil, err
		}

		sec, err := client.Auth().Login(context.Background(), auth)
		if err != nil {
			return nil, err
		}

		_, err = client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
			Secret: sec,
		})
		if err != nil {
			return nil, err
		}

		client.SetToken(sec.Auth.ClientToken)
		return client, nil

	}

	return client, nil
}
