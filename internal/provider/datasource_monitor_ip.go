package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &MonitorIPDataSource{}

type MonitorIPDataSource struct {
	client *apiclient.Client
}

type MonitorIPDataSourceModel struct {
	Id                 types.String           `tfsdk:"id"`
	IpAddress          types.String           `tfsdk:"ipaddress"`
	Timestamp          types.String           `tfsdk:"timestamp"`
	NotifyEmail        types.String           `tfsdk:"notify_email"`
	LastCheckTimestamp types.String           `tfsdk:"last_check_timestamp"`
	Customer           types.String           `tfsdk:"customer"`
	Comment            types.String           `tfsdk:"comment"`
	MonitorIps         []MonitorIPListModel   `tfsdk:"monitor_ips"`
}

type MonitorIPListModel struct {
	Id                 types.String `tfsdk:"id"`
	IpAddress          types.String `tfsdk:"ipaddress"`
	Timestamp          types.String `tfsdk:"timestamp"`
	NotifyEmail        types.String `tfsdk:"notify_email"`
	LastCheckTimestamp types.String `tfsdk:"last_check_timestamp"`
	Customer           types.String `tfsdk:"customer"`
	Comment            types.String `tfsdk:"comment"`
}

func NewMonitorIPDataSource() datasource.DataSource {
	return &MonitorIPDataSource{}
}

func (d *MonitorIPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ip"
}

func (d *MonitorIPDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true},
		"ipaddress":            schema.StringAttribute{Computed: true},
		"timestamp":            schema.StringAttribute{Computed: true},
		"notify_email":         schema.StringAttribute{Computed: true},
		"last_check_timestamp": schema.StringAttribute{Computed: true},
		"customer":             schema.StringAttribute{Computed: true},
		"comment":              schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS monitor IP entries by id or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the entry to fetch; omit to list all.",
			},
			"ipaddress":            schema.StringAttribute{Computed: true},
			"timestamp":            schema.StringAttribute{Computed: true},
			"notify_email":         schema.StringAttribute{Computed: true},
			"last_check_timestamp": schema.StringAttribute{Computed: true},
			"customer":             schema.StringAttribute{Computed: true},
			"comment":              schema.StringAttribute{Computed: true},
			"monitor_ips": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *MonitorIPDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MonitorIPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorIPDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item monitorIPAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/dbl/monitorip/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading monitor IP entry", err.Error())
			return
		}
		setSingleMonitorIP(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	var page struct {
		Results []monitorIPAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/dbl/monitorip/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading monitor IP entries", err.Error())
		return
	}

	state := MonitorIPDataSourceModel{
		Id:                 types.StringNull(),
		IpAddress:          types.StringNull(),
		Timestamp:          types.StringNull(),
		NotifyEmail:        types.StringNull(),
		LastCheckTimestamp: types.StringNull(),
		Customer:           types.StringNull(),
		Comment:            types.StringNull(),
		MonitorIps:         make([]MonitorIPListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.MonitorIps = append(state.MonitorIps, monitorIPToListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleMonitorIP(state *MonitorIPDataSourceModel, item *monitorIPAPIModel) {
	state.Id = types.StringValue(item.Id)
	state.IpAddress = types.StringValue(item.IpAddress)
	state.Timestamp = types.StringValue(item.Timestamp)
	if item.NotifyEmail != nil {
		state.NotifyEmail = types.StringValue(*item.NotifyEmail)
	} else {
		state.NotifyEmail = types.StringNull()
	}
	state.LastCheckTimestamp = types.StringValue(item.LastCheckTimestamp)
	state.Customer = types.StringValue(item.Customer)
	if item.Comment != nil {
		state.Comment = types.StringValue(*item.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	state.MonitorIps = []MonitorIPListModel{}
}

func monitorIPToListModel(item *monitorIPAPIModel) MonitorIPListModel {
	m := MonitorIPListModel{
		Id:                 types.StringValue(item.Id),
		IpAddress:          types.StringValue(item.IpAddress),
		Timestamp:          types.StringValue(item.Timestamp),
		LastCheckTimestamp: types.StringValue(item.LastCheckTimestamp),
		Customer:           types.StringValue(item.Customer),
	}
	if item.NotifyEmail != nil {
		m.NotifyEmail = types.StringValue(*item.NotifyEmail)
	} else {
		m.NotifyEmail = types.StringNull()
	}
	if item.Comment != nil {
		m.Comment = types.StringValue(*item.Comment)
	} else {
		m.Comment = types.StringNull()
	}
	return m
}
