package config

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/argoproj-labs/multi-cluster-kubernetes/api/labels"
)

var resourceName = func(name string) string {
	return fmt.Sprintf("%s-kubeconfig", name)
}

const secretKey = "value"

type Client struct {
	secretsInterface typedcorev1.SecretInterface
}

func New(secretsInterface typedcorev1.SecretInterface) Client {
	return Client{secretsInterface}
}

func (g Client) Add(ctx context.Context, name string, value *clientcmdapi.Config) error {
	resourceName := resourceName(name)
	secret, err := g.secretsInterface.Get(ctx, resourceName, metav1.GetOptions{})
	notFound := errors.IsNotFound(err)
	if notFound {
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   resourceName,
				Labels: map[string]string{labels.KeyKubeConfig: ""},
			},
			Data: map[string][]byte{},
		}
	} else if err != nil {
		return fmt.Errorf("failed to find kubeconfig secret: %w", err)
	}
	oldValue, err := clientcmd.Load(secret.Data[secretKey])
	if err != nil {
		return fmt.Errorf("failed to unmarshal old value: %w", err)
	}
	for k, v := range value.Clusters {
		oldValue.Clusters[k] = v
	}
	for k, v := range value.AuthInfos {
		oldValue.AuthInfos[k] = v
	}
	for k, v := range value.Contexts {
		oldValue.Contexts[k] = v
	}
	secret.Data[secretKey], err = clientcmd.Write(*oldValue)
	if err != nil {
		return fmt.Errorf("failed to marshal modifed value: %w", err)
	}
	if notFound {
		_, err = g.secretsInterface.Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create kubeconfig secret: %w", err)
		}
	} else {
		_, err = g.secretsInterface.Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update kubeconfig secret: %w", err)
		}
	}
	return nil
}

func (g Client) Get(ctx context.Context, name string) (*clientcmdapi.Config, error) {
	resourceName := resourceName(name)
	secret, err := g.secretsInterface.Get(ctx, resourceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return clientcmdapi.NewConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig secret: %w", err)
	}
	data, ok := secret.Data[secretKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret %s/%s", secretKey, secret.Namespace, secret.Name)
	}
	v, err := clientcmd.Load(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall data: %w", err)
	}
	return v, nil
}

func NewClientConfigs(config clientcmdapi.Config) map[string]clientcmd.ClientConfig {
	configs := make(map[string]clientcmd.ClientConfig)
	for contextName := range config.Contexts {
		configs[contextName] = clientcmd.NewNonInteractiveClientConfig(config, contextName, &clientcmd.ConfigOverrides{}, clientcmd.NewDefaultClientConfigLoadingRules())
	}
	return configs
}

func NewRestConfigs(config map[string]clientcmd.ClientConfig) (map[string]*rest.Config, error) {
	configs := make(map[string]*rest.Config)
	for contextName, r := range config {
		clientConfig, err := r.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create client config from rest config: %w", err)
		}
		configs[contextName] = clientConfig
	}
	return configs, nil
}
