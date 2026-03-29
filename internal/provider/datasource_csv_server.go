package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &CsvServerDataSource{}

type CsvServerDataSource struct {
	client *apiclient.Client
}

type CsvServerDataSourceModel struct {
	Name         types.String `tfsdk:"name"`
	Id           types.String `tfsdk:"id"`
	Ufname       types.String `tfsdk:"ufname"`
	Ipv46        types.String `tfsdk:"ipv46"`
	Port         types.Int64  `tfsdk:"port"`
	Type         types.String `tfsdk:"type"`
	Policies     types.List   `tfsdk:"policies"`
	Certificate  types.List   `tfsdk:"certificate"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
	CsvServers   []CsvServerModel `tfsdk:"csv_servers"`
}

type CsvServerModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Ufname       types.String `tfsdk:"ufname"`
	Ipv46        types.String `tfsdk:"ipv46"`
	Port         types.Int64  `tfsdk:"port"`
	Type         types.String `tfsdk:"type"`
	Policies     types.List   `tfsdk:"policies"`
	Certificate  types.List   `tfsdk:"certificate"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csvServerDSAPIModel struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Ufname       string   `json:"ufname"`
	Ipv46        *string  `json:"ipv46,omitempty"`
	Port         *int64   `json:"port,omitempty"`
	Type         string   `json:"type"`
	Policies     []string `json:"policies,omitempty"`
	Customer     *string  `json:"customer,omitempty"`
	Certificate  []string `json:"certificate,omitempty"`
	Loadbalancer *string  `json:"loadbalancer,omitempty"`
}

func NewCsvServerDataSource() datasource.DataSource {
	return &CsvServerDataSource{}
}

func (d *CsvServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_server"
}

func (d *CsvServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	csvAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Computed: true},
		"ufname": schema.StringAttribute{Computed: true},
		"ipv46": schema.StringAttribute{Computed: true},
		"port": schema.Int64Attribute{Computed: true},
		"type": schema.StringAttribute{Computed: true},
		"policies": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"certificate": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"customer": schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS CSV servers. Set `name` or `id` to fetch a single CSV server, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact CSV server name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific CSV server to look up.",
			},
			"ufname": schema.StringAttribute{
				Computed:    true,
				Description: "UF name.",
			},
			"ipv46": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 address.",
			},
			"port": schema.Int64Attribute{
				Computed:    true,
				Description: "Port number.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "CSV server type.",
			},
			"policies": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bound CS policies.",
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bound certificates.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"csv_servers": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All CSV servers (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: csvAttrs},
			},
		},
	}
}

func (d *CsvServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *CsvServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CsvServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item csvServerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading CSV server", err.Error())
			return
		}
		setSingleCsvServer(ctx, &config, &item, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/csvserver/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []csvServerDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading CSV servers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *csvServerDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("CSV server not found",
				fmt.Sprintf("No CSV server with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCsvServer(ctx, &config, match, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	emptyPolicies, dEmptyP := types.ListValueFrom(ctx, types.StringType, []string{})
	resp.Diagnostics.Append(dEmptyP...)
	emptyCert, dEmptyC := types.ListValueFrom(ctx, types.StringType, []string{})
	resp.Diagnostics.Append(dEmptyC...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := CsvServerDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Ufname:       types.StringNull(),
		Ipv46:        types.StringNull(),
		Port:         types.Int64Null(),
		Type:         types.StringNull(),
		Policies:     emptyPolicies,
		Certificate:  emptyCert,
		Customer:     types.StringNull(),
		Loadbalancer: types.StringNull(),
		CsvServers:   make([]CsvServerModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		m, diags := csvServerItemToModel(ctx, &page.Results[i])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.CsvServers = append(state.CsvServers, m)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCsvServer(ctx context.Context, state *CsvServerDataSourceModel, item *csvServerDSAPIModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	state.Ufname = types.StringValue(item.Ufname)
	if item.Ipv46 != nil {
		state.Ipv46 = types.StringValue(*item.Ipv46)
	} else {
		state.Ipv46 = types.StringNull()
	}
	if item.Port != nil {
		state.Port = types.Int64Value(*item.Port)
	} else {
		state.Port = types.Int64Null()
	}
	state.Type = types.StringValue(item.Type)
	pl := item.Policies
	if pl == nil {
		pl = []string{}
	}
	policiesList, d := types.ListValueFrom(ctx, types.StringType, pl)
	diags.Append(d...)
	state.Policies = policiesList
	cert := item.Certificate
	if cert == nil {
		cert = []string{}
	}
	certList, d2 := types.ListValueFrom(ctx, types.StringType, cert)
	diags.Append(d2...)
	state.Certificate = certList
	if item.Customer != nil {
		state.Customer = types.StringValue(*item.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		state.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		state.Loadbalancer = types.StringNull()
	}
	state.CsvServers = []CsvServerModel{}
}

func csvServerItemToModel(ctx context.Context, item *csvServerDSAPIModel) (CsvServerModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := CsvServerModel{
		Id:     types.StringValue(item.Id),
		Name:   types.StringValue(item.Name),
		Ufname: types.StringValue(item.Ufname),
		Type:   types.StringValue(item.Type),
	}
	if item.Ipv46 != nil {
		m.Ipv46 = types.StringValue(*item.Ipv46)
	} else {
		m.Ipv46 = types.StringNull()
	}
	if item.Port != nil {
		m.Port = types.Int64Value(*item.Port)
	} else {
		m.Port = types.Int64Null()
	}
	pl := item.Policies
	if pl == nil {
		pl = []string{}
	}
	policiesList, d := types.ListValueFrom(ctx, types.StringType, pl)
	diags.Append(d...)
	m.Policies = policiesList
	cert := item.Certificate
	if cert == nil {
		cert = []string{}
	}
	certList, d2 := types.ListValueFrom(ctx, types.StringType, cert)
	diags.Append(d2...)
	m.Certificate = certList
	if item.Customer != nil {
		m.Customer = types.StringValue(*item.Customer)
	} else {
		m.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		m.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		m.Loadbalancer = types.StringNull()
	}
	return m, diags
}
