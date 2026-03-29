package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProvider_Metadata(t *testing.T) {
	p := New("1.2.3")()
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "mcs" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "mcs")
	}
	if resp.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", resp.Version, "1.2.3")
	}
}

func TestProvider_Schema(t *testing.T) {
	p := New("dev")()
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned errors: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, name := range []string{"host", "token", "insecure"} {
		if _, ok := attrs[name]; !ok {
			t.Errorf("schema missing attribute %q", name)
		}
	}

	if !attrs["token"].(interface{ IsSensitive() bool }).IsSensitive() {
		t.Error("token should be marked sensitive")
	}
}

func TestProvider_Configure_MissingHost(t *testing.T) {
	t.Setenv("MCS_HOST", "")
	t.Setenv("MCS_TOKEN", "")

	p := New("test")()

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	confResp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: configFromRaw(t, schemaResp.Schema, map[string]interface{}{
			"host":     nil,
			"token":    nil,
			"insecure": nil,
		}),
	}, confResp)

	if !confResp.Diagnostics.HasError() {
		t.Fatal("expected error for missing host")
	}

	found := false
	for _, d := range confResp.Diagnostics.Errors() {
		if d.Summary() == "Missing host" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Missing host' diagnostic")
	}
}

func TestProvider_Configure_MissingToken(t *testing.T) {
	t.Setenv("MCS_HOST", "https://example.com")
	t.Setenv("MCS_TOKEN", "")

	p := New("test")()

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	confResp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: configFromRaw(t, schemaResp.Schema, map[string]interface{}{
			"host":     nil,
			"token":    nil,
			"insecure": nil,
		}),
	}, confResp)

	if !confResp.Diagnostics.HasError() {
		t.Fatal("expected error for missing token")
	}

	found := false
	for _, d := range confResp.Diagnostics.Errors() {
		if d.Summary() == "Missing token" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Missing token' diagnostic")
	}
}

func TestProvider_Configure_ViaEnvVars(t *testing.T) {
	t.Setenv("MCS_HOST", "https://example.com")
	t.Setenv("MCS_TOKEN", "test-token-123")

	p := New("test")()

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	confResp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: configFromRaw(t, schemaResp.Schema, map[string]interface{}{
			"host":     nil,
			"token":    nil,
			"insecure": nil,
		}),
	}, confResp)

	if confResp.Diagnostics.HasError() {
		t.Fatalf("Configure() returned errors: %v", confResp.Diagnostics)
	}
	if confResp.ResourceData == nil {
		t.Error("ResourceData should be set")
	}
	if confResp.DataSourceData == nil {
		t.Error("DataSourceData should be set")
	}
}

func TestProvider_Configure_ExplicitValues(t *testing.T) {
	t.Setenv("MCS_HOST", "")
	t.Setenv("MCS_TOKEN", "")

	p := New("test")()

	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)

	confResp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{
		Config: configFromRaw(t, schemaResp.Schema, map[string]interface{}{
			"host":     "https://mcs.test.local",
			"token":    "my-token",
			"insecure": true,
		}),
	}, confResp)

	if confResp.Diagnostics.HasError() {
		t.Fatalf("Configure() returned errors: %v", confResp.Diagnostics)
	}
	if confResp.ResourceData == nil {
		t.Error("ResourceData should be set after explicit config")
	}
}

func TestProvider_Resources_Count(t *testing.T) {
	p := &MCSProvider{version: "test"}
	resources := p.Resources(context.Background())
	if len(resources) != 22 {
		t.Errorf("expected 22 resources, got %d", len(resources))
	}
}

func TestProvider_DataSources_Count(t *testing.T) {
	p := &MCSProvider{version: "test"}
	dataSources := p.DataSources(context.Background())
	if len(dataSources) != 28 {
		t.Errorf("expected 28 data sources, got %d", len(dataSources))
	}
}
