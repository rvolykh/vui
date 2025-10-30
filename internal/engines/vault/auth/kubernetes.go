package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/utils"
	k8sauth "k8s.io/api/authentication/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	k8scmd "k8s.io/client-go/tools/clientcmd"
)

// authenticateWithKubernetes authenticates using Kubernetes
func (am *AuthManager) authenticateWithKubernetes(client *api.Client, profile *config.Profile) error {
	role := profile.AuthConfig.K8sRole
	if role == "" {
		return fmt.Errorf("k8s_role is required for Kubernetes authentication")
	}

	var (
		token string
		err   error

		opts []kubernetes.LoginOption
	)
	if profile.AuthConfig.K8sToken != "" {
		token = profile.AuthConfig.K8sToken
	} else if profile.AuthConfig.K8sServiceAccount != "" {
		token, err = am.getKubernetesToken(profile)
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes token: %w", err)
		}
	}
	if token != "" {
		opts = append(opts, kubernetes.WithServiceAccountToken(token))
	}

	k8sAuth, err := kubernetes.NewKubernetesAuth(role, opts...)
	if err != nil {
		return fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
	if err != nil {
		return fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return fmt.Errorf("no auth info was returned after login")
	}

	return nil
}

// getKubernetesToken gets Kubernetes service account token
func (am *AuthManager) getKubernetesToken(profile *config.Profile) (string, error) {
	homeDir, _ := os.UserHomeDir()

	var (
		serviceAccount = profile.AuthConfig.K8sServiceAccount

		kubeconfigPath = utils.Coalesce(profile.AuthConfig.K8sConfigPath, filepath.Join(homeDir, ".kube", "config"))
		audience       = utils.Coalesce(profile.AuthConfig.K8sAudience, "vault")
		namespace      = utils.Coalesce(profile.AuthConfig.K8sNamespace, "default")
		ttl            = utils.Coalesce(profile.AuthConfig.K8sTTL, int64(3600))
	)

	config, err := k8scmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to build Kubernetes config: %w", err)
	}

	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	tr := &k8sauth.TokenRequest{
		Spec: k8sauth.TokenRequestSpec{
			Audiences:         []string{audience},
			ExpirationSeconds: &ttl,
		},
	}
	tokenRequest, err := clientset.CoreV1().ServiceAccounts(namespace).
		CreateToken(context.TODO(), serviceAccount, tr, k8smeta.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes token: %w", err)
	}

	return tokenRequest.Status.Token, nil
}
