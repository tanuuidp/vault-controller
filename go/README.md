# Vault Controller

A simple HTTP server meant to be used together with Metacontroller to manage custom Vault Kubernetes
resources.

## Configuration

The tool reads all configuration from environment variables.

| Name                    | Required         | Default | Description                                             |
|-------------------------|------------------|---------|---------------------------------------------------------|
| `LOG_LEVEL`             | false            | info    | Sets logging level: debug, info, warn, error, or fatal. |
| `VAULT_ADDR`            | true             |         | Base URL of the HashiCorp Vault instance.               |
| `VAULT_KUBERNETES_ROLE` | true<sup>1</sup> |         | Vault Kubernetes used for authentication.               |
| `VAULT_TOKEN`           | true<sup>1</sup> |         | Vault authentication token.                             |

<sup>1</sup> Either `VAULT_KUBERNETES_ROLE` or `VAULT_TOKEN` is required.

All environment variables listed in Vault Go-packages
[documentation](https://pkg.go.dev/github.com/hashicorp/vault/api#pkg-constants) valid and usable.
