package cloudflare

// import cloudflare api
import (
	"context"
	"log"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
	"github.com/matty234/dev-space-configure/pkg/config"
)

type DNSProvider interface {
	CreateDNSRecord(ctx context.Context, roothost, host, loadbalancer string) error
	DeleteDNSRecord(ctx context.Context, roothost, host string) error
}

type CloudflareDNSProvider struct {
	client *cf.API
	config *config.CloudflareConfig
}

func NewCloudflareDNSProvider(config *config.CloudflareConfig, token string) (DNSProvider, error) {
	cf, err := cf.New(token, config.Email)
	if err != nil {
		return nil, err
	}

	return &CloudflareDNSProvider{
		client: cf,
		config: config,
	}, nil
}

func (c *CloudflareDNSProvider) CreateDNSRecord(ctx context.Context, roothost, host, loadbalancer string) error {
	zoneID, err := c.client.ZoneIDByName(roothost)
	if err != nil {
		return err
	}

	falsey := false

	_, err = c.client.CreateDNSRecord(ctx, zoneID, cf.DNSRecord{
		Name:    host,
		Type:    "CNAME",
		Content: loadbalancer,
		Proxied: &falsey,
	})

	_, err = c.client.CreateDNSRecord(ctx, zoneID, cf.DNSRecord{
		Name:    "*." + host,
		Type:    "CNAME",
		Content: loadbalancer,
		Proxied: &falsey,
	})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}

	return nil
}

func (c *CloudflareDNSProvider) DeleteDNSRecord(ctx context.Context, roothost, host string) error {
	if len(host) < 2 {
		log.Printf("hostname is too short, skipping delete")
		return nil
	}

	zoneID, err := c.client.ZoneIDByName(roothost)
	if err != nil {
		return err
	}

	dns, err := c.client.DNSRecords(ctx, zoneID, cf.DNSRecord{
		Name: host + "." + roothost,

		Type: "CNAME",
	})

	if err != nil {
		return err
	}

	for _, record := range dns {
		err := c.client.DeleteDNSRecord(ctx, zoneID, record.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
