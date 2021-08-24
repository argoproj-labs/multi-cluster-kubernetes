package api

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

func AddCluster(ctx context.Context, clusterName string, config rest.Config, authInfo api.AuthInfo, secretsInterface typedcorev1.SecretInterface) error {
	data, err := json.Marshal(NewConfig(config, authInfo))
	if err != nil {
		return err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name(clusterName),
			Labels: map[string]string{
				KeyClusterName: clusterName,
			},
		},
		Data: map[string][]byte{
			KeyRestConfig: data,
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

func name(clusterName string) string {
	return fmt.Sprintf("cluster-%s", clusterName)
}

type RestConfigs map[string]*rest.Config

func LoadClusters(ctx context.Context, secretsInterface typedcorev1.SecretInterface) (RestConfigs, error) {
	configs := make(RestConfigs)
	list, err := secretsInterface.List(ctx, metav1.ListOptions{LabelSelector: KeyClusterName})
	if err != nil {
		return nil, err
	}
	for _, secret := range list.Items {
		c := &config{}
		if err := json.Unmarshal(secret.Data[KeyRestConfig], c); err != nil {
			return nil, err
		}
		configs[secret.Labels[KeyClusterName]] = c.restConfig()
	}
	return configs, nil
}

func LoadCluster(ctx context.Context, secretsInterface typedcorev1.SecretInterface, clusterName string) (*rest.Config, error) {
	secret, err := secretsInterface.Get(ctx, name(clusterName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	c := &config{}
	if err := json.Unmarshal(secret.Data[KeyRestConfig], c); err != nil {
		return nil, err
	}
	return c.restConfig(), nil
}
