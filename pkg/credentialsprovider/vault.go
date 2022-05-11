package credentialsprovider

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/matty234/dev-space-configure/pkg/config"
)

type CredentialsProviders interface {
	GetTLSCredentials() (*Credentials, error)

	GetCloudflareCredentials() (string, error)
}

type VaultCredentialsProvider struct {
	vaultClient *vault.Client

	KeyValueVault string
	PkiVault      string
}

func NewVaultCredentialsProvider(vaultConfig *config.VaultConfiguration) (*VaultCredentialsProvider, error) {
	vaultClient, err := vaultConfig.GetClient()
	if err != nil {
		return nil, err
	}

	return &VaultCredentialsProvider{
		vaultClient:   vaultClient,
		KeyValueVault: vaultConfig.Vaults.KeyValueVault,
		PkiVault:      vaultConfig.Vaults.PkiVault,
	}, nil

}

func (vcp *VaultCredentialsProvider) GetCloudflareCredentials(roothost string) (string, error) {
	// check if the roothost contains slashes
	if roothost == "" {
		return "", fmt.Errorf("roothost is empty")
	}

	if roothost[len(roothost)-1] == '/' {
		roothost = roothost[:len(roothost)-1]
	}

	for i := len(roothost) - 1; i >= 0; i-- {
		if roothost[i] == '/' {
			return "", fmt.Errorf("roothost contains slashes")
		}
	}

	path := fmt.Sprintf("%s/data/cloudflare", vcp.KeyValueVault)
	logicalvault := vcp.vaultClient.Logical()
	credentials, err := logicalvault.Read(path)
	if err != nil {
		return "", err
	}

	if credentials == nil {
		return "", fmt.Errorf("credentials not found")
	}

	if credentials.Data == nil {
		return "", fmt.Errorf("credentials not found")
	}

	cloudflareCredentials, ok := credentials.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("credentials not a map")
	}

	cloudflareKey, ok := cloudflareCredentials[roothost].(string)
	if !ok {
		return "", fmt.Errorf("credentials not a string")
	}

	return cloudflareKey, nil
}

func (vcp *VaultCredentialsProvider) GetTLSCredentials(roothost, host string) (*Credentials, error) {
	logicalvault := vcp.vaultClient.Logical()
	secret, err := logicalvault.Write(fmt.Sprintf("%s/issue/%s", vcp.PkiVault, roothost), map[string]interface{}{
		"common_name":          fmt.Sprint(host, ".", roothost),
		"ttl":                  "2592000s",
		"exclude_cn_from_sans": false,
	})

	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, fmt.Errorf("secret is nil")
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("secret data is nil")
	}

	certificate, ok := secret.Data["certificate"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find certificate in secret")
	}

	issuingCa, ok := secret.Data["issuing_ca"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find issuing_ca in secret")
	}

	privateKey, ok := secret.Data["private_key"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find private_key in secret")
	}

	return &Credentials{
		certificate: []byte(certificate),
		ca:          []byte(issuingCa),
		privatekey:  []byte(privateKey),
	}, nil
}
