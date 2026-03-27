package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// mcs_domain data source
// ---------------------------------------------------------------------------

func TestAccDomainDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/domains", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 1, "uuid": "dom-uuid-1", "name": "prod", "description": "Production", "adom": "adom1", "zone": "zone1"},
				{"id": 2, "uuid": "dom-uuid-2", "name": "staging", "description": "Staging", "adom": "adom1", "zone": "zone2"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "test" {
  name = "prod"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain.test", "name", "prod"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "id", "1"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "uuid", "dom-uuid-1"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "description", "Production"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "adom", "adom1"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "zone", "zone1"),
				),
			},
		},
	})
}

func TestAccDomainDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/domains", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 1, "uuid": "dom-uuid-1", "name": "prod", "description": "Production", "adom": "adom1", "zone": "zone1"},
				{"id": 2, "uuid": "dom-uuid-2", "name": "staging", "description": "Staging", "adom": "adom1", "zone": "zone2"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "all" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.0.name", "prod"),
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.1.name", "staging"),
				),
			},
		},
	})
}

func TestAccDomainDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/domains", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Domain not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_job data source
// ---------------------------------------------------------------------------

func TestAccJobDataSource_Read(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/jobs/job", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                   42,
			"jobname":              "deploy-webserver",
			"timestamp":            "2025-01-01T00:00:00Z",
			"endtime":              "2025-01-01T00:05:00Z",
			"message":              "completed successfully",
			"dryrun":               false,
			"continue_on_failure":  true,
			"created_at_timestamp": "2025-01-01T00:00:00Z",
			"updated_at_timestamp": "2025-01-01T00:05:00Z",
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_job" "test" {
  id = 42
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_job.test", "id", "42"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "jobname", "deploy-webserver"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "message", "completed successfully"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "dryrun", "false"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "continue_on_failure", "true"),
				),
			},
		},
	})
}

func TestAccJobDataSource_APIError(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/jobs/job", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"detail":"forbidden"}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_job" "test" {
  id = 999
}`,
				ExpectError: regexpMustCompile(`Error reading job`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_network data source
// ---------------------------------------------------------------------------

func TestAccNetworkDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/networks", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "net-001", "name": "mgmt", "ipv4_prefix": "10.0.0.0/24", "vlan_id": 100},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network" "test" {
  name = "mgmt"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_network.test", "id", "net-001"),
					resource.TestCheckResourceAttr("data.mcs_network.test", "name", "mgmt"),
					resource.TestCheckResourceAttr("data.mcs_network.test", "ipv4_prefix", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("data.mcs_network.test", "vlan_id", "100"),
				),
			},
		},
	})
}

func TestAccNetworkDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/networks", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "net-001", "name": "mgmt", "ipv4_prefix": "10.0.0.0/24", "vlan_id": 100},
				{"id": "net-002", "name": "data", "ipv4_prefix": "10.1.0.0/24", "vlan_id": 200},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network" "all" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_network.all", "networks.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_network.all", "networks.0.name", "mgmt"),
					resource.TestCheckResourceAttr("data.mcs_network.all", "networks.1.name", "data"),
				),
			},
		},
	})
}

func TestAccNetworkDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/networks", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Network not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_tenant data source
// ---------------------------------------------------------------------------

func TestAccTenantDataSource_Read(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":1,"name":"My Tenant","description":"Main tenant account"}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_tenant" "test" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_tenant.test", "id", "1"),
					resource.TestCheckResourceAttr("data.mcs_tenant.test", "name", "My Tenant"),
					resource.TestCheckResourceAttr("data.mcs_tenant.test", "description", "Main tenant account"),
					resource.TestCheckResourceAttrSet("data.mcs_tenant.test", "raw_json"),
				),
			},
		},
	})
}

func TestAccTenantDataSource_AlternateIDKeys(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"pk":99,"name":"Alt Tenant"}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_tenant" "alt" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_tenant.alt", "id", "99"),
					resource.TestCheckResourceAttr("data.mcs_tenant.alt", "name", "Alt Tenant"),
				),
			},
		},
	})
}

func TestAccTenantDataSource_APIError(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"detail":"unauthorized"}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_tenant" "err" {
}`,
				ExpectError: regexpMustCompile(`Error reading tenant`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_zone data source
// ---------------------------------------------------------------------------

func TestAccZoneDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/zones", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"uuid": "zone-uuid-1", "name": "dmz", "description": "DMZ Zone", "adom": "adom1", "transit_vrf": "vrf1"},
				{"uuid": "zone-uuid-2", "name": "internal", "description": "Internal Zone", "adom": "adom1", "transit_vrf": "vrf2"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_zone" "test" {
  name = "dmz"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_zone.test", "name", "dmz"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "uuid", "zone-uuid-1"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "description", "DMZ Zone"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "transit_vrf", "vrf1"),
				),
			},
		},
	})
}

func TestAccZoneDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/zones", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"uuid": "zone-uuid-1", "name": "dmz", "description": "DMZ", "adom": "a1", "transit_vrf": "v1"},
				{"uuid": "zone-uuid-2", "name": "internal", "description": "Internal", "adom": "a1", "transit_vrf": "v2"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_zone" "all" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.0.name", "dmz"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.1.name", "internal"),
				),
			},
		},
	})
}

func TestAccZoneDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/zones", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_zone" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Zone not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall data source
// ---------------------------------------------------------------------------

func TestAccFirewallDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	fws := []map[string]interface{}{
		{"id": "fw-001", "name": "internet-fw", "description": "Internet firewall", "customer": "acme", "type": "internet"},
		{"id": "fw-002", "name": "wan-fw", "description": "WAN firewall", "customer": "acme", "type": "wan"},
	}

	mock.On("/api/networking/firewalls", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": fws})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_firewall" "test" {
  name = "internet-fw"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_firewall.test", "id", "fw-001"),
					resource.TestCheckResourceAttr("data.mcs_firewall.test", "name", "internet-fw"),
					resource.TestCheckResourceAttr("data.mcs_firewall.test", "type", "internet"),
				),
			},
		},
	})
}

func TestAccFirewallDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	fws := []map[string]interface{}{
		{"id": "fw-001", "name": "internet-fw", "description": "Internet firewall", "customer": "acme", "type": "internet"},
		{"id": "fw-002", "name": "wan-fw", "description": "WAN firewall", "customer": "acme", "type": "wan"},
	}

	mock.On("/api/networking/firewalls", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": fws})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_firewall" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "firewalls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "firewalls.0.name", "internet-fw"),
				),
			},
		},
	})
}

func TestAccFirewallDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/firewalls", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_firewall" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Firewall not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_interface data source
// ---------------------------------------------------------------------------

func TestAccInterfaceDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	ifaces := []map[string]interface{}{
		{"id": "if-001", "name": "eth0", "ipaddress": "10.0.0.5", "ipv6address": "::1", "network": "net-001", "macAddress": "AA:BB:CC:DD:EE:FF", "vm_name": "web-01"},
	}

	mock.On("/api/virtualization/interface", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": ifaces})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_interface" "test" {
  name = "eth0"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_interface.test", "id", "if-001"),
					resource.TestCheckResourceAttr("data.mcs_interface.test", "ipaddress", "10.0.0.5"),
					resource.TestCheckResourceAttr("data.mcs_interface.test", "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr("data.mcs_interface.test", "vm_name", "web-01"),
				),
			},
		},
	})
}

func TestAccInterfaceDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	ifaces := []map[string]interface{}{
		{"id": "if-001", "name": "eth0", "ipaddress": "10.0.0.5", "ipv6address": "::1", "network": "net-001", "macAddress": "AA:BB:CC:DD:EE:FF", "vm_name": "web-01"},
		{"id": "if-002", "name": "eth1", "ipaddress": "10.0.0.6", "ipv6address": "::2", "network": nil, "macAddress": "11:22:33:44:55:66", "vm_name": "db-01"},
	}

	mock.On("/api/virtualization/interface", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": ifaces})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_interface" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_interface.all", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_interface.all", "interfaces.0.name", "eth0"),
				),
			},
		},
	})
}

func TestAccInterfaceDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/interface", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_interface" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Interface not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_ippool data source
// ---------------------------------------------------------------------------

func TestAccIPPoolDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pools := []map[string]interface{}{
		{"id": "pool-001", "name": "Public Pool", "subnet": "203.0.113.0/24", "customer": "acme"},
	}

	mock.On("/api/networking/ippools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": pools})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ippool" "test" {
  name = "Public Pool"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_ippool.test", "id", "pool-001"),
					resource.TestCheckResourceAttr("data.mcs_ippool.test", "subnet", "203.0.113.0/24"),
					resource.TestCheckResourceAttr("data.mcs_ippool.test", "customer", "acme"),
				),
			},
		},
	})
}

func TestAccIPPoolDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pools := []map[string]interface{}{
		{"id": "pool-001", "name": "Public Pool", "subnet": "203.0.113.0/24", "customer": "acme"},
		{"id": "pool-002", "name": "Private Pool", "subnet": "10.0.0.0/8", "customer": nil},
	}

	mock.On("/api/networking/ippools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": pools})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ippool" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_ippool.all", "ip_pools.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_ippool.all", "ip_pools.0.name", "Public Pool"),
				),
			},
		},
	})
}

func TestAccIPPoolDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/ippools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ippool" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`IP pool not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_networkpool data source
// ---------------------------------------------------------------------------

func TestAccNetworkPoolDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pools := []map[string]interface{}{
		{"id": "np-001", "name": "LAN Pool", "network": "10.0.0.0/8", "description": "Internal LAN", "type": "lan", "enabled": true},
	}

	mock.On("/api/networking/networkpools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": pools})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_networkpool" "test" {
  name = "LAN Pool"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "id", "np-001"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "network", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "type", "lan"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccNetworkPoolDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pools := []map[string]interface{}{
		{"id": "np-001", "name": "LAN Pool", "network": "10.0.0.0/8", "description": "Internal LAN", "type": "lan", "enabled": true},
		{"id": "np-002", "name": "WAN Pool", "network": "172.16.0.0/12", "description": "WAN transit", "type": "wan", "enabled": false},
	}

	mock.On("/api/networking/networkpools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": pools})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_networkpool" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_networkpool.all", "network_pools.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.all", "network_pools.0.name", "LAN Pool"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.all", "network_pools.1.enabled", "false"),
				),
			},
		},
	})
}

func TestAccNetworkPoolDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/networkpools", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_networkpool" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Network pool not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtualmachine data source
// ---------------------------------------------------------------------------

func TestAccVirtualMachineDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	vms := []map[string]interface{}{
		{
			"id": "vm-001", "name": "web-01", "cpu": 4, "memory": 8192, "os": "Ubuntu 22.04",
			"disks": []map[string]interface{}{
				{"id": "disk-001", "name": "sda", "size": 100, "path": "/dev/sda", "type": "ssd"},
			},
			"interfaces": []map[string]interface{}{
				{"id": "if-001", "name": "eth0", "ipaddress": "10.0.0.5", "ipv6address": "::1", "network": "net-001", "macAddress": "AA:BB:CC:DD:EE:FF"},
			},
		},
	}

	mock.On("/api/virtualization/virtualmachine", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": vms})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtualmachine" "test" {
  name = "web-01"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "id", "vm-001"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "cpu", "4"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "memory", "8192"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "os", "Ubuntu 22.04"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "disks.0.name", "sda"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "interfaces.0.ipaddress", "10.0.0.5"),
				),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	vms := []map[string]interface{}{
		{
			"id": "vm-001", "name": "web-01", "cpu": 4, "memory": 8192, "os": "Ubuntu 22.04",
			"disks":      []map[string]interface{}{},
			"interfaces": []map[string]interface{}{},
		},
		{
			"id": "vm-002", "name": "db-01", "cpu": 8, "memory": 16384, "os": "Debian 12",
			"disks":      []map[string]interface{}{},
			"interfaces": []map[string]interface{}{},
		},
	}

	mock.On("/api/virtualization/virtualmachine", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": vms})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtualmachine" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.0.name", "web-01"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.1.name", "db-01"),
				),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/virtualmachine", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtualmachine" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Virtual machine not found`),
			},
		},
	})
}
