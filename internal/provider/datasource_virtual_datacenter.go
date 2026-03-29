package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VirtualDatacenterDataSource{}

type VirtualDatacenterDataSource struct {
	client *apiclient.Client
}

type VirtualDatacenterDataSourceModel struct {
	Name               types.String                 `tfsdk:"name"`
	Id                 types.String                 `tfsdk:"id"`
	Customer           types.String                 `tfsdk:"customer"`
	VirtualDatacenters []VirtualDatacenterListModel `tfsdk:"virtual_datacenters"`
}

type VirtualDatacenterListModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Customer types.String `tfsdk:"customer"`
}

type virtualDatacenterDSAPIModel struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Customer string `json:"customer"`
}

func NewVirtualDatacenterDataSource() datasource.DataSource {
	return &VirtualDatacenterDataSource{}
}

func (d *VirtualDatacenterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_datacenter"
}

func (d *VirtualDatacenterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	vdcAttrs := map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true},
		"name":     schema.StringAttribute{Computed: true},
		"customer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS virtual datacenters. Set `name` or `id` to fetch a single datacenter, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact virtual datacenter name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific virtual datacenter to look up.",
			},
			"customer": schema.StringAttribute{
				Computed: true,
			},
			"virtual_datacenters": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All virtual datacenters (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: vdcAttrs},
			},
		},
	}
}

func (d *VirtualDatacenterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VirtualDatacenterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VirtualDatacenterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var vdc virtualDatacenterDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/virtualization/virtualdatacenter/%s/", config.Id.ValueString()), &vdc)
		if err != nil {
			resp.Diagnostics.AddError("Error reading virtual datacenter", err.Error())
			return
		}
		setSingleVirtualDatacenter(&config, &vdc)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/virtualization/virtualdatacenter/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__contains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []virtualDatacenterDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading virtual datacenters", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *virtualDatacenterDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Virtual datacenter not found",
				fmt.Sprintf("No virtual datacenter with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleVirtualDatacenter(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := VirtualDatacenterDataSourceModel{
		Name:               types.StringNull(),
		Id:                 types.StringNull(),
		Customer:           types.StringNull(),
		VirtualDatacenters: make([]VirtualDatacenterListModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.VirtualDatacenters = append(state.VirtualDatacenters, VirtualDatacenterListModel{
			Id:       types.StringValue(item.Id),
			Name:     types.StringValue(item.Name),
			Customer: types.StringValue(item.Customer),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleVirtualDatacenter(state *VirtualDatacenterDataSourceModel, vdc *virtualDatacenterDSAPIModel) {
	state.Id = types.StringValue(vdc.Id)
	state.Name = types.StringValue(vdc.Name)
	state.Customer = types.StringValue(vdc.Customer)
	state.VirtualDatacenters = []VirtualDatacenterListModel{}
}
