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

var _ datasource.DataSource = &LbMonitorDataSource{}

type LbMonitorDataSource struct {
	client *apiclient.Client
}

type LbMonitorDataSourceModel struct {
	Name         types.String       `tfsdk:"name"`
	Id           types.String       `tfsdk:"id"`
	Type         types.String       `tfsdk:"type"`
	Interval     types.Int64        `tfsdk:"interval"`
	Resptimeout  types.Int64        `tfsdk:"resptimeout"`
	Downtime     types.Int64        `tfsdk:"downtime"`
	Respcode     types.String       `tfsdk:"respcode"`
	Secure       types.String       `tfsdk:"secure"`
	Httprequest  types.String       `tfsdk:"httprequest"`
	Loadbalancer types.String       `tfsdk:"loadbalancer"`
	Protected    types.Bool         `tfsdk:"protected"`
	Customer     types.String       `tfsdk:"customer"`
	LbMonitors   []LbMonitorListModel `tfsdk:"lb_monitors"`
}

type LbMonitorListModel struct {
	Name         types.String `tfsdk:"name"`
	Id           types.String `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Interval     types.Int64  `tfsdk:"interval"`
	Resptimeout  types.Int64  `tfsdk:"resptimeout"`
	Downtime     types.Int64  `tfsdk:"downtime"`
	Respcode     types.String `tfsdk:"respcode"`
	Secure       types.String `tfsdk:"secure"`
	Httprequest  types.String `tfsdk:"httprequest"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
	Protected    types.Bool   `tfsdk:"protected"`
	Customer     types.String `tfsdk:"customer"`
}

type lbMonitorDSAPIModel struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Type         *string `json:"type,omitempty"`
	Interval     *int64  `json:"interval,omitempty"`
	Resptimeout  *int64  `json:"resptimeout,omitempty"`
	Downtime     *int64  `json:"downtime,omitempty"`
	Respcode     *string `json:"respcode,omitempty"`
	Secure       *string `json:"secure,omitempty"`
	Httprequest  *string `json:"httprequest,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
	Protected    *bool   `json:"protected,omitempty"`
	Customer     *string `json:"customer,omitempty"`
}

func NewLbMonitorDataSource() datasource.DataSource {
	return &LbMonitorDataSource{}
}

func (d *LbMonitorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_monitor"
}

func (d *LbMonitorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	monAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{Computed: true},
		"interval": schema.Int64Attribute{
			Computed: true,
		},
		"resptimeout": schema.Int64Attribute{
			Computed: true,
		},
		"downtime": schema.Int64Attribute{
			Computed: true,
		},
		"respcode":    schema.StringAttribute{Computed: true},
		"secure":      schema.StringAttribute{Computed: true},
		"httprequest": schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{
			Computed: true,
		},
		"protected": schema.BoolAttribute{
			Computed: true,
		},
		"customer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancer monitors. Set `name` or `id` to fetch a single monitor, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact monitor name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific monitor to look up.",
			},
			"type": schema.StringAttribute{Computed: true},
			"interval": schema.Int64Attribute{
				Computed: true,
			},
			"resptimeout": schema.Int64Attribute{
				Computed: true,
			},
			"downtime": schema.Int64Attribute{
				Computed: true,
			},
			"respcode":    schema.StringAttribute{Computed: true},
			"secure":      schema.StringAttribute{Computed: true},
			"httprequest": schema.StringAttribute{Computed: true},
			"loadbalancer": schema.StringAttribute{
				Computed: true,
			},
			"protected": schema.BoolAttribute{
				Computed: true,
			},
			"customer": schema.StringAttribute{Computed: true},
			"lb_monitors": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All monitors (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: monAttrs},
			},
		},
	}
}

func (d *LbMonitorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LbMonitorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LbMonitorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var m lbMonitorDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/monitor/%s/", config.Id.ValueString()), &m)
		if err != nil {
			resp.Diagnostics.AddError("Error reading lb_monitor", err.Error())
			return
		}
		setSingleLbMonitor(&config, &m)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/monitor/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []lbMonitorDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading lb_monitors", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *lbMonitorDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("LB monitor not found",
				fmt.Sprintf("No lb_monitor with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleLbMonitor(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := LbMonitorDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Type:         types.StringNull(),
		Interval:     types.Int64Null(),
		Resptimeout:  types.Int64Null(),
		Downtime:     types.Int64Null(),
		Respcode:     types.StringNull(),
		Secure:       types.StringNull(),
		Httprequest:  types.StringNull(),
		Loadbalancer: types.StringNull(),
		Protected:    types.BoolNull(),
		Customer:     types.StringNull(),
		LbMonitors:   make([]LbMonitorListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.LbMonitors = append(state.LbMonitors, toLbMonitorListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleLbMonitor(state *LbMonitorDataSourceModel, m *lbMonitorDSAPIModel) {
	state.Id = types.StringValue(m.Id)
	state.Name = types.StringValue(m.Name)
	state.Type = types.StringPointerValue(m.Type)
	if m.Interval != nil {
		state.Interval = types.Int64Value(*m.Interval)
	} else {
		state.Interval = types.Int64Null()
	}
	if m.Resptimeout != nil {
		state.Resptimeout = types.Int64Value(*m.Resptimeout)
	} else {
		state.Resptimeout = types.Int64Null()
	}
	if m.Downtime != nil {
		state.Downtime = types.Int64Value(*m.Downtime)
	} else {
		state.Downtime = types.Int64Null()
	}
	state.Respcode = types.StringPointerValue(m.Respcode)
	state.Secure = types.StringPointerValue(m.Secure)
	state.Httprequest = types.StringPointerValue(m.Httprequest)
	state.Loadbalancer = types.StringPointerValue(m.Loadbalancer)
	if m.Protected != nil {
		state.Protected = types.BoolValue(*m.Protected)
	} else {
		state.Protected = types.BoolNull()
	}
	state.Customer = types.StringPointerValue(m.Customer)
	state.LbMonitors = []LbMonitorListModel{}
}

func toLbMonitorListModel(m *lbMonitorDSAPIModel) LbMonitorListModel {
	interval := types.Int64Null()
	if m.Interval != nil {
		interval = types.Int64Value(*m.Interval)
	}
	resptimeout := types.Int64Null()
	if m.Resptimeout != nil {
		resptimeout = types.Int64Value(*m.Resptimeout)
	}
	downtime := types.Int64Null()
	if m.Downtime != nil {
		downtime = types.Int64Value(*m.Downtime)
	}
	protected := types.BoolNull()
	if m.Protected != nil {
		protected = types.BoolValue(*m.Protected)
	}
	return LbMonitorListModel{
		Id:           types.StringValue(m.Id),
		Name:         types.StringValue(m.Name),
		Type:         types.StringPointerValue(m.Type),
		Interval:     interval,
		Resptimeout:  resptimeout,
		Downtime:     downtime,
		Respcode:     types.StringPointerValue(m.Respcode),
		Secure:       types.StringPointerValue(m.Secure),
		Httprequest:  types.StringPointerValue(m.Httprequest),
		Loadbalancer: types.StringPointerValue(m.Loadbalancer),
		Protected:    protected,
		Customer:     types.StringPointerValue(m.Customer),
	}
}
