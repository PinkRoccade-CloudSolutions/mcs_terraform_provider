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
				{
					"uuid": "zone-uuid-1", "name": "dmz", "description": "DMZ Zone", "adom": "adom1", "transit_vrf": "vrf1",
					"loadbalancers": []map[string]interface{}{
						{"id": "lb-001", "name": "lb-dmz-1"},
						{"id": "lb-002", "name": "lb-dmz-2"},
					},
				},
				{
					"uuid": "zone-uuid-2", "name": "internal", "description": "Internal Zone", "adom": "adom1", "transit_vrf": "vrf2",
					"loadbalancers": []map[string]interface{}{
						{"id": "lb-003", "name": "lb-internal-1"},
					},
				},
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
					resource.TestCheckResourceAttr("data.mcs_zone.test", "loadbalancers.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "loadbalancers.0.id", "lb-001"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "loadbalancers.0.name", "lb-dmz-1"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "loadbalancers.1.id", "lb-002"),
					resource.TestCheckResourceAttr("data.mcs_zone.test", "loadbalancers.1.name", "lb-dmz-2"),
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
				{
					"uuid": "zone-uuid-1", "name": "dmz", "description": "DMZ", "adom": "a1", "transit_vrf": "v1",
					"loadbalancers": []map[string]interface{}{
						{"id": "lb-001", "name": "lb-dmz-1"},
					},
				},
				{
					"uuid": "zone-uuid-2", "name": "internal", "description": "Internal", "adom": "a1", "transit_vrf": "v2",
					"loadbalancers": []map[string]interface{}{},
				},
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
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.0.loadbalancers.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.0.loadbalancers.0.id", "lb-001"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.0.loadbalancers.0.name", "lb-dmz-1"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.1.name", "internal"),
					resource.TestCheckResourceAttr("data.mcs_zone.all", "zones.1.loadbalancers.#", "0"),
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

// ---------------------------------------------------------------------------
// mcs_public_ip_address data source
// ---------------------------------------------------------------------------

func TestAccPublicIPAddressDataSource_ByIPAddress(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	addrs := []map[string]interface{}{
		{"id": "ip-001", "ip_address": "203.0.113.10", "pool": "pool-001", "description": "Web server", "status": "assigned", "type": "nat", "customer": "acme"},
		{"id": "ip-002", "ip_address": "203.0.113.20", "pool": "pool-001", "description": "DB server", "status": "available", "type": "nat", "customer": "acme"},
	}

	mock.On("/api/networking/publicipaddresss", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": addrs})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_public_ip_address" "test" {
  ip_address = "203.0.113.10"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "id", "ip-001"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "ip_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "description", "Web server"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "status", "assigned"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "type", "nat"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "customer", "acme"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "pool", "pool-001"),
				),
			},
		},
	})
}

func TestAccPublicIPAddressDataSource_ById(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/publicipaddresss", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "ip-001", "ip_address": "203.0.113.10", "pool": "pool-001",
			"description": "Web server", "status": "assigned", "type": "nat", "customer": "acme",
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_public_ip_address" "test" {
  id = "ip-001"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "id", "ip-001"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "ip_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "status", "assigned"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "type", "nat"),
				),
			},
		},
	})
}

func TestAccPublicIPAddressDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	addrs := []map[string]interface{}{
		{"id": "ip-001", "ip_address": "203.0.113.10", "pool": "pool-001", "description": "Web server", "status": "assigned", "type": "nat", "customer": "acme"},
		{"id": "ip-002", "ip_address": "203.0.113.20", "pool": nil, "description": "Spare IP", "status": "available", "type": "vip", "customer": nil},
	}

	mock.On("/api/networking/publicipaddresss", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": addrs})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_public_ip_address" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.all", "public_ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.all", "public_ip_addresses.0.ip_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.all", "public_ip_addresses.0.status", "assigned"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.all", "public_ip_addresses.1.ip_address", "203.0.113.20"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.all", "public_ip_addresses.1.status", "available"),
				),
			},
		},
	})
}

func TestAccPublicIPAddressDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/publicipaddresss", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_public_ip_address" "missing" {
  ip_address = "192.0.2.99"
}`,
				ExpectError: regexpMustCompile(`Public IP address not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_certificate data source
// ---------------------------------------------------------------------------

func TestAccCertificateDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/certificate", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "cert-001", "name": "web-cert", "ca": false, "valid_to_timestamp": "2026-12-31T23:59:59Z", "loadbalancer": "lb-001"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_certificate" "test" {
  name = "web-cert"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_certificate.test", "id", "cert-001"),
					resource.TestCheckResourceAttr("data.mcs_certificate.test", "name", "web-cert"),
					resource.TestCheckResourceAttr("data.mcs_certificate.test", "ca", "false"),
				),
			},
		},
	})
}

func TestAccCertificateDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/certificate", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "cert-001", "name": "web-cert", "ca": false},
				{"id": "cert-002", "name": "api-cert", "ca": true},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_certificate" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_certificate.all", "certificates.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_certificate.all", "certificates.0.name", "web-cert"),
					resource.TestCheckResourceAttr("data.mcs_certificate.all", "certificates.1.name", "api-cert"),
				),
			},
		},
	})
}

func TestAccCertificateDataSource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/certificate", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_certificate" "missing" {
  name = "nonexistent"
}`,
				ExpectError: regexpMustCompile(`Certificate not found`),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_action data source
// ---------------------------------------------------------------------------

func TestAccCsActionDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csaction", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "act-001", "name": "redirect-action", "lbvserver": "lbv-001", "customer": "acme"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_cs_action" "test" {
  name = "redirect-action"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_cs_action.test", "id", "act-001"),
					resource.TestCheckResourceAttr("data.mcs_cs_action.test", "name", "redirect-action"),
				),
			},
		},
	})
}

func TestAccCsActionDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csaction", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "act-001", "name": "redirect-action"},
				{"id": "act-002", "name": "rewrite-action"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_cs_action" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_cs_action.all", "cs_actions.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_cs_action.all", "cs_actions.0.name", "redirect-action"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_policy data source
// ---------------------------------------------------------------------------

func TestAccCsPolicyDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/cspolicy", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "pol-001", "name": "routing-policy", "expression": "HTTP.REQ.URL.PATH.STARTSWITH(\"/api\")"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_cs_policy" "test" {
  name = "routing-policy"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_cs_policy.test", "id", "pol-001"),
					resource.TestCheckResourceAttr("data.mcs_cs_policy.test", "name", "routing-policy"),
				),
			},
		},
	})
}

func TestAccCsPolicyDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/cspolicy", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "pol-001", "name": "routing-policy"},
				{"id": "pol-002", "name": "fallback-policy"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_cs_policy" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_cs_policy.all", "cs_policies.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_cs_policy.all", "cs_policies.0.name", "routing-policy"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_csv_server data source
// ---------------------------------------------------------------------------

func TestAccCsvServerDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csvserver", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "csv-001", "name": "web-frontend", "ufname": "web-uf", "type": "ssl", "port": 443, "ipaddress": "pip-uuid-1", "policies": []string{"pol-1"}, "certificate": []string{"cert-1"}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_csv_server" "test" {
  name = "web-frontend"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_csv_server.test", "id", "csv-001"),
					resource.TestCheckResourceAttr("data.mcs_csv_server.test", "name", "web-frontend"),
					resource.TestCheckResourceAttr("data.mcs_csv_server.test", "type", "ssl"),
					resource.TestCheckResourceAttr("data.mcs_csv_server.test", "ipaddress", "pip-uuid-1"),
				),
			},
		},
	})
}

func TestAccCsvServerDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csvserver", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "csv-001", "name": "web-frontend", "ufname": "web-uf", "type": "ssl", "port": 443, "ipaddress": "pip-uuid-1", "policies": []string{}, "certificate": []string{}},
				{"id": "csv-002", "name": "api-frontend", "ufname": "api-uf", "type": "http", "port": 80, "ipaddress": "pip-uuid-2", "policies": []string{}, "certificate": []string{}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_csv_server" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_csv_server.all", "csv_servers.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_csv_server.all", "csv_servers.0.name", "web-frontend"),
					resource.TestCheckResourceAttr("data.mcs_csv_server.all", "csv_servers.0.ipaddress", "pip-uuid-1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup data source
// ---------------------------------------------------------------------------

func TestAccLbServicegroupDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroup", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "sg-001", "name": "backend-sg", "type": "HTTP", "state": "enable", "members": []string{}, "healthmonitor": "YES"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lb_servicegroup" "test" {
  name = "backend-sg"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup.test", "id", "sg-001"),
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup.test", "name", "backend-sg"),
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup.test", "type", "HTTP"),
				),
			},
		},
	})
}

func TestAccLbServicegroupDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroup", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "sg-001", "name": "backend-sg", "type": "HTTP", "state": "enable", "members": []string{}, "healthmonitor": "YES"},
				{"id": "sg-002", "name": "api-sg", "type": "SSL", "state": "enable", "members": []string{}, "healthmonitor": "NO"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lb_servicegroup" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup.all", "lb_servicegroups.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup.all", "lb_servicegroups.0.name", "backend-sg"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup_member data source
// ---------------------------------------------------------------------------

func TestAccLbServicegroupMemberDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroupmember", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "sgm-001", "address": "10.0.0.5", "port": 8080, "servername": "web1", "weight": 100},
				{"id": "sgm-002", "address": "10.0.0.6", "port": 8080, "servername": "web2", "weight": 100},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lb_servicegroup_member" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup_member.all", "lb_servicegroup_members.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_lb_servicegroup_member.all", "lb_servicegroup_members.0.address", "10.0.0.5"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lbv_server data source
// ---------------------------------------------------------------------------

func TestAccLbvServerDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbvserver", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "lbv-001", "name": "web-lb", "type": "ssl", "port": 443, "ipaddress": "pip-uuid-2", "servicegroup": []string{"sg-1"}, "certificate": []string{"cert-1"}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lbv_server" "test" {
  name = "web-lb"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lbv_server.test", "id", "lbv-001"),
					resource.TestCheckResourceAttr("data.mcs_lbv_server.test", "name", "web-lb"),
					resource.TestCheckResourceAttr("data.mcs_lbv_server.test", "type", "ssl"),
					resource.TestCheckResourceAttr("data.mcs_lbv_server.test", "ipaddress", "pip-uuid-2"),
				),
			},
		},
	})
}

func TestAccLbvServerDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbvserver", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "lbv-001", "name": "web-lb", "type": "ssl", "port": 443, "ipaddress": "pip-uuid-2", "servicegroup": []string{}, "certificate": []string{}},
				{"id": "lbv-002", "name": "api-lb", "type": "http", "port": 80, "ipaddress": "pip-uuid-3", "servicegroup": []string{}, "certificate": []string{}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lbv_server" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lbv_server.all", "lbv_servers.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_lbv_server.all", "lbv_servers.0.name", "web-lb"),
					resource.TestCheckResourceAttr("data.mcs_lbv_server.all", "lbv_servers.0.ipaddress", "pip-uuid-2"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_monitor data source
// ---------------------------------------------------------------------------

func TestAccLbMonitorDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/monitor", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "mon-001", "name": "http-monitor", "type": "HTTP", "interval": 5, "resptimeout": 2, "downtime": 30, "protected": false},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lb_monitor" "test" {
  name = "http-monitor"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lb_monitor.test", "id", "mon-001"),
					resource.TestCheckResourceAttr("data.mcs_lb_monitor.test", "name", "http-monitor"),
					resource.TestCheckResourceAttr("data.mcs_lb_monitor.test", "type", "HTTP"),
				),
			},
		},
	})
}

func TestAccLbMonitorDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/monitor", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "mon-001", "name": "http-monitor", "type": "HTTP", "interval": 5, "resptimeout": 2, "downtime": 30, "protected": false},
				{"id": "mon-002", "name": "tcp-monitor", "type": "TCP", "interval": 10, "resptimeout": 5, "downtime": 60, "protected": true},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_lb_monitor" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_lb_monitor.all", "lb_monitors.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_lb_monitor.all", "lb_monitors.0.name", "http-monitor"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_dbl data source
// ---------------------------------------------------------------------------

func TestAccDblDataSource_ByIpAddress(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 5, "ipaddress": "10.0.0.1", "timestamp": "2025-01-01T00:00:00Z",
			"source": "manual", "occurrence": 1, "persistent": true, "hostname": "bad-host",
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_dbl" "test" {
  ipaddress = "10.0.0.1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_dbl.test", "id", "5"),
					resource.TestCheckResourceAttr("data.mcs_dbl.test", "ipaddress", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.mcs_dbl.test", "hostname", "bad-host"),
				),
			},
		},
	})
}

func TestAccDblDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 1, "ipaddress": "10.0.0.1", "timestamp": "2025-01-01T00:00:00Z", "source": "manual", "occurrence": 1, "persistent": true, "hostname": "host-1"},
				{"id": 2, "ipaddress": "10.0.0.2", "timestamp": "2025-01-01T00:00:00Z", "source": "auto", "occurrence": 3, "persistent": false, "hostname": "host-2"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_dbl" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.0.ipaddress", "10.0.0.1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_domain_dbl data source
// ---------------------------------------------------------------------------

func TestAccDomainDblDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/domaindbl", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 1, "domainname": "evil.example.com", "timestamp": "2025-01-01T00:00:00Z", "source": "manual", "persistent": true, "occurrence": 1},
				{"id": 2, "domainname": "bad.example.com", "timestamp": "2025-01-01T00:00:00Z", "source": "auto", "persistent": false, "occurrence": 5},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain_dbl" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.0.domainname", "evil.example.com"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_monitor_ip data source
// ---------------------------------------------------------------------------

func TestAccMonitorIPDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/monitorip", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "mip-001", "ipaddress": "10.0.0.100", "timestamp": "2025-01-01T00:00:00Z", "customer": "acme", "comment": "test", "notify_email": "ops@example.com", "last_check_timestamp": "2025-01-01T12:00:00Z"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_monitor_ip" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_monitor_ip.all", "monitor_ips.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_monitor_ip.all", "monitor_ips.0.ipaddress", "10.0.0.100"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_contact data source
// ---------------------------------------------------------------------------

func TestAccContactDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/contacts", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 10, "company": "ACME", "firstname": "John", "lastname": "Doe", "email": "john@example.com", "phone": "123-456", "address": "123 Main St"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_contact" "test" {
  name = "ACME"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_contact.test", "id", "10"),
					resource.TestCheckResourceAttr("data.mcs_contact.test", "company", "ACME"),
					resource.TestCheckResourceAttr("data.mcs_contact.test", "firstname", "John"),
				),
			},
		},
	})
}

func TestAccContactDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/contacts", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 10, "company": "ACME", "firstname": "John", "lastname": "Doe", "email": "john@example.com", "phone": "123", "address": "123 Main St"},
				{"id": 11, "company": "Globex", "firstname": "Jane", "lastname": "Smith", "email": "jane@globex.com", "phone": "456", "address": "456 Oak Ave"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_contact" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_contact.all", "contacts.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_contact.all", "contacts.0.company", "ACME"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_customer data source
// ---------------------------------------------------------------------------

func TestAccCustomerDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/customers", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "cust-001", "name": "Test Customer", "contractid": "C-001", "admin_contacts": []int{1}, "tech_contacts": []int{2}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_customer" "test" {
  name = "Test Customer"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_customer.test", "id", "cust-001"),
					resource.TestCheckResourceAttr("data.mcs_customer.test", "name", "Test Customer"),
					resource.TestCheckResourceAttr("data.mcs_customer.test", "contractid", "C-001"),
				),
			},
		},
	})
}

func TestAccCustomerDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/customers", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "cust-001", "name": "Test Customer", "contractid": "", "admin_contacts": []int{}, "tech_contacts": []int{}},
				{"id": "cust-002", "name": "Other Customer", "contractid": "C-002", "admin_contacts": []int{}, "tech_contacts": []int{}},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_customer" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_customer.all", "customers.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_customer.all", "customers.0.name", "Test Customer"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_nat_translation data source
// ---------------------------------------------------------------------------

func TestAccNATTranslationDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/nattranslations", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "nat-001", "public_ip": "pip-1", "interface": "if-1", "firewall": "fw-1", "translation": "10.0.0.1 -> 192.168.1.1", "private_ip": "192.168.1.1", "translation_type": "one_to_one", "protocol": "tcp", "customer": "acme", "description": "Web NAT", "state": "synced", "enabled": true},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_nat_translation" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_nat_translation.all", "nat_translations.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_nat_translation.all", "nat_translations.0.translation_type", "one_to_one"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_site_to_site_vpn data source
// ---------------------------------------------------------------------------

func TestAccSiteToSiteVPNDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/vpn/site_to_site", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 7, "uuid": "vpn-uuid-001", "name": "hq-to-branch", "state": "up", "last_status": "ok", "resets": 0, "last_check": "2025-01-01T12:00:00Z", "last_reset": "", "created_at_timestamp": "2025-01-01T00:00:00Z", "updated_at_timestamp": "2025-01-01T00:00:00Z"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_site_to_site_vpn" "test" {
  name = "hq-to-branch"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_site_to_site_vpn.test", "id", "7"),
					resource.TestCheckResourceAttr("data.mcs_site_to_site_vpn.test", "name", "hq-to-branch"),
					resource.TestCheckResourceAttr("data.mcs_site_to_site_vpn.test", "uuid", "vpn-uuid-001"),
				),
			},
		},
	})
}

func TestAccSiteToSiteVPNDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/vpn/site_to_site", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 7, "uuid": "vpn-uuid-001", "name": "hq-to-branch", "state": "up", "last_status": "ok", "resets": 0, "last_check": "", "last_reset": "", "created_at_timestamp": "2025-01-01T00:00:00Z", "updated_at_timestamp": "2025-01-01T00:00:00Z"},
				{"id": 8, "uuid": "vpn-uuid-002", "name": "dc-to-dc", "state": "down", "last_status": "error", "resets": 2, "last_check": "", "last_reset": "", "created_at_timestamp": "2025-01-01T00:00:00Z", "updated_at_timestamp": "2025-01-01T00:00:00Z"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_site_to_site_vpn" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_site_to_site_vpn.all", "vpns.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_site_to_site_vpn.all", "vpns.0.name", "hq-to-branch"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtual_datacenter data source
// ---------------------------------------------------------------------------

func TestAccVirtualDatacenterDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/virtualdatacenter", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "vdc-001", "name": "prod-dc", "customer": "acme"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtual_datacenter" "test" {
  name = "prod-dc"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.test", "id", "vdc-001"),
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.test", "name", "prod-dc"),
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.test", "customer", "acme"),
				),
			},
		},
	})
}

func TestAccVirtualDatacenterDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/virtualdatacenter", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "vdc-001", "name": "prod-dc", "customer": "acme"},
				{"id": "vdc-002", "name": "staging-dc", "customer": "acme"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtual_datacenter" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.all", "virtual_datacenters.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.all", "virtual_datacenters.0.name", "prod-dc"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_disk data source
// ---------------------------------------------------------------------------

func TestAccDiskDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/disk", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "disk-001", "name": "sda", "size": 100, "path": "/dev/sda", "type": "thin"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_disk" "test" {
  name = "sda"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_disk.test", "id", "disk-001"),
					resource.TestCheckResourceAttr("data.mcs_disk.test", "name", "sda"),
					resource.TestCheckResourceAttr("data.mcs_disk.test", "size", "100"),
					resource.TestCheckResourceAttr("data.mcs_disk.test", "type", "thin"),
				),
			},
		},
	})
}

func TestAccDiskDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/disk", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": "disk-001", "name": "sda", "size": 100, "path": "/dev/sda", "type": "thin"},
				{"id": "disk-002", "name": "sdb", "size": 200, "path": "/dev/sdb", "type": "thick"},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_disk" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_disk.all", "disks.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_disk.all", "disks.0.name", "sda"),
					resource.TestCheckResourceAttr("data.mcs_disk.all", "disks.1.name", "sdb"),
				),
			},
		},
	})
}
