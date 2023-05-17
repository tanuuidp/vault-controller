# Vault Controller

Manage Vault roles and policies easily through a
[Metacontroller](https://github.com/metacontroller/metacontroller)-module. No more manual permission
mingle for each cluster resource!

## Requirements

* Kubernetes cluster
* Metacontroller v4.10+ (tested with given version, older might work as well)
* Vault running in the cluster (in vault namespace)
* Kubernetes authentication enabled, with role which enables the following policy:

    ```hcl
    path "auth/kubernetes/role/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
    }

    path "sys/policies/acl/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
    }
    ```

## Deplyoment

0. Have Metacontroller and Vault deployed.
1. Build the application using provided [`Dockerfile`](go/Dockerfile) and push to an image registry.
1. Fill in image details to [`vault-controller.yaml`](metacontroller/vault-controller.yaml).
1. Deploy the required CRDs with [`crdv1.yaml`](v1/crdv1.yaml).
1. Deploy the [`vault-controller.yaml`](metacontroller/vault-controller.yaml)-manifest.

## Usage

After the module is deployed and running, it should register all `VaultKubernetesRole` and
`VaultPolicy` resources deployed into the vault-namespace and create corresponding roles and
policies into Vault.

There are examples for each resource-type in the [`examples/`](examples/)-directory.
