package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func regexpMustCompile(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

func configFromRaw(t *testing.T, s schema.Schema, values map[string]interface{}) tfsdk.Config {
	t.Helper()

	attrTypes := make(map[string]tftypes.Type, len(s.Attributes))
	for k, attr := range s.Attributes {
		attrTypes[k] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	attrValues := make(map[string]tftypes.Value, len(s.Attributes))
	for k, attrType := range attrTypes {
		v, ok := values[k]
		if !ok || v == nil {
			attrValues[k] = tftypes.NewValue(attrType, nil)
		} else {
			attrValues[k] = tftypes.NewValue(attrType, v)
		}
	}

	raw := tftypes.NewValue(objType, attrValues)
	return tfsdk.Config{
		Schema: s,
		Raw:    raw,
	}
}

type mockRoute struct {
	prefix  string
	handler func(w http.ResponseWriter, r *http.Request, body []byte)
}

type mockAPIServer struct {
	mu     sync.Mutex
	routes []mockRoute
	calls  []recordedCall
	srv    *httptest.Server
}

type recordedCall struct {
	Method string
	Path   string
	Body   string
}

func newMockAPIServer() *mockAPIServer {
	m := &mockAPIServer{}
	m.srv = httptest.NewServer(http.HandlerFunc(m.handle))
	return m
}

func (m *mockAPIServer) URL() string {
	return m.srv.URL
}

func (m *mockAPIServer) Close() {
	m.srv.Close()
}

func (m *mockAPIServer) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	m.mu.Lock()
	m.calls = append(m.calls, recordedCall{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   string(body),
	})
	m.mu.Unlock()

	path := r.URL.Path

	m.mu.Lock()
	routes := make([]mockRoute, len(m.routes))
	copy(routes, m.routes)
	m.mu.Unlock()

	for _, route := range routes {
		if strings.HasPrefix(path, route.prefix) {
			route.handler(w, r, body)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, `{"detail":"Not found: %s %s"}`, r.Method, path)
}

func (m *mockAPIServer) On(prefix string, handler func(w http.ResponseWriter, r *http.Request, body []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = append(m.routes, mockRoute{prefix: prefix, handler: handler})
}

func (m *mockAPIServer) OnCRUD(basePath string, resourceJSON string) {
	var nextID = 1

	m.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var merged map[string]interface{}
			_ = json.Unmarshal([]byte(resourceJSON), &merged)
			if len(body) > 0 {
				var reqBody map[string]interface{}
				_ = json.Unmarshal(body, &reqBody)
				for k, v := range reqBody {
					merged[k] = v
				}
			}
			if _, ok := merged["id"]; !ok {
				merged["id"] = nextID
				nextID++
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(merged)

		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, resourceJSON)

		case http.MethodPut, http.MethodPatch:
			var merged map[string]interface{}
			_ = json.Unmarshal([]byte(resourceJSON), &merged)
			if len(body) > 0 {
				var reqBody map[string]interface{}
				_ = json.Unmarshal(body, &reqBody)
				for k, v := range reqBody {
					merged[k] = v
				}
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(merged)

		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (m *mockAPIServer) Calls() []recordedCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]recordedCall, len(m.calls))
	copy(result, m.calls)
	return result
}

func testProtoV6ProviderFactories(serverURL string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"mcs": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func providerConfigBlock(serverURL string) string {
	return fmt.Sprintf(`
provider "mcs" {
  host     = %q
  token    = "test-token"
  insecure = true
}
`, serverURL)
}
