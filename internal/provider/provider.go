package provider

import (
	"context"
	"os"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &MCSProvider{}

type MCSProvider struct {
	version string
}

type MCSProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MCSProvider{version: version}
	}
}

func (p *MCSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mcs"
	resp.Version = p.version
}

func (p *MCSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Mijn Cloud Solutions (MCS) API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Base URL of the MCS API (e.g. https://mcs.example.com). Can also be set via MCS_HOST environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API token for MCS authentication. Can also be set via MCS_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Only use for development with self-signed certificates. Defaults to false.",
				Optional:    true,
			},
		},
	}
}

func (p *MCSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MCSProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("MCS_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddError("Missing host", "The MCS API host must be set via the provider configuration or MCS_HOST environment variable.")
		return
	}

	token := os.Getenv("MCS_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError("Missing token", "The MCS API token must be set via the provider configuration or MCS_TOKEN environment variable.")
		return
	}

	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	client := apiclient.NewClient(host, token, insecure)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *MCSProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCertificateResource,
		NewContactResource,
		NewCsActionResource,
		NewCsPolicyResource,
		NewCsvServerResource,
		NewCustomerResource,
		NewDblResource,
		NewDomainDblResource,
		NewFirewallObjectResource,
		NewFirewallObjectGroupResource,
		NewFirewallRuleResource,
		NewFirewallServiceResource,
		NewFirewallServiceGroupResource,
		NewLbMonitorResource,
		NewLbServicegroupResource,
		NewLbServicegroupMemberResource,
		NewLbvServerResource,
		NewMonitorIPResource,
		NewNATTranslationResource,
		NewPublicIPAddressResource,
		NewSiteToSiteVPNResource,
		NewVirtualDatacenterResource,
	}
}

func (p *MCSProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCertificateDataSource,
		NewContactDataSource,
		NewCsActionDataSource,
		NewCsPolicyDataSource,
		NewCsvServerDataSource,
		NewCustomerDataSource,
		NewDblDataSource,
		NewDiskDataSource,
		NewDomainDataSource,
		NewDomainDblDataSource,
		NewFirewallDataSource,
		NewInterfaceDataSource,
		NewIPPoolDataSource,
		NewJobDataSource,
		NewLbMonitorDataSource,
		NewLbServicegroupDataSource,
		NewLbServicegroupMemberDataSource,
		NewLbvServerDataSource,
		NewMonitorIPDataSource,
		NewNATTranslationDataSource,
		NewNetworkDataSource,
		NewNetworkPoolDataSource,
		NewPublicIPAddressDataSource,
		NewSiteToSiteVPNDataSource,
		NewTenantDataSource,
		NewVirtualDatacenterDataSource,
		NewVirtualMachineDataSource,
		NewZoneDataSource,
	}
}
