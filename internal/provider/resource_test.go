package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// mcs_alert
// ---------------------------------------------------------------------------

func TestAccAlertResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/alerts/alert", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 1, "resource": req["resource"], "event": req["event"],
				"createtime": "2025-01-01T00:00:00Z", "lastupdate": "2025-01-01T00:00:00Z",
				"duplicate_count": 0, "timeout": 300,
			}
			if v, ok := req["severity"]; ok {
				resp["severity"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 1, "resource": "web-server", "event": "cpu_high",
				"createtime": "2025-01-01T00:00:00Z", "lastupdate": "2025-01-01T00:00:00Z",
				"duplicate_count": 0, "timeout": 300, "severity": "major",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 1, "resource": req["resource"], "event": req["event"],
				"createtime": "2025-01-01T00:00:00Z", "lastupdate": "2025-01-01T00:00:00Z",
				"duplicate_count": 0, "timeout": 300,
			}
			if v, ok := req["severity"]; ok {
				resp["severity"] = v
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_alert" "test" {
  resource = "web-server"
  event    = "cpu_high"
  severity = "major"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_alert.test", "resource", "web-server"),
					resource.TestCheckResourceAttr("mcs_alert.test", "event", "cpu_high"),
					resource.TestCheckResourceAttr("mcs_alert.test", "id", "1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_certificate
// ---------------------------------------------------------------------------

func TestAccCertificateResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/certificate", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "cert-001"}
			for k, v := range req {
				resp[k] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "cert-001", "name": "my-cert",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "cert-001"}
			for k, v := range req {
				resp[k] = v
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_certificate" "test" {
  name = "my-cert"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_certificate.test", "id", "cert-001"),
					resource.TestCheckResourceAttr("mcs_certificate.test", "name", "my-cert"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_contact
// ---------------------------------------------------------------------------

func TestAccContactResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/contacts", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 10, "company": req["company"], "tenant": req["tenant"],
				"firstname": req["firstname"], "lastname": req["lastname"],
				"email": req["email"], "phone": req["phone"], "address": req["address"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 10, "company": "ACME", "tenant": 1,
				"firstname": "John", "lastname": "Doe",
				"email": "john@example.com", "phone": "123-456", "address": "123 Main St",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 10, "company": req["company"], "tenant": req["tenant"],
				"firstname": req["firstname"], "lastname": req["lastname"],
				"email": req["email"], "phone": req["phone"], "address": req["address"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_contact" "test" {
  company   = "ACME"
  firstname = "John"
  lastname  = "Doe"
  email     = "john@example.com"
  phone     = "123-456"
  address   = "123 Main St"
  tenant    = 1
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_contact.test", "id", "10"),
					resource.TestCheckResourceAttr("mcs_contact.test", "company", "ACME"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_customer
// ---------------------------------------------------------------------------

func TestAccCustomerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	customerResponse := map[string]interface{}{
		"id": "cust-001", "name": "Test Customer", "contractid": "",
		"tenant": 1, "sdm": 0,
		"tech_contacts": []int{}, "admin_contacts": []int{},
		"created_at_timestamp": "2025-01-01T00:00:00Z",
		"updated_at_timestamp": "2025-01-01T00:00:00Z",
	}

	mock.On("/api/tenant/customers", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(customerResponse)
		case http.MethodGet:
			json.NewEncoder(w).Encode(customerResponse)
		case http.MethodPut:
			json.NewEncoder(w).Encode(customerResponse)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name           = "Test Customer"
  contractid     = ""
  tenant         = 1
  sdm            = 0
  tech_contacts  = []
  admin_contacts = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "id", "cust-001"),
					resource.TestCheckResourceAttr("mcs_customer.test", "name", "Test Customer"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_dbl
// ---------------------------------------------------------------------------

func TestAccDblResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 5, "ipaddress": req["ipaddress"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "bad-host",
			}
			if v, ok := req["source"]; ok {
				resp["source"] = v
			}
			if v, ok := req["persistent"]; ok {
				resp["persistent"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 5, "ipaddress": "10.0.0.1",
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "bad-host",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 5, "ipaddress": req["ipaddress"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 2, "hostname": "bad-host",
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dbl" "test" {
  ipaddress = "10.0.0.1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dbl.test", "id", "5"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "ipaddress", "10.0.0.1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_domain_dbl
// ---------------------------------------------------------------------------

func TestAccDomainDblResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/domaindbl", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 3, "domainname": req["domainname"], "source": req["source"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 3, "domainname": "evil.example.com", "source": "manual",
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1,
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 3, "domainname": req["domainname"], "source": req["source"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 2,
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_dbl" "test" {
  domainname = "evil.example.com"
  source     = "manual"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "id", "3"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "domainname", "evil.example.com"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_action
// ---------------------------------------------------------------------------

func TestAccCsActionResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csaction", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "act-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "act-001", "name": "redirect-action",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "act-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_action" "test" {
  name = "redirect-action"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_cs_action.test", "id", "act-001"),
					resource.TestCheckResourceAttr("mcs_cs_action.test", "name", "redirect-action"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_policy
// ---------------------------------------------------------------------------

func TestAccCsPolicyResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/cspolicy", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "pol-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "pol-001", "name": "routing-policy",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "pol-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_policy" "test" {
  name = "routing-policy"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "id", "pol-001"),
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "name", "routing-policy"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_csv_server
// ---------------------------------------------------------------------------

func TestAccCsvServerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csvserver", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "csv-001", "name": req["name"], "ufname": req["ufname"],
				"type": req["type"],
			}
			if v, ok := req["policies"]; ok {
				resp["policies"] = v
			}
			if v, ok := req["certificate"]; ok {
				resp["certificate"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "csv-001", "name": "my-csv", "ufname": "my-csv-uf",
				"type": "HTTP",
				"policies": []string{"pol-1"}, "certificate": []string{"cert-1"},
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "csv-001", "name": req["name"], "ufname": req["ufname"],
				"type": req["type"],
			}
			if v, ok := req["policies"]; ok {
				resp["policies"] = v
			}
			if v, ok := req["certificate"]; ok {
				resp["certificate"] = v
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name        = "my-csv"
  ufname      = "my-csv-uf"
  type        = "HTTP"
  policies    = ["pol-1"]
  certificate = ["cert-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_csv_server.test", "id", "csv-001"),
					resource.TestCheckResourceAttr("mcs_csv_server.test", "name", "my-csv"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_object
// ---------------------------------------------------------------------------

func TestAccFirewallObjectResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "fw-obj-uuid-001", "used": false,
			}
			if v, ok := req["address"]; ok {
				resp["address"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "test-obj", "uuid": "fw-obj-uuid-001", "used": false,
				"address": "10.0.0.1",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "test-domain"
  name    = "test-obj"
  address = "10.0.0.1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "uuid", "fw-obj-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "name", "test-obj"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_object_group
// ---------------------------------------------------------------------------

func TestAccFirewallObjectGroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "grp-uuid-001", "used": false,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "test-group", "uuid": "grp-uuid-001", "used": false,
				"member": []string{"obj-1"},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "test-domain"
  name   = "test-group"
  member = ["obj-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "uuid", "grp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "name", "test-group"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_rule
// ---------------------------------------------------------------------------

func TestAccFirewallRuleResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			resp := map[string]interface{}{
				"enabled": true, "action": true, "used": false, "compliant": true,
				"uuid": "rule-uuid-001", "policyid": 100, "hit_count": 0,
				"last_hit": "", "group": "",
				"src": []string{"all"}, "dst": []string{"all"},
				"src_intf": []string{"any"}, "dst_intf": []string{"any"},
				"service": []string{"HTTP"},
				"compliancy_errors": []string{},
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"enabled": true, "action": true, "used": false, "compliant": true,
				"uuid": "rule-uuid-001", "policyid": 100, "hit_count": 0,
				"last_hit": "", "group": "",
				"src": []string{"all"}, "dst": []string{"all"},
				"src_intf": []string{"any"}, "dst_intf": []string{"any"},
				"service": []string{"HTTP"},
				"compliancy_errors": []string{},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain   = "test-domain"
  enabled  = true
  action   = true
  src      = ["all"]
  dst      = ["all"]
  src_intf = ["any"]
  dst_intf = ["any"]
  service  = ["HTTP"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "uuid", "rule-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "policyid", "100"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_service
// ---------------------------------------------------------------------------

func TestAccFirewallServiceResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "svc-uuid-001", "used": false,
			}
			if v, ok := req["tcp_portrange"]; ok {
				resp["tcp_portrange"] = v
			}
			if v, ok := req["udp_portrange"]; ok {
				resp["udp_portrange"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "https-svc", "uuid": "svc-uuid-001", "used": false,
				"tcp_portrange": []string{"443"}, "udp_portrange": []string{},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service" "test" {
  domain        = "test-domain"
  name          = "https-svc"
  tcp_portrange = ["443"]
  udp_portrange = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "uuid", "svc-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "name", "https-svc"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_service_group
// ---------------------------------------------------------------------------

func TestAccFirewallServiceGroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "svcgrp-uuid-001", "used": false,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "web-services", "uuid": "svcgrp-uuid-001", "used": false,
				"member": []string{"svc-1"},
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service_group" "test" {
  domain = "test-domain"
  name   = "web-services"
  member = ["svc-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "uuid", "svcgrp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "name", "web-services"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_monitor
// ---------------------------------------------------------------------------

func TestAccLbMonitorResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/monitor", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "mon-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mon-001", "name": "http-monitor",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mon-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name = "http-monitor"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_monitor.test", "id", "mon-001"),
					resource.TestCheckResourceAttr("mcs_lb_monitor.test", "name", "http-monitor"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup
// ---------------------------------------------------------------------------

func TestAccLbServicegroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroup", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "sg-001", "name": req["name"], "type": req["type"],
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sg-001", "name": "backend-sg", "type": "HTTP",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sg-001", "name": req["name"], "type": req["type"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name = "backend-sg"
  type = "HTTP"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "id", "sg-001"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "name", "backend-sg"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup_member
// ---------------------------------------------------------------------------

func TestAccLbServicegroupMemberResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroupmember", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "sgm-001", "address": req["address"], "servername": req["servername"],
				"port": 0,
			}
			if v, ok := req["port"]; ok {
				resp["port"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sgm-001", "address": "10.0.0.5", "servername": "web1", "port": 0,
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sgm-001", "address": req["address"], "servername": req["servername"],
				"port": req["port"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup_member" "test" {
  address    = "10.0.0.5"
  servername = "web1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup_member.test", "id", "sgm-001"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup_member.test", "address", "10.0.0.5"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lbv_server
// ---------------------------------------------------------------------------

func TestAccLbvServerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbvserver", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "lbv-001", "name": req["name"],
				"servicegroup": req["servicegroup"],
			}
			if v, ok := req["certificate"]; ok {
				resp["certificate"] = v
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "lbv-001", "name": "my-lbv",
				"servicegroup": []string{"sg-1"}, "certificate": []string{"cert-1"},
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "lbv-001", "name": req["name"],
				"servicegroup": req["servicegroup"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  servicegroup = ["sg-1"]
  certificate  = ["cert-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "id", "lbv-001"),
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "name", "my-lbv"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_monitor_ip
// ---------------------------------------------------------------------------

func TestAccMonitorIPResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/monitorip", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "mip-001", "ipaddress": req["ipaddress"], "customer": req["customer"],
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mip-001", "ipaddress": "10.0.0.100", "customer": "acme",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mip-001", "ipaddress": req["ipaddress"], "customer": req["customer"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_monitor_ip" "test" {
  ipaddress = "10.0.0.100"
  customer  = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_monitor_ip.test", "id", "mip-001"),
					resource.TestCheckResourceAttr("mcs_monitor_ip.test", "ipaddress", "10.0.0.100"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_site_to_site_vpn
// ---------------------------------------------------------------------------

func TestAccSiteToSiteVPNResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/vpn/site_to_site", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": req["name"],
				"state": req["state"], "last_status": req["last_status"],
				"resets": req["resets"], "last_check": req["last_check"],
				"last_reset": req["last_reset"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": "hq-to-branch",
				"state": "up", "last_status": "ok", "resets": 0,
				"last_check": "2025-01-01T12:00:00Z", "last_reset": "",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": req["name"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_site_to_site_vpn" "test" {
  name        = "hq-to-branch"
  state       = "up"
  last_status = "ok"
  resets      = 0
  last_check  = "2025-01-01T12:00:00Z"
  last_reset  = ""
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "id", "7"),
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "name", "hq-to-branch"),
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "uuid", "vpn-uuid-001"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtual_datacenter
// ---------------------------------------------------------------------------

func TestAccVirtualDatacenterResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/virtualdatacenter", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "vdc-001", "name": req["name"], "customer": req["customer"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "vdc-001", "name": "prod-dc", "customer": "acme",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "vdc-001", "name": req["name"], "customer": req["customer"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_virtual_datacenter" "test" {
  name     = "prod-dc"
  customer = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_virtual_datacenter.test", "id", "vdc-001"),
					resource.TestCheckResourceAttr("mcs_virtual_datacenter.test", "name", "prod-dc"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Resource not found → removal from state
// ---------------------------------------------------------------------------

func TestAccAlertResource_ReadNotFound_RemovesFromState(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	callCount := 0
	mock.On("/api/alerts/alert", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			callCount++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": callCount, "resource": "web", "event": "test",
				"createtime": "2025-01-01T00:00:00Z", "lastupdate": "2025-01-01T00:00:00Z",
				"duplicate_count": 0, "timeout": 300,
			})
		case http.MethodGet:
			if callCount > 1 {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id": callCount, "resource": "web", "event": "test",
					"createtime": "2025-01-01T00:00:00Z", "lastupdate": "2025-01-01T00:00:00Z",
					"duplicate_count": 0, "timeout": 300,
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"detail":"Not found."}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_alert" "test" {
  resource = "web"
  event    = "test"
}`,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
