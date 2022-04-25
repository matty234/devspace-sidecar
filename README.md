# Dev Space Configuration Container

This init container sets up the development environment with correct DNS settings and PKI from Hashicorp Vault provider.

The following layout is anticipated in Hashicorp Vault:

-> kv/cloudflare/{{root host name}}
-> pki/{{root host name}}/ca

## Anticipated role configuration

```
path "kv/cloudflare/{{root host name}}" {
    capabilities = ["read"]
}
​
path "pki/{{root host name}}" {
    capabilities = ["read", "create", "update", "list", "delete"]
}
​
```

## Anticipated Service

A service should be created with a label specifying that it points to correct dev-space host:

```
apiVersion: v1
kind: Service
metadata:
  name: dev-space-service
  labels:
    for-devspace: "abcds1233"
spec:
    type: LoadBalancer
    ports:
    - port: 80
      targetPort: 80
    selector:
        app: dev-space-service

```