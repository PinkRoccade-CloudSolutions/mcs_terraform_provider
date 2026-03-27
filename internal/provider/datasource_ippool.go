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

var _ datasource.DataSource = &IPPoolDataSource{}

type IPPoolDataSource struct {
	client *apiclient.Client
}

type IPPoolDataSourceModel struct {
	Name     types.String  `tfsdk:"name"`
	Id       types.String  `tfsdk:"id"`
	Subnet   types.String  `tfsdk:"subnet"`
	Customer types.String  `tfsdk:"customer"`
	IPPools  []IPPoolModel `tfsdk:"ip_pools"`
}

type IPPoolModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Subnet   types.String `tfsdk:"subnet"`
	Customer types.String `tfsdk:"customer"`
}

type ippoolAPIModel struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Subnet   string  `json:"subnet"`
	Customer *string `json:"customer"`
}

func NewIPPoolDataSource() datasource.DataSource {
	return &IPPoolDataSource{}
}

func (d *IPPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ippool"
}

func (d *IPPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	poolAttrs := map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true},
		"name":     schema.StringAttribute{Computed: true},
		"subnet":   schema.StringAttribute{Computed: true},
		"customer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS IP pools. Set `name` or `id` to fetch a single pool, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact IP pool name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific IP pool to look up.",
			},
			"subnet": schema.StringAttribute{
				Computed:    true,
				Description: "Subnet of the matched IP pool.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer of the matched IP pool.",
			},
			"ip_pools": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All IP pools (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: poolAttrs},
			},
		},
	}
}

func (d *IPPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IPPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var pool ippoolAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/ippools/%s/", config.Id.ValueString()), &pool)
		if err != nil {
			resp.Diagnostics.AddError("Error reading IP pool", err.Error())
			return
		}
		setSingleIPPool(&config, &pool)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/networking/ippools/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []ippoolAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading IP pools", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *ippoolAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("IP pool not found",
				fmt.Sprintf("No IP pool with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleIPPool(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := IPPoolDataSourceModel{
		Name:     types.StringNull(),
		Id:       types.StringNull(),
		Subnet:   types.StringNull(),
		Customer: types.StringNull(),
		IPPools:  make([]IPPoolModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		customer := types.StringNull()
		if item.Customer != nil {
			customer = types.StringValue(*item.Customer)
		}
		state.IPPools = append(state.IPPools, IPPoolModel{
			Id:       types.StringValue(item.Id),
			Name:     types.StringValue(item.Name),
			Subnet:   types.StringValue(item.Subnet),
			Customer: customer,
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleIPPool(state *IPPoolDataSourceModel, pool *ippoolAPIModel) {
	state.Id = types.StringValue(pool.Id)
	state.Name = types.StringValue(pool.Name)
	state.Subnet = types.StringValue(pool.Subnet)
	if pool.Customer != nil {
		state.Customer = types.StringValue(*pool.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	state.IPPools = []IPPoolModel{}
}
