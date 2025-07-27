# Temporal Worker with JWT Authentication

An example of how to use a JWT to authenticate a Temporal Worker to a [Temporal Cloud Proxy](https://github.com/temporal-sa/temporal-cloud-proxy).

## Prerequisites
* Go 1.24.4+
* A Temporal Cloud account and namespace
* [Temporal Cloud Proxy](https://github.com/temporal-sa/temporal-cloud-proxy) configured for JWT authentication (see below)
* A JWT source (e.g., Hashicorp Vault)

## How to Run

```bash
# Run the worker
go run ./worker

# In another shell, run the starter/workflow
go run ./starter
````

## Example Worker Config

Add headers to the Temporal Client
```go
c, err := client.Dial(client.Options{
    HostPort:  "127.0.0.1:9000", // address of Temporal Cloud proxy
    Namespace: "<namespace>.<account>", // Temporal Cloud namespace must be provided
    HeadersProvider: &vaultauth.VaultHeadersProvider{
        Config: vaultauth.VaultConfig{
            VaultAddr: "http://127.0.0.1:8200",
            Username:  "username",
            Password:  "password",
            OidcRole:  "temporal-worker",
    },
    WorkloadId: "my-workload", // maps to proxy config
    },
})
```

## Example Proxy Config

```yaml
- workload_id: "my-workload"
  ...
  authentication:
    type: "jwt"
    config:
      jwks-url: "http://localhost:8200/v1/identity/oidc/.well-known/keys"
      audiences:
        - "temporal_cloud_proxy"
```