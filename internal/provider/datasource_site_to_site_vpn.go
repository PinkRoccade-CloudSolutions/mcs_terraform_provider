package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SiteToSiteVPNDataSource{}

type SiteToSiteVPNDataSource struct {
	client *apiclient.Client
}

type SiteToSiteVPNDataSourceModel struct {
	Id                 types.String             `tfsdk:"id"`
	Name               types.String             `tfsdk:"name"`
	Uuid               types.String             `tfsdk:"uuid"`
	State              types.String             `tfsdk:"state"`
	LastStatus         types.String             `tfsdk:"last_status"`
	Resets             types.Int64              `tfsdk:"resets"`
	LastCheck          types.String             `tfsdk:"last_check"`
	LastReset          types.String             `tfsdk:"last_reset"`
	CreatedAtTimestamp types.String             `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String             `tfsdk:"updated_at_timestamp"`
	Vpns               []SiteToSiteVPNListModel `tfsdk:"vpns"`
}

type SiteToSiteVPNListModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Uuid               types.String `tfsdk:"uuid"`
	State              types.String `tfsdk:"state"`
	LastStatus         types.String `tfsdk:"last_status"`
	Resets             types.Int64  `tfsdk:"resets"`
	LastCheck          types.String `tfsdk:"last_check"`
	LastReset          types.String `tfsdk:"last_reset"`
	CreatedAtTimestamp types.String `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String `tfsdk:"updated_at_timestamp"`
}

type siteToSiteVPNDSAPIModel struct {
	Id                 int    `json:"id"`
	Uuid               string `json:"uuid"`
	Name               string `json:"name"`
	State              string `json:"state"`
	LastStatus         string `json:"last_status"`
	Resets             int64  `json:"resets"`
	LastCheck          string `json:"last_check"`
	LastReset          string `json:"last_reset"`
	CreatedAtTimestamp string `json:"created_at_timestamp"`
	UpdatedAtTimestamp string `json:"updated_at_timestamp"`
}

func NewSiteToSiteVPNDataSource() datasource.DataSource {
	return &SiteToSiteVPNDataSource{}
}

func (d *SiteToSiteVPNDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_to_site_vpn"
}

func (d *SiteToSiteVPNDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	vpnAttrs := map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true},
		"name":                 schema.StringAttribute{Computed: true},
		"uuid":                 schema.StringAttribute{Computed: true},
		"state":                schema.StringAttribute{Computed: true},
		"last_status":          schema.StringAttribute{Computed: true},
		"resets":               schema.Int64Attribute{Computed: true},
		"last_check":           schema.StringAttribute{Computed: true},
		"last_reset":           schema.StringAttribute{Computed: true},
		"created_at_timestamp": schema.StringAttribute{Computed: true},
		"updated_at_timestamp": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS site-to-site VPNs. Set `name` or `id` to fetch a single VPN, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact VPN name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Numeric API id of a specific VPN to look up, or the id of the matched VPN when filtering by name.",
			},
			"uuid": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"last_status": schema.StringAttribute{
				Computed: true,
			},
			"resets": schema.Int64Attribute{
				Computed: true,
			},
			"last_check": schema.StringAttribute{
				Computed: true,
			},
			"last_reset": schema.StringAttribute{
				Computed: true,
			},
			"created_at_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"vpns": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All site-to-site VPNs (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: vpnAttrs},
			},
		},
	}
}

func (d *SiteToSiteVPNDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SiteToSiteVPNDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SiteToSiteVPNDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var vpn siteToSiteVPNDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", config.Id.ValueString()), &vpn)
		if err != nil {
			resp.Diagnostics.AddError("Error reading site-to-site VPN", err.Error())
			return
		}
		setSingleSiteToSiteVPN(&config, &vpn)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/vpn/site_to_site/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []siteToSiteVPNDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading site-to-site VPNs", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *siteToSiteVPNDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Site-to-site VPN not found",
				fmt.Sprintf("No site-to-site VPN with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleSiteToSiteVPN(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := SiteToSiteVPNDataSourceModel{
		Id:                 types.StringNull(),
		Name:               types.StringNull(),
		Uuid:               types.StringNull(),
		State:              types.StringNull(),
		LastStatus:         types.StringNull(),
		Resets:             types.Int64Null(),
		LastCheck:          types.StringNull(),
		LastReset:          types.StringNull(),
		CreatedAtTimestamp: types.StringNull(),
		UpdatedAtTimestamp: types.StringNull(),
		Vpns:               make([]SiteToSiteVPNListModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Vpns = append(state.Vpns, toSiteToSiteVPNListModel(&item))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleSiteToSiteVPN(state *SiteToSiteVPNDataSourceModel, vpn *siteToSiteVPNDSAPIModel) {
	state.Id = types.StringValue(strconv.Itoa(vpn.Id))
	state.Name = types.StringValue(vpn.Name)
	state.Uuid = types.StringValue(vpn.Uuid)
	state.State = types.StringValue(vpn.State)
	state.LastStatus = types.StringValue(vpn.LastStatus)
	state.Resets = types.Int64Value(vpn.Resets)
	state.LastCheck = types.StringValue(vpn.LastCheck)
	state.LastReset = types.StringValue(vpn.LastReset)
	state.CreatedAtTimestamp = types.StringValue(vpn.CreatedAtTimestamp)
	state.UpdatedAtTimestamp = types.StringValue(vpn.UpdatedAtTimestamp)
	state.Vpns = []SiteToSiteVPNListModel{}
}

func toSiteToSiteVPNListModel(vpn *siteToSiteVPNDSAPIModel) SiteToSiteVPNListModel {
	return SiteToSiteVPNListModel{
		Id:                 types.StringValue(strconv.Itoa(vpn.Id)),
		Name:               types.StringValue(vpn.Name),
		Uuid:               types.StringValue(vpn.Uuid),
		State:              types.StringValue(vpn.State),
		LastStatus:         types.StringValue(vpn.LastStatus),
		Resets:             types.Int64Value(vpn.Resets),
		LastCheck:          types.StringValue(vpn.LastCheck),
		LastReset:          types.StringValue(vpn.LastReset),
		CreatedAtTimestamp: types.StringValue(vpn.CreatedAtTimestamp),
		UpdatedAtTimestamp: types.StringValue(vpn.UpdatedAtTimestamp),
	}
}
