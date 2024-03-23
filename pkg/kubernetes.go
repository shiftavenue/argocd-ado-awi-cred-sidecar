package pkg

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesHelper struct {
	logger    logr.Logger
	clientSet *kubernetes.Clientset
}

func NewKubernetesHelper(log logr.Logger) (*KubernetesHelper, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Info("Failed to get in cluster config, trying kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", "kubeconfig")
		if err != nil {
			log.Error(err, "Failed to get kubeconfig")
			return nil, err
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create clientset")
		return nil, err
	}
	return &KubernetesHelper{
		logger:    log,
		clientSet: clientSet,
	}, nil
}

func (k *KubernetesHelper) SearchSecret(namespace string, url string) (*corev1.Secret, error) {
	ctx := context.Background()
	secretList, err := k.clientSet.CoreV1().Secrets(namespace).List(ctx, v1.ListOptions{
		LabelSelector: LabelSelector,
	})
	if err != nil {
		k.logger.Error(err, "Failed to retrieve secretsList", "namespace", namespace, "url", url)
		return nil, err
	}
	for _, secret := range secretList.Items {
		k.logger.Info("Checking secret", "secret", secret.Name)

		secretUrl := string(secret.Data["url"])
		if strings.Contains(secretUrl, url) {
			k.logger.Info("Found secret", "name", secret.Name, "namespace", secret.Namespace)
			return &secret, nil
		}
	}
	return nil, nil
}

func (k *KubernetesHelper) UpdateSecret(accessToken string, secret *corev1.Secret) error {
	updatedSecret := secret.DeepCopy()
	updatedSecret.Data["password"] = []byte(accessToken)
	_, err := k.clientSet.CoreV1().Secrets(updatedSecret.Namespace).Update(context.Background(), updatedSecret, v1.UpdateOptions{})
	if err != nil {
		k.logger.Error(err, "Failed to update secret", "namespace", updatedSecret.Namespace, "name", updatedSecret.Name)
		return err
	}
	return nil
}
