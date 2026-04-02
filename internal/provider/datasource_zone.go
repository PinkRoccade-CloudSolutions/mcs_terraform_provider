package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &ZoneDataSource{}

type ZoneDataSource struct {
	client *apiclient.Client
}

type ZoneDataSourceModel struct {
	Name          types.String          `tfsdk:"name"`
	Uuid          types.String          `tfsdk:"uuid"`
	Description   types.String          `tfsdk:"description"`
	Adom          types.String          `tfsdk:"adom"`
	TransitVrf    types.String          `tfsdk:"transit_vrf"`
	Loadbalancers []ZoneLoadbalancerModel `tfsdk:"loadbalancers"`
	Zones         []ZoneModel           `tfsdk:"zones"`
}

type ZoneModel struct {
	Uuid          types.String          `tfsdk:"uuid"`
	Name          types.String          `tfsdk:"name"`
	Description   types.String          `tfsdk:"description"`
	Adom          types.String          `tfsdk:"adom"`
	TransitVrf    types.String          `tfsdk:"transit_vrf"`
	Loadbalancers []ZoneLoadbalancerModel `tfsdk:"loadbalancers"`
}

type ZoneLoadbalancerModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type zoneAPIModel struct {
	Uuid          string                    `json:"uuid"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	Adom          string                    `json:"adom"`
	TransitVrf    string                    `json:"transit_vrf"`
	Loadbalancers []zoneLoadbalancerAPIModel `json:"loadbalancers"`
}

type zoneLoadbalancerAPIModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

func (d *ZoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	loadbalancerAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
	}

	zoneAttrs := map[string]schema.Attribute{
		"uuid": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"adom": schema.StringAttribute{
			Computed: true,
		},
		"transit_vrf": schema.StringAttribute{
			Computed: true,
		},
		"loadbalancers": schema.ListNestedAttribute{
			Computed:    true,
			Description: "Load balancers available in this zone.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: loadbalancerAttrs,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS zones. Set `name` to fetch a single zone by exact name, or omit it to list all zones.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact zone name to look up. When set, the data source returns a single zone and the `zones` list is empty.",
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the matched zone (only set when `name` is provided).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched zone (only set when `name` is provided).",
			},
			"adom": schema.StringAttribute{
				Computed:    true,
				Description: "ADOM of the matched zone (only set when `name` is provided).",
			},
			"transit_vrf": schema.StringAttribute{
				Computed:    true,
				Description: "Transit VRF of the matched zone (only set when `name` is provided).",
			},
			"loadbalancers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Load balancers available in the matched zone (only set when `name` is provided).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadbalancerAttrs,
				},
			},
			"zones": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All zones (populated when `name` is not set).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: zoneAttrs,
				},
			},
		},
	}
}

func (d *ZoneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ZoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/api/networking/zones/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__contains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []zoneAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading zones", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *zoneAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Zone not found",
				fmt.Sprintf("No zone with exact name %q was found.", config.Name.ValueString()))
			return
		}
		config.Uuid = types.StringValue(match.Uuid)
		config.Description = types.StringValue(match.Description)
		config.Adom = types.StringValue(match.Adom)
		config.TransitVrf = types.StringValue(match.TransitVrf)
		config.Loadbalancers = mapLoadbalancers(match.Loadbalancers)
		config.Zones = []ZoneModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := ZoneDataSourceModel{
		Name:          types.StringNull(),
		Uuid:          types.StringNull(),
		Description:   types.StringNull(),
		Adom:          types.StringNull(),
		TransitVrf:    types.StringNull(),
		Loadbalancers: []ZoneLoadbalancerModel{},
		Zones:         make([]ZoneModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Zones = append(state.Zones, ZoneModel{
			Uuid:          types.StringValue(item.Uuid),
			Name:          types.StringValue(item.Name),
			Description:   types.StringValue(item.Description),
			Adom:          types.StringValue(item.Adom),
			TransitVrf:    types.StringValue(item.TransitVrf),
			Loadbalancers: mapLoadbalancers(item.Loadbalancers),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapLoadbalancers(lbs []zoneLoadbalancerAPIModel) []ZoneLoadbalancerModel {
	out := make([]ZoneLoadbalancerModel, 0, len(lbs))
	for _, lb := range lbs {
		out = append(out, ZoneLoadbalancerModel{
			Id:   types.StringValue(lb.Id),
			Name: types.StringValue(lb.Name),
		})
	}
	return out
}
