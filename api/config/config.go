package config

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
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// InClusterName is a reserved name for the in-cluster configuration.
// This is intentionally an invalid DNS name, as this cannot be the same name as one from a secret (which must be a valid DNS name)
const InClusterName = "#in-cluster"
const configKey = "config"

// WithServer allows you to override the cluster server, useful for when the host name is different from within the cluster.
func WithServer(v string) interface{} {
	return func(c *clientcmdapi.Config) {
		for _, c := range c.Clusters {
			c.Server = v
		}
	}
}

func WithNamespace(v string) interface{} {
	return func(c *clientcmdapi.Config) {
		for _, c := range c.Contexts {
			c.Namespace = v
		}
	}
}

func WithImpersonate(v string) interface{} {
	return func(c *clientcmdapi.Config) {
		for _, c := range c.AuthInfos {
			c.Impersonate = v
		}
	}
}

// WithLabel labels the config so you can query it later on.
func WithLabel(k, v string) interface{} {
	return func(m metav1.ObjectMeta) { m.GetLabels()[k] = v }
}

func AddConfig(ctx context.Context, configName string, config *clientcmdapi.Config, contextName string, secretsInterface typedcorev1.SecretInterface, opts ...interface{}) (*clientcmdapi.Config, error) {
	kubeContext := config.Contexts[contextName]
	clusterName := kubeContext.Cluster
	authInfoName := kubeContext.AuthInfo
	c := &clientcmdapi.Config{
		Kind:           config.Kind,
		APIVersion:     config.APIVersion,
		Clusters:       map[string]*clientcmdapi.Cluster{clusterName: config.Clusters[clusterName]},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{authInfoName: config.AuthInfos[authInfoName]},
		Contexts:       map[string]*clientcmdapi.Context{contextName: kubeContext},
		CurrentContext: contextName,
	}
	m := metav1.ObjectMeta{
		Name:   secretName(configName),
		Labels: map[string]string{labels.KeyConfigName: configName},
	}
	for _, opt := range opts {
		switch x := opt.(type) {
		case func(*clientcmdapi.Config):
			x(c)
		case func(metav1.ObjectMeta):
			x(m)
		default:
			return nil, fmt.Errorf("unsupported option type %T", opt)
		}
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	secret := &corev1.Secret{
		ObjectMeta: m,
		Data:       map[string][]byte{configKey: data},
	}
	_, err = secretsInterface.Create(ctx, secret, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		data, err = json.Marshal(secret)
		if err != nil {
			return nil, err
		}
		_, err = secretsInterface.Patch(ctx, secret.Name, types.MergePatchType, data, metav1.PatchOptions{})
	}
	return c, err
}

func RemoveConfig(ctx context.Context, configName string, secretsInterface typedcorev1.SecretInterface) error {
	err := secretsInterface.Delete(ctx, secretName(configName), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

func secretName(configName string) string {
	return fmt.Sprintf("config-%s", configName)
}

type Configs map[string]clientcmd.ClientConfig

// WithInCluster adds the provide config under InClusterName
func WithInCluster(r clientcmd.ClientConfig) func(configs Configs) {
	return func(configs Configs) {
		configs[InClusterName] = r
	}
}

// WithLabelSelector allows you to select only certain configs.
func WithLabelSelector(v string) func(opts metav1.ListOptions) {
	return func(opts metav1.ListOptions) {
		opts.LabelSelector += v
	}
}

func NewConfigs(ctx context.Context, secretsInterface typedcorev1.SecretInterface, opts ...interface{}) (Configs, error) {

	options := metav1.ListOptions{LabelSelector: labels.KeyConfigName}

	for _, opt := range opts {
		switch x := opt.(type) {
		case func(*clientcmdapi.Config), func(Configs):
			// noop
		case func(metav1.ListOptions):
			x(options)
		default:
			return nil, fmt.Errorf("unsupported option type %T", opt)
		}
	}
	list, err := secretsInterface.List(ctx, options)
	if err != nil {
		return nil, err
	}
	configs := make(Configs)
	for _, secret := range list.Items {
		c := &clientcmdapi.Config{}
		data, ok := secret.Data[configKey]
		if !ok {
			return nil, fmt.Errorf("key %q not found in secret %s/%s", configKey, secret.Namespace, secret.Name)
		}
		if err := json.Unmarshal(data, c); err != nil {
			return nil, err
		}
		for _, opt := range opts {
			switch x := opt.(type) {
			case func(c *clientcmdapi.Config):
				x(c)
			}
		}
		configs[secret.Labels[labels.KeyConfigName]] = clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{})
	}
	for _, opt := range opts {
		switch x := opt.(type) {
		case func(Configs):
			x(configs)
		}
	}
	return configs, nil
}

func NewConfig(ctx context.Context, secretsInterface typedcorev1.SecretInterface, clusterName string, opts ...interface{}) (*clientcmdapi.Config, error) {
	secret, err := secretsInterface.Get(ctx, secretName(clusterName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	c := &clientcmdapi.Config{}
	data, ok := secret.Data[configKey]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret %s/%s", configKey, secret.Namespace, secret.Name)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		switch x := opt.(type) {
		case func(*clientcmdapi.Config):
			x(c)
		default:
			return nil, fmt.Errorf("unsupported option type %T", opt)
		}
	}
	return c, nil
}
