package cmd

import (
	"context"
	"log"

	"github.com/go-playground/validator"
	"github.com/matty234/dev-space-configure/pkg/cloudflare"
	"github.com/matty234/dev-space-configure/pkg/config"
	"github.com/matty234/dev-space-configure/pkg/credentialsprovider"
	"github.com/matty234/dev-space-configure/pkg/servicemeta"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// envoyKeySetupCmd represents the envoyKeySetup command
var envoyKeySetupCmd = &cobra.Command{
	Use:   "create-keys",
	Short: `Generate keys using Hashicorp CA for Envoy to use`,
	Run: func(cmd *cobra.Command, args []string) {

		var cfg config.Config

		err := viper.Unmarshal(&cfg)
		if err != nil {
			log.Fatalf("unable to decode config into struct, %v", err)
		}

		validate := validator.New()
		err = validate.Struct(cfg)
		if err != nil {
			log.Fatalf("config validation failed, %v", err)
		}

		go func() {
			ctx := context.Background()
			bringup(ctx, cfg)
		}()
		<-cmd.Context().Done()
		teardown(context.Background(), cfg)

	},
}

func init() {
	rootCmd.AddCommand(envoyKeySetupCmd)
}

func bringup(ctx context.Context, cfg config.Config) {
	log.Print("[INFO] Starting up...")
	v, err := credentialsprovider.NewVaultCredentialsProvider(cfg.VaultConfiguration)
	if err != nil {
		log.Fatalf("Error getting credentials: %v", err)
	}

	log.Print("[INFO] Getting cloudflare credentials...")
	cloudflaretoken, err := v.GetCloudflareCredentials(cfg.Domains.RootDomain)
	if err != nil {
		log.Fatalf("Error getting credentials: %v", err)
	}

	log.Print("[INFO] Creating tls credentials...")
	certs, err := v.GetTLSCredentials(cfg.Domains.RootDomain, cfg.Domains.Subdomain)
	if err != nil {
		log.Fatalf("Error getting credentials: %v", err)
	}

	certificateOutputDir := "/tmp/certs"
	err = certs.WriteToFileSystem(certificateOutputDir)
	if err != nil {
		log.Fatalf("could not write certificates to %s: %v", certificateOutputDir, err)
	}

	cf, err := cloudflare.NewCloudflareDNSProvider(cfg.CloudflareConfig, cloudflaretoken)
	if err != nil {
		log.Fatalf("could not create cloudflare provider: %v", err)
	}

	svclb, err := servicemeta.NewKubernetesServiceMetaProvider(cfg.KubernetesConfig)
	if err != nil {
		log.Fatalf("could not create service meta provider: %v", err)
	}

	lb, err := svclb.GetServiceMeta(ctx, cfg.Domains.Subdomain)
	if err != nil {
		log.Fatalf("could not get service meta: %v", err)
	}

	log.Printf("[INFO] Creating DNS record for %s", lb.LoadBalancerHost)
	err = cf.CreateDNSRecord(ctx, cfg.Domains.RootDomain, cfg.Domains.Subdomain, "example.com")
	if err != nil {
		log.Fatalf("could not create DNS record: %v", err)
	}

	log.Print("[INFO] Setup complete!")
}

func teardown(ctx context.Context, cfg config.Config) {

	v, err := credentialsprovider.NewVaultCredentialsProvider(cfg.VaultConfiguration)
	if err != nil {
		panic(err)
	}

	cloudflaretoken, err := v.GetCloudflareCredentials(cfg.Domains.RootDomain)
	if err != nil {
		log.Fatalf("Error getting credentials: %v", err)
	}

	cf, err := cloudflare.NewCloudflareDNSProvider(cfg.CloudflareConfig, cloudflaretoken)
	if err != nil {
		log.Fatalf("could not create cloudflare provider: %v", err)
	}

	err = cf.DeleteDNSRecord(ctx, cfg.Domains.RootDomain, cfg.Domains.Subdomain)
	if err != nil {
		log.Fatalf("could not delete DNS record: %v", err)
	}

	log.Print("[INFO] Shutdown complete!")
}
