# Terraform Provider MCS

The MCS (Mijn Cloud Solutions) Terraform provider allows you to manage cloud infrastructure resources through the MCS API. It supports managing networking, firewalls, load balancing, virtual machines, VPN, alerting, and tenant administration.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23 (to build the provider)

## Using the Provider

Configure the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    mcs = {
      source  = "PinkRoccade-CloudSolutions/mcs"
      version = "~> 0.0.2"
    }
  }
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}
```

Authentication can be configured via provider arguments or environment variables:

```bash
export MCS_HOST="https://mcs.example.com"
export MCS_TOKEN="your-api-token"
```

For full documentation on all supported resources and data sources, see the [docs](./docs/index.md).

## Developing the Provider

### Building

```bash
make build
```

### Installing locally

```bash
make install
```

### Running tests

```bash
make test
```

### Running acceptance tests

Acceptance tests create real resources and require a configured MCS environment:

```bash
make testacc
```

### Using development overrides

For local development without installing, use the `dev.tfrc` CLI override:

```bash
export TF_CLI_CONFIG_FILE=/path/to/this/repo/dev.tfrc
terraform plan
```

## Releasing

Releases are automated via GitHub Actions. To create a new release:

1. Ensure the `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets are configured in the GitHub repository settings.
2. Tag the commit:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. The release workflow will build binaries for all platforms, sign the checksums, and create a draft GitHub release.
4. Review and publish the draft release to make it available on the Terraform Registry.
