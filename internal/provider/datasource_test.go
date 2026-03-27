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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		fmt.Fprint(w, `{"detail":"forbidden"}`)
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		fmt.Fprint(w, `{"id":1,"name":"My Tenant","description":"Main tenant account"}`)
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
		fmt.Fprint(w, `{"pk":99,"name":"Alt Tenant"}`)
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
		fmt.Fprint(w, `{"detail":"unauthorized"}`)
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
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
