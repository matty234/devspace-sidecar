package config

type KubernetesConfig struct {
	UseClusterConfig   bool   `yaml:"useClusterConfig" mapstructure:"useClusterConfig" json:"useClusterConfig"`
	KubeConfigLocation string `yaml:"kubeConfigLocation" mapstructure:"kubeConfigLocation" json:"kubeConfigLocation"`
	Namespace          string `yaml:"namespace" mapstructure:"namespace" json:"namespace"`
}
