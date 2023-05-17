package vault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	log "github.com/sirupsen/logrus"
)

const (
	EnvAddr     = "VAULT_ADDR"
	EnvKubeRole = "VAULT_KUBERNETES_ROLE"
	EnvToken    = "VAULT_TOKEN"
)

type Vault struct {
	Address string
	Auth    struct {
		Token          string
		KubernetesRole string
	}
	Client *vault.Client
}

type KubernetesRole struct {
	Name                     string   `json:"name"`
	ServiceAccountNames      []string `json:"serviceAccountNames"`
	ServiceAccountNamespaces []string `json:"serviceAccountNamespaces"`
	Policies                 []string `json:"policies"`
	TokenTtl                 int64    `json:"tokenTtl"`
}

type Policy struct {
	Name     string       `json:"name"`
	Policies []PolicyRule `json:"policies"`
}

type PolicyRule struct {
	Path         string   `json:"path"`
	Capabilities []string `json:"capabilities"`
}

// New returns a new Vault struct. Configurations are read from environment variables.
func New() (*Vault, error) {
	v := Vault{}
	fields := log.Fields{}

	if e := os.Getenv(EnvAddr); e != "" {
		v.Address = e
		fields["url"] = e
	} else {
		msg := fmt.Sprintf("%s not defined, cannot connect to Vault", EnvAddr)
		return nil, errors.New(msg)
	}

	if e := os.Getenv(EnvKubeRole); e != "" {
		v.Auth.KubernetesRole = e
		v.Auth.Token = ""
		fields["kubernetes-role"] = e
	} else if e := os.Getenv(EnvToken); e != "" {
		v.Auth.KubernetesRole = ""
		v.Auth.Token = e
	} else {
		msg := fmt.Sprintf("%s or %s not defined, cannot authenticate to Vault", EnvKubeRole, EnvToken)
		return nil, errors.New(msg)
	}

	config := vault.DefaultConfig()
	config.Address = v.Address

	log.WithFields(fields).Infof("Connecting to HashiCorp Vault")

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	if v.Auth.KubernetesRole != "" {
		// Kubernetes auth
		k8sAuth, err := auth.NewKubernetesAuth(v.Auth.KubernetesRole)
		if err != nil {
			return nil, err
		}

		authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
		if err != nil {
			return nil, err
		}
		if authInfo == nil {
			return nil, errors.New("No auth info was returned after login")
		}

	} else {
		// Token auth
		client.SetToken(v.Auth.Token)
	}

	v.Client = client

	return &v, nil
}

func (v Vault) DeleteKubernetesRole(r KubernetesRole) error {
	path := fmt.Sprintf("auth/kubernetes/role/%s", r.Name)
	if _, err := v.Client.Logical().Delete(path); err != nil {
		return err
	}

	return nil
}

func (v Vault) DeletePolicy(p Policy) error {
	if err := v.Client.Sys().DeletePolicy(p.Name); err != nil {
		return err
	}

	return nil
}

func (v Vault) UpdateKubernetesRole(r KubernetesRole) error {
	path := fmt.Sprintf("auth/kubernetes/role/%s", r.Name)
	data := map[string]interface{}{
		"bound_service_account_names":      r.ServiceAccountNames,
		"bound_service_account_namespaces": r.ServiceAccountNamespaces,
		"policies":                         r.Policies,
		"token_ttl":                        r.TokenTtl,
	}

	if _, err := v.Client.Logical().Write(path, data); err != nil {
		return err
	}

	return nil
}

func (v Vault) UpdatePolicy(p Policy) error {
	if err := v.Client.Sys().PutPolicy(p.Name, p.toString()); err != nil {
		return err
	}

	return nil
}

func (p Policy) toString() string {
	var s string
	for _, policy := range p.Policies {
		capabilities := fmt.Sprintf(`"%s"`, strings.Join(policy.Capabilities, `", "`))
		s = s + fmt.Sprintf("path %q {\n\tcapabilities = [%s]\n}\n", policy.Path, capabilities)
	}

	return s
}
