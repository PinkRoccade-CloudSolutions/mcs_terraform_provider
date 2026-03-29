# MCS Terraform Provider

The MCS (Mijn Cloud Solutions) Terraform provider allows you to manage cloud infrastructure resources through the MCS API. It supports managing networking, firewalls, load balancing, virtual machines, VPN, alerting, and tenant administration.

**Provider source:** `registry.terraform.io/PinkRoccade-CloudSolutions/mcs`

---

## Table of Contents

- [Provider Configuration](#provider-configuration)
- [Authentication](#authentication)
- [Data Sources](#data-sources)
  - [mcs_tenant](#mcs_tenant)
  - [mcs_domain](#mcs_domain)
  - [mcs_zone](#mcs_zone)
  - [mcs_network](#mcs_network)
  - [mcs_networkpool](#mcs_networkpool)
  - [mcs_firewall](#mcs_firewall)
  - [mcs_interface](#mcs_interface)
  - [mcs_ippool](#mcs_ippool)
  - [mcs_virtualmachine](#mcs_virtualmachine)
  - [mcs_job](#mcs_job)
- [Resources](#resources)
  - [Tenant & Customer Management](#tenant--customer-management)
    - [mcs_contact](#mcs_contact)
    - [mcs_customer](#mcs_customer)
  - [Networking](#networking)
    - [mcs_public_ip_address](#mcs_public_ip_address)
    - [mcs_nat_translation](#mcs_nat_translation)
    - [mcs_site_to_site_vpn](#mcs_site_to_site_vpn)
  - [Firewall](#firewall)
    - [mcs_firewall_object](#mcs_firewall_object)
    - [mcs_firewall_object_group](#mcs_firewall_object_group)
    - [mcs_firewall_rule](#mcs_firewall_rule)
    - [mcs_firewall_service](#mcs_firewall_service)
    - [mcs_firewall_service_group](#mcs_firewall_service_group)
  - [Load Balancing](#load-balancing)
    - [mcs_certificate](#mcs_certificate)
    - [mcs_lb_monitor](#mcs_lb_monitor)
    - [mcs_lb_servicegroup](#mcs_lb_servicegroup)
    - [mcs_lb_servicegroup_member](#mcs_lb_servicegroup_member)
    - [mcs_lbv_server](#mcs_lbv_server)
    - [mcs_csv_server](#mcs_csv_server)
    - [mcs_cs_action](#mcs_cs_action)
    - [mcs_cs_policy](#mcs_cs_policy)
  - [Virtualization](#virtualization)
    - [mcs_virtual_datacenter](#mcs_virtual_datacenter)
  - [Alerting & Monitoring](#alerting--monitoring)
    - [mcs_alert](#mcs_alert)
    - [mcs_monitor_ip](#mcs_monitor_ip)
  - [Deny/Block Lists](#denyblock-lists)
    - [mcs_dbl](#mcs_dbl)
    - [mcs_domain_dbl](#mcs_domain_dbl)
- [Complete Example](#complete-example)

---

## Provider Configuration

```hcl
terraform {
  required_providers {
    mcs = {
      source = "registry.terraform.io/PinkRoccade-CloudSolutions/mcs"
    }
  }
}

provider "mcs" {
  host     = "https://mcs.example.com"
  token    = var.mcs_token
  insecure = false
}
```

### Provider Arguments

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `host`     | String | Yes      | Base URL of the MCS API (e.g. `https://mcs.example.com`). Can also be set via the `MCS_HOST` environment variable. |
| `token`    | String | Yes      | API token for MCS authentication. Can also be set via the `MCS_TOKEN` environment variable. This value is marked sensitive and will not appear in plan output. |
| `insecure` | Bool   | No       | Skip TLS certificate verification. Only use for development with self-signed certificates. Defaults to `false`. |

> **Note:** Both `host` and `token` are marked as optional in the schema to allow configuration via environment variables alone. However, at least one source (config or environment variable) must provide a value for each; otherwise the provider will return an error.

---

## Authentication

The provider authenticates to the MCS API using token-based authentication. You can supply credentials in two ways:

### Option 1: Provider block

```hcl
variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}
```

### Option 2: Environment variables

```bash
export MCS_HOST="https://mcs.example.com"
export MCS_TOKEN="your-api-token-here"
```

```hcl
provider "mcs" {}
```

Environment variables are evaluated first, then overridden by explicit provider block values if set.

---

## Data Sources

Data sources allow you to look up existing MCS objects for use in your Terraform configuration. Most data sources support two modes:

- **Single lookup** — provide `name` or `id` to fetch a specific object.
- **List mode** — omit the filter attributes to retrieve all objects in a list attribute.

---

### mcs_tenant

Retrieves information about the current tenant.

#### Example

```hcl
data "mcs_tenant" "current" {}

output "tenant_name" {
  value = data.mcs_tenant.current.name
}
```

#### Attributes

All attributes are computed (read-only):

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Tenant identifier. |
| `name`      | String | Tenant name. |
| `description` | String | Tenant description. |
| `raw_json`  | String | Full JSON response body (useful for debugging). |

---

### mcs_domain

Look up firewall domains. Provide `name` for a single match, or omit to list all domains.

#### Example — Single lookup

```hcl
data "mcs_domain" "production" {
  name = "VDOM-PROD"
}

output "domain_uuid" {
  value = data.mcs_domain.production.uuid
}
```

#### Example — List all

```hcl
data "mcs_domain" "all" {}

output "all_domains" {
  value = data.mcs_domain.all.domains
}
```

#### Attributes

| Attribute   | Type   | Mode     | Description |
|------------|--------|----------|-------------|
| `name`     | String | Optional | Exact domain name to look up. |
| `id`       | String | Computed | Domain ID (set when a single domain is matched). |
| `uuid`     | String | Computed | Domain UUID. |
| `description` | String | Computed | Domain description. |
| `adom`     | String | Computed | Administrative domain. |
| `zone`     | String | Computed | Associated zone. |
| `domains`  | List   | Computed | List of all domains (populated when `name` is not set). |

**Nested `domains` attributes:** `id`, `uuid`, `name`, `description`, `adom`, `zone` — all String, Computed.

---

### mcs_zone

Look up networking zones. Provide `name` for a single match, or omit to list all zones.

#### Example — Single lookup

```hcl
data "mcs_zone" "bc_zone" {
  name = "Pink Private Cloud - Business Critical - AM5"
}

output "zone_transit_vrf" {
  value = data.mcs_zone.bc_zone.transit_vrf
}
```

#### Example — List all

```hcl
data "mcs_zone" "all" {}

locals {
  target_zone = [for z in data.mcs_zone.all.zones : z if z.name == "PPC MC"][0]
}
```

#### Attributes

| Attribute     | Type   | Mode     | Description |
|--------------|--------|----------|-------------|
| `name`       | String | Optional | Exact zone name to look up. |
| `uuid`       | String | Computed | Zone UUID. |
| `description` | String | Computed | Zone description. |
| `adom`       | String | Computed | Administrative domain. |
| `transit_vrf` | String | Computed | Transit VRF identifier. |
| `zones`      | List   | Computed | List of all zones (populated when `name` is not set). |

**Nested `zones` attributes:** `uuid`, `name`, `description`, `adom`, `transit_vrf` — all String, Computed.

---

### mcs_network

Look up networks. Provide `name` for a single match, or omit to list all.

#### Example

```hcl
data "mcs_network" "production" {
  name = "001.2002.USR"
}

output "network_vlan" {
  value = data.mcs_network.production.vlan_id
}
```

#### Attributes

| Attribute     | Type   | Mode     | Description |
|--------------|--------|----------|-------------|
| `name`       | String | Optional | Exact network name to look up. |
| `id`         | String | Computed | Network ID. |
| `ipv4_prefix` | String | Computed | IPv4 CIDR prefix. |
| `vlan_id`    | String | Computed | VLAN identifier. |
| `networks`   | List   | Computed | List of all networks (populated when `name` is not set). |

**Nested `networks` attributes:** `id`, `name`, `ipv4_prefix`, `vlan_id` — all String, Computed.

---

### mcs_networkpool

Look up network pools. Provide `name` or `id` for a single match, or omit both to list all.

#### Example

```hcl
data "mcs_networkpool" "lan_pool" {
  name = "Production LAN Pool"
}
```

#### Attributes

| Attribute      | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `name`        | String | Optional | Exact network pool name. |
| `id`          | String | Optional | Network pool ID for direct lookup. |
| `network`     | String | Computed | Associated network. |
| `description` | String | Computed | Description. |
| `type`        | String | Computed | Pool type: `lan`, `wan`, or `transit`. |
| `enabled`     | Bool   | Computed | Whether the pool is enabled. |
| `network_pools` | List | Computed | List of all network pools (populated when `name` and `id` are not set). |

**Nested `network_pools` attributes:** `id`, `name`, `network`, `description`, `type` (String); `enabled` (Bool) — all Computed.

---

### mcs_firewall

Look up firewalls. Provide `name` or `id` for a single match, or omit both to list all.

#### Example

```hcl
data "mcs_firewall" "internet" {
  name = "Gemnet"
}

output "firewall_id" {
  value = data.mcs_firewall.internet.id
}
```

#### Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `name`      | String | Optional | Exact firewall name. |
| `id`        | String | Optional/Computed | Firewall ID (use as filter or read from result). |
| `description` | String | Computed | Firewall description. |
| `customer`  | String | Computed | Associated customer. |
| `type`      | String | Computed | Firewall type (e.g. `internet`, `wan`). |
| `firewalls` | List   | Computed | List of all firewalls (populated when `name` and `id` are not set). |

**Nested `firewalls` attributes:** `id`, `name`, `description`, `customer`, `type` — all String, Computed.

---

### mcs_interface

Look up network interfaces. Provide `name` or `id` for a single match, or omit both to list all.

#### Example

```hcl
data "mcs_interface" "eth0" {
  name = "eth0"
}

output "interface_ip" {
  value = data.mcs_interface.eth0.ipaddress
}
```

#### Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `name`      | String | Optional | Exact interface name. |
| `id`        | String | Optional/Computed | Interface ID. |
| `ipaddress` | String | Computed | IPv4 address. |
| `ipv6address` | String | Computed | IPv6 address. |
| `network`   | String | Computed | Associated network. |
| `mac_address` | String | Computed | MAC address. |
| `vm_name`   | String | Computed | Name of the VM this interface belongs to. |
| `interfaces` | List  | Computed | List of all interfaces (populated when `name` and `id` are not set). |

**Nested `interfaces` attributes:** `id`, `name`, `ipaddress`, `ipv6address`, `network`, `mac_address`, `vm_name` — all String, Computed.

---

### mcs_ippool

Look up IP pools. Provide `name` or `id` for a single match, or omit both to list all.

#### Example

```hcl
data "mcs_ippool" "nat_pool" {
  name = "NAT pool voor MKS"
}

output "pool_id" {
  value = data.mcs_ippool.nat_pool.id
}
```

#### Attributes

| Attribute  | Type   | Mode     | Description |
|-----------|--------|----------|-------------|
| `name`    | String | Optional | Exact pool name. |
| `id`      | String | Optional/Computed | Pool ID. |
| `subnet`  | String | Computed | Pool subnet. |
| `customer` | String | Computed | Associated customer. |
| `ip_pools` | List  | Computed | List of all IP pools (populated when `name` and `id` are not set). |

**Nested `ip_pools` attributes:** `id`, `name`, `subnet`, `customer` — all String, Computed.

---

### mcs_virtualmachine

Look up virtual machines. Provide `name` or `id` for a single VM, or omit both to list all VMs. Returns detailed information including disks and network interfaces.

#### Example — Single lookup

```hcl
data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

output "vm_cpu" {
  value = data.mcs_virtualmachine.webserver.cpu
}

output "vm_primary_ip" {
  value = data.mcs_virtualmachine.webserver.interfaces[0].ipaddress
}
```

#### Example — List all

```hcl
data "mcs_virtualmachine" "all" {}

output "vm_names" {
  value = [for vm in data.mcs_virtualmachine.all.virtual_machines : vm.name]
}
```

#### Attributes

| Attribute          | Type   | Mode     | Description |
|-------------------|--------|----------|-------------|
| `name`            | String | Optional | Exact VM name to look up. |
| `id`              | String | Optional/Computed | VM UUID. |
| `cpu`             | Number | Computed | Number of vCPUs (set when a single VM is matched). |
| `memory`          | Number | Computed | Memory in MB (set when a single VM is matched). |
| `os`              | String | Computed | Operating system (set when a single VM is matched). |
| `disks`           | List   | Computed | Disks attached to the matched VM. |
| `interfaces`      | List   | Computed | Network interfaces of the matched VM. |
| `virtual_machines` | List  | Computed | All VMs (populated when neither `name` nor `id` is set). |

**Nested `disks` attributes:**

| Attribute | Type   | Description |
|----------|--------|-------------|
| `id`     | String | Disk ID. |
| `name`   | String | Disk name. |
| `size`   | Number | Size in GB. |
| `path`   | String | Disk path. |
| `type`   | String | Disk type. |

**Nested `interfaces` attributes:**

| Attribute     | Type   | Description |
|--------------|--------|-------------|
| `id`         | String | Interface ID. |
| `name`       | String | Interface name. |
| `ipaddress`  | String | IPv4 address. |
| `ipv6address` | String | IPv6 address. |
| `network`    | String | Associated network. |
| `mac_address` | String | MAC address. |

---

### mcs_job

Look up a specific job by ID. This data source always requires an `id`.

#### Example

```hcl
data "mcs_job" "deployment" {
  id = 12345
}

output "job_status" {
  value = data.mcs_job.deployment.message
}
```

#### Attributes

| Attribute            | Type   | Mode     | Description |
|---------------------|--------|----------|-------------|
| `id`                | Number | **Required** | Job ID. |
| `jobname`           | String | Computed | Name of the job. |
| `timestamp`         | String | Computed | Job start timestamp. |
| `endtime`           | String | Computed | Job end timestamp. |
| `message`           | String | Computed | Job status message. |
| `dryrun`            | Bool   | Computed | Whether this was a dry run. |
| `continue_on_failure` | Bool | Computed | Whether the job continues on failure. |

---

## Resources

Resources allow you to create, update, and delete objects in MCS.

---

### Tenant & Customer Management

#### mcs_contact

Manages a contact record for a tenant.

##### Example

```hcl
resource "mcs_contact" "admin" {
  company   = "Example Corp"
  firstname = "John"
  lastname  = "Doe"
  email     = "john.doe@example.com"
  phone     = "+31 6 12345678"
}
```

##### Attributes

| Attribute   | Type   | Required | Description |
|------------|--------|----------|-------------|
| `company`  | String | **Yes**  | Company name. |
| `firstname` | String | No      | First name. |
| `lastname` | String | No       | Last name. |
| `email`    | String | No       | Email address. |
| `phone`    | String | No       | Phone number. |
| `address`  | String | No       | Address. |

**Read-only attributes:** `id` (String) — The contact ID.

---

#### mcs_customer

Manages a customer within a tenant.

##### Example

```hcl
resource "mcs_customer" "example" {
  name           = "Example Customer"
  tech_contacts  = [mcs_contact.admin.id]
  admin_contacts = [mcs_contact.admin.id]
}
```

##### Attributes

| Attribute        | Type         | Required | Description |
|-----------------|-------------|----------|-------------|
| `name`          | String       | **Yes**  | Customer name. |
| `contractid`    | String       | No       | Contract identifier. |
| `sdm`           | Number       | No       | Service Delivery Manager ID. |
| `tech_contacts` | List(Number) | No       | List of technical contact IDs. |
| `admin_contacts` | List(Number) | No      | List of administrative contact IDs. |

**Read-only attributes:** `id` (String) — The customer ID.

---

### Networking

#### mcs_public_ip_address

Manages a public IP address allocation from an IP pool.

##### Example

```hcl
resource "mcs_public_ip_address" "web" {
  pool        = data.mcs_ippool.nat_pool.id
  description = "Web server public IP"
  type        = "nat"
  customer    = mcs_customer.example.id
}
```

##### Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `pool`      | String | No       | UUID of the IP pool to allocate from. |
| `description` | String | No     | Description of the public IP address. |
| `type`      | String | No       | Type: `nat`, `vip`, or `loadbalancer`. |
| `customer`  | String | No       | Customer identifier. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | UUID of the public IP address. |
| `ip_address` | String | The assigned public IP address (determined by the API). |

---

#### mcs_nat_translation

Manages a NAT translation between a public IP and a private interface. Supports both 1:1 NAT and port forwarding.

##### Example — 1:1 NAT

```hcl
resource "mcs_nat_translation" "web_nat" {
  public_ip        = mcs_public_ip_address.web.id
  interface        = data.mcs_virtualmachine.webserver.interfaces[0].id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "one_to_one"
  customer         = mcs_customer.example.id
  description      = "1:1 NAT for web server"
}
```

##### Example — Port forward

```hcl
resource "mcs_nat_translation" "https_forward" {
  public_ip        = mcs_public_ip_address.web.id
  interface        = data.mcs_interface.eth0.id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "port_forward"
  public_port      = 443
  private_port     = 8443
  protocol         = "tcp"
  customer         = mcs_customer.example.id
  description      = "HTTPS port forward"
}
```

##### Attributes

| Attribute          | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `public_ip`       | String | **Yes**  | UUID of the public IP address. |
| `interface`       | String | **Yes**  | UUID of the private interface. |
| `firewall`        | String | **Yes**  | UUID of the firewall. |
| `translation_type` | String | **Yes** | Translation type: `one_to_one` or `port_forward`. |
| `customer`        | String | **Yes**  | Customer identifier. |
| `public_port`     | Number | No       | Public port (required for port forwarding). |
| `private_port`    | Number | No       | Private port (required for port forwarding). |
| `protocol`        | String | No       | Protocol: `tcp` or `udp`. |
| `description`     | String | No       | Description of the NAT translation. |
| `enabled`         | Bool   | No       | Whether the NAT translation is enabled. Defaults to `true`. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | UUID of the NAT translation. |
| `private_ip` | String | Resolved private IP address. |

---

#### mcs_site_to_site_vpn

Manages a site-to-site VPN tunnel.

##### Example

```hcl
resource "mcs_site_to_site_vpn" "office" {
  name  = "office-vpn-tunnel"
  state = "up"
}
```

##### Attributes

| Attribute     | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `name`       | String | **Yes**  | VPN tunnel name. |
| `state`      | String | No       | Desired tunnel state. |
| `last_status` | String | No      | Last known status. |
| `resets`     | Number | No       | Number of tunnel resets. |
| `last_check` | String | No       | Timestamp of the last health check. |
| `last_reset` | String | No       | Timestamp of the last reset. |

**Read-only attributes:** `id` (String), `uuid` (String).

---

### Firewall

> **Important:** Firewall resources (`mcs_firewall_object`, `mcs_firewall_object_group`, `mcs_firewall_rule`, `mcs_firewall_service`, `mcs_firewall_service_group`) support **create and delete only**. Updates require destroying and recreating the resource. Use `lifecycle { create_before_destroy = true }` if you need zero-downtime changes.

#### mcs_firewall_object

Manages a firewall address object.

##### Example

```hcl
resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-001"
  name    = "web-server-prod"
  address = "10.0.1.10"
  subnet  = "255.255.255.255"
  comment = "Production web server"
}
```

##### Attributes

| Attribute | Type   | Required | Description |
|----------|--------|----------|-------------|
| `domain` | String | **Yes**  | Firewall domain (VDOM). |
| `name`   | String | **Yes**  | Object name. |
| `address` | String | No      | IP address. |
| `subnet` | String | No       | Subnet mask. |
| `comment` | String | No      | Comment / description. |

**Read-only attributes:** `id` (String), `uuid` (String — firewall UUID), `used` (Bool — whether the object is in use by a policy).

---

#### mcs_firewall_object_group

Manages a firewall address group (a collection of firewall objects).

##### Example

```hcl
resource "mcs_firewall_object_group" "web_servers" {
  domain  = "VDOM-001"
  name    = "web-servers-group"
  comment = "All web server objects"
  member  = [
    mcs_firewall_object.web_server.name,
  ]
}
```

##### Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). |
| `name`   | String       | **Yes**  | Group name. |
| `comment` | String      | No       | Comment / description. |
| `member` | List(String) | No       | List of member object names. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

---

#### mcs_firewall_rule

Manages a firewall policy rule.

##### Example

```hcl
resource "mcs_firewall_rule" "allow_https" {
  domain   = "VDOM-001"
  enabled  = true
  src      = [mcs_firewall_object.client.name]
  dst      = [mcs_firewall_object_group.web_servers.name]
  src_intf = [data.mcs_network.production.name]
  dst_intf = ["any"]
  service  = ["HTTPS"]
  action   = true
  comment  = "Allow HTTPS to web servers"
}
```

##### Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). |
| `enabled` | Bool        | **Yes**  | Whether the rule is enabled. |
| `action` | Bool         | **Yes**  | `true` = allow, `false` = deny. |
| `src`    | List(String) | No       | Source address objects/groups. |
| `dst`    | List(String) | No       | Destination address objects/groups. |
| `src_intf` | List(String) | No     | Source interfaces. |
| `dst_intf` | List(String) | No     | Destination interfaces. |
| `service` | List(String) | No      | Service objects/groups (e.g. `"HTTPS"`, `"SSH"`). |
| `comment` | String      | No       | Comment / description. |

**Read-only attributes:**

| Attribute           | Type         | Description |
|--------------------|-------------|-------------|
| `id`               | String       | Rule ID. |
| `uuid`             | String       | Firewall UUID. |
| `policyid`         | Number       | Policy ID (used internally for API operations). |
| `used`             | Bool         | Whether the rule is in use. |
| `compliant`        | Bool         | Whether the rule is compliant. |
| `hit_count`        | Number       | Number of times the rule has been matched. |
| `last_hit`         | String       | Timestamp of the last hit. |
| `group`            | String       | Rule group. |
| `compliancy_errors` | List(String) | List of compliancy error messages. |

---

#### mcs_firewall_service

Manages a custom firewall service definition.

##### Example

```hcl
resource "mcs_firewall_service" "custom_app" {
  domain        = "VDOM-PROD"
  name          = "APP-8443"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["8443"]
  comment       = "Custom application port"
}
```

##### Attributes

| Attribute       | Type         | Required | Description |
|----------------|-------------|----------|-------------|
| `domain`       | String       | **Yes**  | Firewall domain (VDOM). |
| `name`         | String       | **Yes**  | Service name. |
| `protocol`     | String       | No       | Protocol type (e.g. `TCP/UDP/SCTP`). |
| `tcp_portrange` | List(String) | No      | TCP port ranges (e.g. `["80", "443", "8000-8999"]`). |
| `udp_portrange` | List(String) | No      | UDP port ranges. |
| `comment`      | String       | No       | Comment / description. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

---

#### mcs_firewall_service_group

Manages a group of firewall services.

##### Example

```hcl
resource "mcs_firewall_service_group" "web_services" {
  domain  = "VDOM-PROD"
  name    = "web-services"
  comment = "HTTP and HTTPS services"
  member  = ["HTTP", "HTTPS", mcs_firewall_service.custom_app.name]
}
```

##### Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). |
| `name`   | String       | **Yes**  | Service group name. |
| `comment` | String      | No       | Comment / description. |
| `member` | List(String) | No       | List of member service names. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

---

### Load Balancing

#### mcs_certificate

Manages a load balancer SSL certificate.

##### Example

```hcl
resource "mcs_certificate" "web_cert" {
  name         = "web-ssl-cert"
  ca           = false
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute          | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `name`            | String | No       | Certificate name. |
| `ca`              | Bool   | No       | Whether this is a CA certificate. |
| `valid_to_timestamp` | String | No    | Certificate expiry timestamp. |
| `loadbalancer`    | String | No       | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_lb_monitor

Manages a load balancer health monitor.

##### Example

```hcl
resource "mcs_lb_monitor" "http_monitor" {
  name         = "http-health-check"
  type         = "HTTP"
  interval     = 5
  resptimeout  = 2
  downtime     = 30
  respcode     = "200"
  httprequest  = "GET /health"
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
  customer     = mcs_customer.example.id
}
```

##### Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Monitor name. |
| `type`        | String | No       | Monitor type (e.g. `HTTP`, `TCP`). |
| `interval`    | Number | No       | Check interval in seconds. |
| `resptimeout` | Number | No       | Response timeout in seconds. |
| `downtime`    | Number | No       | Downtime threshold in seconds. |
| `respcode`    | String | No       | Expected response code (e.g. `"200"`). |
| `secure`      | String | No       | Whether to use secure connections. |
| `httprequest` | String | No       | HTTP request string (e.g. `"GET /health"`). |
| `loadbalancer` | String | No      | UUID of the load balancer. |
| `protected`   | Bool   | No       | Whether the monitor is protected. |
| `customer`    | String | No       | Customer identifier. |

**Read-only attributes:** `id` (String).

---

#### mcs_lb_servicegroup

Manages a load balancer service group.

##### Example

```hcl
resource "mcs_lb_servicegroup" "web_backend" {
  name          = "web-backend-sg"
  type          = "HTTP"
  state         = "enable"
  healthmonitor = "YES"
  customer      = mcs_customer.example.id
  loadbalancer  = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute       | Type         | Required | Description |
|----------------|-------------|----------|-------------|
| `name`         | String       | **Yes**  | Service group name. |
| `type`         | String       | **Yes**  | Service type (e.g. `HTTP`, `SSL`, `TCP`). |
| `state`        | String       | No       | State (e.g. `enable`, `disable`). |
| `members`      | List(String) | No       | List of member IDs. |
| `healthmonitor` | String      | No       | Health monitor setting. |
| `customer`     | String       | No       | Customer identifier. |
| `loadbalancer` | String       | No       | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_lb_servicegroup_member

Manages a member of a load balancer service group.

##### Example

```hcl
resource "mcs_lb_servicegroup_member" "web1" {
  address      = "10.0.1.10"
  port         = 8080
  servername   = "web-server-01"
  weight       = 100
  customer     = mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `address`     | String | **Yes**  | IP address of the backend server. |
| `servername`  | String | **Yes**  | Server name. |
| `port`        | Number | No       | Port number. Defaults to `0`. |
| `weight`      | Number | No       | Load balancing weight. |
| `state`       | String | No       | Member state. |
| `customer`    | String | No       | Customer identifier. |
| `loadbalancer` | String | No      | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_lbv_server

Manages a load balancer virtual server (LB vServer).

##### Example

```hcl
resource "mcs_lbv_server" "web_lb" {
  name         = "web-lbvserver"
  ipv46        = "10.0.0.100"
  port         = 443
  type         = "ssl"
  servicegroup = [mcs_lb_servicegroup.web_backend.id]
  customer     = mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute      | Type         | Required | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | **Yes**  | Virtual server name. |
| `servicegroup` | List(String) | **Yes** | List of service group IDs to bind. |
| `ipv46`       | String       | No       | IP address (v4 or v6). |
| `port`        | Number       | No       | Listening port. |
| `type`        | String       | No       | Protocol type (e.g. `ssl`, `http`, `tcp`). |
| `certificate` | List(String) | No       | List of SSL certificate IDs. |
| `customer`    | String       | No       | Customer identifier. |
| `loadbalancer` | String      | No       | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_csv_server

Manages a content switching virtual server (CS vServer).

##### Example

```hcl
resource "mcs_csv_server" "web_frontend" {
  name         = "web-csvserver"
  ufname       = "web-frontend"
  type         = "ssl"
  port         = 443
  ipv46        = "10.0.0.200"
  policies     = [mcs_cs_policy.routing.id]
  certificate  = [mcs_certificate.web_cert.id]
  customer     = mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute      | Type         | Required | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | **Yes**  | CS vServer name. |
| `ufname`      | String       | **Yes**  | User-friendly name. |
| `type`        | String       | **Yes**  | Protocol type (e.g. `ssl`, `http`). |
| `ipv46`       | String       | No       | IP address (v4 or v6). |
| `port`        | Number       | No       | Listening port. |
| `policies`    | List(String) | No       | List of CS policy IDs. |
| `certificate` | List(String) | No       | List of SSL certificate IDs. |
| `customer`    | String       | No       | Customer identifier. |
| `loadbalancer` | String      | No       | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_cs_action

Manages a content switching action.

##### Example

```hcl
resource "mcs_cs_action" "route_to_backend" {
  name         = "route-to-web"
  lbvserver    = mcs_lbv_server.web_lb.name
  customer     = mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Action name. |
| `lbvserver`   | String | No       | Target LB vServer name. |
| `customer`    | String | No       | Customer identifier. |
| `loadbalancer` | String | No      | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

#### mcs_cs_policy

Manages a content switching policy.

##### Example

```hcl
resource "mcs_cs_policy" "by_url" {
  name         = "route-by-url"
  action       = mcs_cs_action.route_to_backend.name
  expression   = "HTTP.REQ.URL.PATH.STARTSWITH(\"/api\")"
  customer     = mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

##### Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Policy name. |
| `action`      | String | No       | CS action name to invoke. |
| `expression`  | String | No       | Policy expression. |
| `customer`    | String | No       | Customer identifier. |
| `application` | String | No       | Application identifier. |
| `loadbalancer` | String | No      | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

---

### Virtualization

#### mcs_virtual_datacenter

Manages a virtual datacenter.

##### Example

```hcl
resource "mcs_virtual_datacenter" "production" {
  name     = "production-vdc"
  customer = mcs_customer.example.id
}
```

##### Attributes

| Attribute  | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `name`    | String | **Yes**  | Virtual datacenter name. |
| `customer` | String | No      | Customer identifier. |

**Read-only attributes:** `id` (String).

---

### Alerting & Monitoring

#### mcs_alert

Manages an alert in the MCS alerting system.

##### Example

```hcl
resource "mcs_alert" "high_cpu" {
  resource    = "web-server-01"
  event       = "HighCPUUtilization"
  severity    = "warning"
  status      = "open"
  environment = "production"
  service     = "web"
  text        = "CPU utilization exceeded 90% threshold"
  value       = "95%"
}
```

##### Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `resource`  | String | **Yes**  | Resource identifier the alert is related to. |
| `event`     | String | **Yes**  | Event name. |
| `environment` | String | No     | Environment (e.g. `production`, `staging`). |
| `correlate` | String | No       | Correlation key for grouping related alerts. |
| `service`   | String | No       | Service name. |
| `value`     | String | No       | Alert value. |
| `status`    | String | No       | Alert status: `open` or `closed`. |
| `text`      | String | No       | Alert description text. |
| `type`      | String | No       | Alert type. |
| `origin`    | String | No       | Alert origin. |
| `tags`      | String | No       | Tags. |
| `severity`  | String | No       | Severity level: `informational`, `minor`, `warning`, `major`, `critical`, or `fatal`. |

**Read-only attributes:**

| Attribute         | Type   | Description |
|------------------|--------|-------------|
| `id`             | String | Alert ID. |
| `url`            | String | Alert URL. |
| `createtime`     | String | Creation timestamp. |
| `lastupdate`     | String | Last update timestamp. |
| `duplicate_count` | Number | Number of duplicate alerts. |
| `timeout`        | Number | Alert timeout. |

---

#### mcs_monitor_ip

Manages an IP address monitoring entry.

##### Example

```hcl
resource "mcs_monitor_ip" "web_check" {
  ipaddress    = "203.0.113.10"
  customer     = mcs_customer.example.id
  notify_email = "ops@example.com"
  comment      = "Monitor production web server"
}
```

##### Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `ipaddress`   | String | **Yes**  | IP address to monitor. |
| `customer`    | String | **Yes**  | Customer identifier. |
| `notify_email` | String | No      | Email address for notifications. |
| `comment`     | String | No       | Comment. |

**Read-only attributes:**

| Attribute              | Type   | Description |
|-----------------------|--------|-------------|
| `id`                  | String | Monitor entry ID. |
| `timestamp`           | String | Creation timestamp. |
| `last_check_timestamp` | String | Timestamp of the last check. |

---

### Deny/Block Lists

#### mcs_dbl

Manages an IP address entry in the deny/block list (DBL).

##### Example

```hcl
resource "mcs_dbl" "blocked_ip" {
  ipaddress  = "192.0.2.100"
  source     = "manual"
  persistent = true
}
```

##### Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `ipaddress` | String | **Yes**  | IP address to block. |
| `source`    | String | No       | Source of the block entry (e.g. `manual`). |
| `persistent` | Bool  | No       | Whether the entry persists across resets. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Entry ID. |
| `timestamp` | String | Creation timestamp. |
| `occurrence` | Number | Number of occurrences. |
| `hostname`  | String | Resolved hostname. |

> **Note:** Read, update, and delete operations use the `ipaddress` as the lookup key, not the `id`.

---

#### mcs_domain_dbl

Manages a domain name entry in the deny/block list.

##### Example

```hcl
resource "mcs_domain_dbl" "blocked_domain" {
  domainname = "malicious-site.example"
  source     = "manual"
  persistent = true
}
```

##### Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `domainname` | String | **Yes** | Domain name to block. |
| `source`    | String | **Yes**  | Source of the block entry. |
| `persistent` | Bool  | No       | Whether the entry persists across resets. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Entry ID. |
| `timestamp` | String | Creation timestamp. |
| `occurrence` | Number | Number of occurrences. |

---

## Complete Example

The following example demonstrates a realistic workflow: setting up a customer, looking up infrastructure, allocating a public IP, and creating a NAT translation.

```hcl
terraform {
  required_providers {
    mcs = {
      source = "registry.terraform.io/PinkRoccade-CloudSolutions/mcs"
    }
  }
}

# Configure provider using environment variables or explicit values
variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}

# --- Look up existing infrastructure ---

data "mcs_tenant" "current" {}

data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_ippool" "nat_pool" {
  name = "NAT Pool"
}

data "mcs_network" "production" {
  name = "001.2002.USR"
}

data "mcs_zone" "bc_zone" {
  name = "Business Critical - AM5"
}

# --- Create tenant resources ---

resource "mcs_contact" "admin" {
  company   = "Example Corp"
  firstname = "Jane"
  lastname  = "Smith"
  email     = "jane.smith@example.com"
  phone     = "+31 6 98765432"
}

resource "mcs_customer" "production" {
  name           = "Production Customer"
  tech_contacts  = [mcs_contact.admin.id]
  admin_contacts = [mcs_contact.admin.id]
}

# --- Networking ---

resource "mcs_public_ip_address" "web" {
  pool        = data.mcs_ippool.nat_pool.id
  description = "Web server public IP"
  type        = "nat"
  customer    = mcs_customer.production.id
}

resource "mcs_nat_translation" "web_nat" {
  public_ip        = mcs_public_ip_address.web.id
  interface        = data.mcs_virtualmachine.webserver.interfaces[0].id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "one_to_one"
  customer         = mcs_customer.production.id
  description      = "1:1 NAT for web server"
}

# --- Firewall rules ---

resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-001"
  name    = "web-server-prod"
  address = data.mcs_virtualmachine.webserver.interfaces[0].ipaddress
  subnet  = "255.255.255.255"
  comment = "Production web server"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = "VDOM-001"
  enabled  = true
  src      = ["all"]
  dst      = [mcs_firewall_object.web_server.name]
  src_intf = ["any"]
  dst_intf = [data.mcs_network.production.name]
  service  = ["HTTPS"]
  action   = true
  comment  = "Allow HTTPS to web server"
}

# --- Virtual Datacenter ---

resource "mcs_virtual_datacenter" "production" {
  name     = "production-vdc"
  customer = mcs_customer.production.id
}

# --- Outputs ---

output "public_ip" {
  value = mcs_public_ip_address.web.ip_address
}

output "vm_details" {
  value = {
    name   = data.mcs_virtualmachine.webserver.name
    cpu    = data.mcs_virtualmachine.webserver.cpu
    memory = data.mcs_virtualmachine.webserver.memory
    os     = data.mcs_virtualmachine.webserver.os
  }
}
```
