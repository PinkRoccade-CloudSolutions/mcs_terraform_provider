package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &NetworkPoolDataSource{}

type NetworkPoolDataSource struct {
	client *apiclient.Client
}

type NetworkPoolDataSourceModel struct {
	Name         types.String       `tfsdk:"name"`
	Id           types.String       `tfsdk:"id"`
	Network      types.String       `tfsdk:"network"`
	Description  types.String       `tfsdk:"description"`
	Type         types.String       `tfsdk:"type"`
	Enabled      types.Bool         `tfsdk:"enabled"`
	NetworkPools []NetworkPoolModel `tfsdk:"network_pools"`
}

type NetworkPoolModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Network     types.String `tfsdk:"network"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

type networkPoolAPIModel struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Network     string `json:"network"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
}

func NewNetworkPoolDataSource() datasource.DataSource {
	return &NetworkPoolDataSource{}
}

func (d *NetworkPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networkpool"
}

func (d *NetworkPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	poolAttrs := map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"name":        schema.StringAttribute{Computed: true},
		"network":     schema.StringAttribute{Computed: true},
		"description": schema.StringAttribute{Computed: true},
		"type":        schema.StringAttribute{Computed: true, Description: "Pool type: lan, wan, or transit."},
		"enabled":     schema.BoolAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS network pools. Set `name` or `id` to fetch a single pool, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact network pool name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific network pool to look up.",
			},
			"network": schema.StringAttribute{
				Computed:    true,
				Description: "Network CIDR of the matched pool (e.g. 10.0.0.0/8).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched pool.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Pool type: lan, wan, or transit.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the pool is enabled.",
			},
			"network_pools": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All network pools (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: poolAttrs},
			},
		},
	}
}

func (d *NetworkPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NetworkPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var pool networkPoolAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/networkpools/%s/", config.Id.ValueString()), &pool)
		if err != nil {
			resp.Diagnostics.AddError("Error reading network pool", err.Error())
			return
		}
		setSingleNetworkPool(&config, &pool)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/networking/networkpools/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []networkPoolAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading network pools", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *networkPoolAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Network pool not found",
				fmt.Sprintf("No network pool with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleNetworkPool(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := NetworkPoolDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Network:      types.StringNull(),
		Description:  types.StringNull(),
		Type:         types.StringNull(),
		Enabled:      types.BoolNull(),
		NetworkPools: make([]NetworkPoolModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.NetworkPools = append(state.NetworkPools, NetworkPoolModel{
			Id:          types.StringValue(item.Id),
			Name:        types.StringValue(item.Name),
			Network:     types.StringValue(item.Network),
			Description: types.StringValue(item.Description),
			Type:        types.StringValue(item.Type),
			Enabled:     types.BoolValue(item.Enabled),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleNetworkPool(state *NetworkPoolDataSourceModel, pool *networkPoolAPIModel) {
	state.Id = types.StringValue(pool.Id)
	state.Name = types.StringValue(pool.Name)
	state.Network = types.StringValue(pool.Network)
	state.Description = types.StringValue(pool.Description)
	state.Type = types.StringValue(pool.Type)
	state.Enabled = types.BoolValue(pool.Enabled)
	state.NetworkPools = []NetworkPoolModel{}
}
