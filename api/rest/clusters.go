package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/labels"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// InClusterName is a reserved name for the in-cluster configuration.
// This is intentionally an invalid DNS name, as this cannot be the same name as one from a secret (which must be a valid DNS name)
const InClusterName = "#in-cluster"

type ConfigOpt func(config)

func WithHost(v string) ConfigOpt {
	return func(c config) {
		if v != "" {
			c.Host = v
		}
	}
}

func WithImpersonate(v rest.ImpersonationConfig) ConfigOpt {
	return func(c config) { c.Impersonate = v }
}

func AddConfig(ctx context.Context, clusterName string, config rest.Config, authInfo api.AuthInfo, secretsInterface typedcorev1.SecretInterface, opts ...ConfigOpt) error {
	c := newConfig(config, authInfo)
	for _, opt := range opts {
		opt(c)
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName(clusterName),
			Labels: map[string]string{
				labels.KeyClusterName: clusterName,
			},
		},
		Data: map[string][]byte{
			labels.KeyRestConfig: data,
		},
	}
	_, err = secretsInterface.Create(ctx, secret, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		data, err := json.Marshal(secret)
		if err != nil {
			return err
		}
		_, err = secretsInterface.Patch(ctx, secret.Name, types.MergePatchType, data, metav1.PatchOptions{})
		return err
	} else {
		return err
	}
}

func RemoveConfig(ctx context.Context, clusterName string, secretsInterface typedcorev1.SecretInterface) error {
	err := secretsInterface.Delete(ctx, secretName(clusterName), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

func secretName(clusterName string) string {
	return fmt.Sprintf("cluster-%s", clusterName)
}

type Configs map[string]*rest.Config

type NewConfigsOpt func(Configs)

func WithInClusterConfig(r *rest.Config) NewConfigsOpt {
	return func(configs Configs) {
		configs[InClusterName] = r
	}
}

func NewConfigs(ctx context.Context, secretsInterface typedcorev1.SecretInterface, opts ...NewConfigsOpt) (Configs, error) {
	configs := make(Configs)
	list, err := secretsInterface.List(ctx, metav1.ListOptions{LabelSelector: labels.KeyClusterName})
	if err != nil {
		return nil, err
	}
	for _, secret := range list.Items {
		c := &config{}
		data, ok := secret.Data[labels.KeyRestConfig]
		if !ok {
			return nil, fmt.Errorf("key %q not found in secret %s/%s", labels.KeyRestConfig, secret.Namespace, secret.Name)
		}
		if err := json.Unmarshal(data, c); err != nil {
			return nil, err
		}
		configs[secret.Labels[labels.KeyClusterName]] = c.restConfig()
	}
	for _, opt := range opts {
		opt(configs)
	}
	return configs, nil
}

func NewConfig(ctx context.Context, secretsInterface typedcorev1.SecretInterface, clusterName string, opts ...ConfigOpt) (*rest.Config, error) {
	secret, err := secretsInterface.Get(ctx, secretName(clusterName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	c := config{}
	data, ok := secret.Data[labels.KeyRestConfig]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret %s/%s", labels.KeyRestConfig, secret.Namespace, secret.Name)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.restConfig(), nil
}
