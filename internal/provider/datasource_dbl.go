package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &DblDataSource{}

type DblDataSource struct {
	client *apiclient.Client
}

type DblDataSourceModel struct {
	IpAddress   types.String  `tfsdk:"ipaddress"`
	Id          types.String  `tfsdk:"id"`
	Timestamp   types.String  `tfsdk:"timestamp"`
	Source      types.String  `tfsdk:"source"`
	Occurrence  types.Int64   `tfsdk:"occurrence"`
	Persistent  types.Bool    `tfsdk:"persistent"`
	Hostname    types.String  `tfsdk:"hostname"`
	Dbls        []DblListModel `tfsdk:"dbls"`
}

type DblListModel struct {
	Id         types.String `tfsdk:"id"`
	IpAddress  types.String `tfsdk:"ipaddress"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Hostname   types.String `tfsdk:"hostname"`
}

func NewDblDataSource() datasource.DataSource {
	return &DblDataSource{}
}

func (d *DblDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbl"
}

func (d *DblDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	dblAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"ipaddress": schema.StringAttribute{
			Computed: true,
		},
		"timestamp":  schema.StringAttribute{Computed: true},
		"source":     schema.StringAttribute{Computed: true},
		"occurrence": schema.Int64Attribute{Computed: true},
		"persistent": schema.BoolAttribute{Computed: true},
		"hostname":   schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS DBL entries by ipaddress or list all.",
		Attributes: map[string]schema.Attribute{
			"ipaddress": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IP address to look up; omit to list all.",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timestamp":  schema.StringAttribute{Computed: true},
			"source":     schema.StringAttribute{Computed: true},
			"occurrence": schema.Int64Attribute{Computed: true},
			"persistent": schema.BoolAttribute{Computed: true},
			"hostname":   schema.StringAttribute{Computed: true},
			"dbls": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dblAttrs,
				},
			},
		},
	}
}

func (d *DblDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DblDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DblDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.IpAddress.IsNull() && config.IpAddress.ValueString() != "" {
		var item dblAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/dbl/%s/", config.IpAddress.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading dbl entry", err.Error())
			return
		}
		setSingleDbl(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	var page struct {
		Results []dblAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/dbl/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading dbl entries", err.Error())
		return
	}

	state := DblDataSourceModel{
		IpAddress:  types.StringNull(),
		Id:         types.StringNull(),
		Timestamp:  types.StringNull(),
		Source:     types.StringNull(),
		Occurrence: types.Int64Null(),
		Persistent: types.BoolNull(),
		Hostname:   types.StringNull(),
		Dbls:       make([]DblListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.Dbls = append(state.Dbls, dblToListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleDbl(state *DblDataSourceModel, item *dblAPIModel) {
	state.IpAddress = types.StringValue(item.IpAddress)
	state.Id = types.StringValue(strconv.Itoa(item.Id))
	state.Timestamp = types.StringValue(item.Timestamp)
	if item.Source != nil {
		state.Source = types.StringValue(*item.Source)
	} else {
		state.Source = types.StringNull()
	}
	state.Occurrence = types.Int64Value(int64(item.Occurrence))
	if item.Persistent != nil {
		state.Persistent = types.BoolValue(*item.Persistent)
	} else {
		state.Persistent = types.BoolNull()
	}
	state.Hostname = types.StringValue(item.Hostname)
	state.Dbls = []DblListModel{}
}

func dblToListModel(item *dblAPIModel) DblListModel {
	m := DblListModel{
		Id:         types.StringValue(strconv.Itoa(item.Id)),
		IpAddress:  types.StringValue(item.IpAddress),
		Timestamp:  types.StringValue(item.Timestamp),
		Occurrence: types.Int64Value(int64(item.Occurrence)),
		Hostname:   types.StringValue(item.Hostname),
	}
	if item.Source != nil {
		m.Source = types.StringValue(*item.Source)
	} else {
		m.Source = types.StringNull()
	}
	if item.Persistent != nil {
		m.Persistent = types.BoolValue(*item.Persistent)
	} else {
		m.Persistent = types.BoolNull()
	}
	return m
}
