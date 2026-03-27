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

var _ datasource.DataSource = &FirewallDataSource{}

type FirewallDataSource struct {
	client *apiclient.Client
}

type FirewallDataSourceModel struct {
	Name        types.String    `tfsdk:"name"`
	Id          types.String    `tfsdk:"id"`
	Description types.String    `tfsdk:"description"`
	Customer    types.String    `tfsdk:"customer"`
	Type        types.String    `tfsdk:"type"`
	Firewalls   []FirewallModel `tfsdk:"firewalls"`
}

type FirewallModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Customer    types.String `tfsdk:"customer"`
	Type        types.String `tfsdk:"type"`
}

type firewallAPIModel struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Customer    string `json:"customer"`
	Type        string `json:"type"`
}

func NewFirewallDataSource() datasource.DataSource {
	return &FirewallDataSource{}
}

func (d *FirewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (d *FirewallDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	fwAttrs := map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"name":        schema.StringAttribute{Computed: true},
		"description": schema.StringAttribute{Computed: true},
		"customer":    schema.StringAttribute{Computed: true},
		"type":        schema.StringAttribute{Computed: true, Description: "Firewall type: internet or wan."},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS firewalls. Set `name` or `id` to fetch a single firewall, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact firewall name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific firewall to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched firewall.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer of the matched firewall.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Firewall type: internet or wan.",
			},
			"firewalls": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All firewalls (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: fwAttrs},
			},
		},
	}
}

func (d *FirewallDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var fw firewallAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/firewalls/%s/", config.Id.ValueString()), &fw)
		if err != nil {
			resp.Diagnostics.AddError("Error reading firewall", err.Error())
			return
		}
		setSingleFirewall(&config, &fw)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/networking/firewalls/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []firewallAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading firewalls", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *firewallAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Firewall not found",
				fmt.Sprintf("No firewall with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleFirewall(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := FirewallDataSourceModel{
		Name:        types.StringNull(),
		Id:          types.StringNull(),
		Description: types.StringNull(),
		Customer:    types.StringNull(),
		Type:        types.StringNull(),
		Firewalls:   make([]FirewallModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Firewalls = append(state.Firewalls, FirewallModel{
			Id:          types.StringValue(item.Id),
			Name:        types.StringValue(item.Name),
			Description: types.StringValue(item.Description),
			Customer:    types.StringValue(item.Customer),
			Type:        types.StringValue(item.Type),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleFirewall(state *FirewallDataSourceModel, fw *firewallAPIModel) {
	state.Id = types.StringValue(fw.Id)
	state.Name = types.StringValue(fw.Name)
	state.Description = types.StringValue(fw.Description)
	state.Customer = types.StringValue(fw.Customer)
	state.Type = types.StringValue(fw.Type)
	state.Firewalls = []FirewallModel{}
}
