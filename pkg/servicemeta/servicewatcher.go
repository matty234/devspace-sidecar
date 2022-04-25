package servicemeta

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/matty234/dev-space-configure/pkg/config"
	apicorev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ServiceMeta struct {
	LoadBalancerHost string
}

type ServiceMetaProvider interface {
	GetServiceMeta(ctx context.Context, forHost string) (*ServiceMeta, error)
}

type KubernetesServiceMetaProvider struct {
	coreapi   corev1.CoreV1Interface
	namespace string
}

func NewKubernetesServiceMetaProvider(cfg config.KubernetesConfig) (ServiceMetaProvider, error) {
	var k8scfg *rest.Config

	if cfg.UseClusterConfig {
		var err error
		k8scfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {

		var kc *string

		if cfg.KubeConfigLocation == "" {
			if home := homedir.HomeDir(); home != "" {
				kc = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			} else {
				kc = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			}
		} else {
			kc = &cfg.KubeConfigLocation
		}
		if kc == nil || *kc == "" {
			return nil, fmt.Errorf("kubeconfig location not supported")
		}
		var err error
		// use the current context in kubeconfig
		k8scfg, err = clientcmd.BuildConfigFromFlags("", *kc)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		return nil, err
	}

	return &KubernetesServiceMetaProvider{
		coreapi: clientset.CoreV1(),
	}, nil
}

func (k *KubernetesServiceMetaProvider) GetServiceMeta(ctx context.Context, forHost string) (*ServiceMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*3)
	defer cancel()
	svcapi := k.coreapi.Services(k.namespace)

	foundServices, err := svcapi.List(ctx, v1.ListOptions{
		LabelSelector: fmt.Sprintf("for-devspace=%s", forHost),
	})

	if err != nil {
		return nil, err
	}

	if len(foundServices.Items) == 0 {
		return nil, fmt.Errorf(`no service found for host %s, make sure the following label is applied: for-devspace: %s`, forHost, forHost)
	}

	if len(foundServices.Items) > 1 {
		return nil, fmt.Errorf(`multiple services found for host %s, make sure only one exists`, forHost)
	}

	foundService := foundServices.Items[0]

	if foundService.Spec.Type != "LoadBalancer" {
		return nil, fmt.Errorf(`service %s is not a load balancer`, foundService.Name)
	}

	if !doesServiceHaveALoadbalancerHostname(&foundService) {
		// wait for the service to be ready
		err = waitForService(ctx, svcapi, foundService.Name)
		if err != nil {
			return nil, err
		}
	}

	if !doesServiceHaveALoadbalancerHostname(&foundService) {
		return nil, fmt.Errorf(`service %s does not have a load balancer hostname`, foundService.Name)
	}

	return &ServiceMeta{
		LoadBalancerHost: foundService.Status.LoadBalancer.Ingress[0].Hostname,
	}, nil
}

// waitForService waits for the service to be ready
func waitForService(ctx context.Context, svcapi corev1.ServiceInterface, serviceName string) error {
	for {
		svc, err := svcapi.Get(ctx, serviceName, v1.GetOptions{})
		if err != nil {
			return err
		}

		if doesServiceHaveALoadbalancerHostname(svc) {
			return nil
		}

		// wait for the service to be ready
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service %s to be ready", serviceName)
		case <-time.After(time.Second):
		}
	}
}

func doesServiceHaveALoadbalancerHostname(svc *apicorev1.Service) bool {
	return svc.Status.LoadBalancer.Ingress != nil && len(svc.Status.LoadBalancer.Ingress) > 0 && svc.Status.LoadBalancer.Ingress[0].Hostname != ""
}
