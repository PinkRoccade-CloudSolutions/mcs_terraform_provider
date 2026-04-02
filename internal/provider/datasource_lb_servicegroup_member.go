package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &LbServicegroupMemberDataSource{}

type LbServicegroupMemberDataSource struct {
	client *apiclient.Client
}

type LbServicegroupMemberDataSourceModel struct {
	Id                    types.String                   `tfsdk:"id"`
	Address               types.String                   `tfsdk:"address"`
	Port                  types.Int64                    `tfsdk:"port"`
	Servername            types.String                   `tfsdk:"servername"`
	Weight                types.Int64                    `tfsdk:"weight"`
	Customer              types.String                   `tfsdk:"customer"`
	Loadbalancer          types.String                   `tfsdk:"loadbalancer"`
	LbServicegroupMembers []LbServicegroupMemberListModel `tfsdk:"lb_servicegroup_members"`
}

type LbServicegroupMemberListModel struct {
	Id           types.String `tfsdk:"id"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	Servername   types.String `tfsdk:"servername"`
	Weight       types.Int64  `tfsdk:"weight"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupMemberDSAPIModel struct {
	Id           string  `json:"id"`
	Address      string  `json:"address"`
	Port         *int64  `json:"port,omitempty"`
	Servername   string  `json:"servername"`
	Weight       *int64  `json:"weight,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupMemberDataSource() datasource.DataSource {
	return &LbServicegroupMemberDataSource{}
}

func (d *LbServicegroupMemberDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup_member"
}

func (d *LbServicegroupMemberDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	mAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"address": schema.StringAttribute{
			Computed: true,
		},
		"port": schema.Int64Attribute{
			Computed: true,
		},
		"servername": schema.StringAttribute{Computed: true},
		"weight": schema.Int64Attribute{
			Computed: true,
		},
		"customer":     schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancer service group members. Set `id` to fetch a single member, or omit it to list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific service group member to look up.",
			},
			"address": schema.StringAttribute{
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Computed: true,
			},
			"servername": schema.StringAttribute{Computed: true},
			"weight": schema.Int64Attribute{
				Computed: true,
			},
			"customer":     schema.StringAttribute{Computed: true},
			"loadbalancer": schema.StringAttribute{Computed: true},
			"lb_servicegroup_members": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All service group members (populated when `id` is not set).",
				NestedObject: schema.NestedAttributeObject{Attributes: mAttrs},
			},
		},
	}
}

func (d *LbServicegroupMemberDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LbServicegroupMemberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LbServicegroupMemberDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var m lbServicegroupMemberDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", config.Id.ValueString()), &m)
		if err != nil {
			resp.Diagnostics.AddError("Error reading lb_servicegroup_member", err.Error())
			return
		}
		setSingleLbServicegroupMember(&config, &m)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/lbservicegroupmember/?page_size=1000"
	var page struct {
		Results []lbServicegroupMemberDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading lb_servicegroup_members", err.Error())
		return
	}

	state := LbServicegroupMemberDataSourceModel{
		Id:                    types.StringNull(),
		Address:               types.StringNull(),
		Port:                  types.Int64Null(),
		Servername:            types.StringNull(),
		Weight:                types.Int64Null(),
		Customer:              types.StringNull(),
		Loadbalancer:          types.StringNull(),
		LbServicegroupMembers: make([]LbServicegroupMemberListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.LbServicegroupMembers = append(state.LbServicegroupMembers, toLbServicegroupMemberListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleLbServicegroupMember(state *LbServicegroupMemberDataSourceModel, m *lbServicegroupMemberDSAPIModel) {
	state.Id = types.StringValue(m.Id)
	state.Address = types.StringValue(m.Address)
	if m.Port != nil {
		state.Port = types.Int64Value(*m.Port)
	} else {
		state.Port = types.Int64Null()
	}
	state.Servername = types.StringValue(m.Servername)
	if m.Weight != nil {
		state.Weight = types.Int64Value(*m.Weight)
	} else {
		state.Weight = types.Int64Null()
	}
	state.Customer = types.StringPointerValue(m.Customer)
	state.Loadbalancer = types.StringPointerValue(m.Loadbalancer)
	state.LbServicegroupMembers = []LbServicegroupMemberListModel{}
}

func toLbServicegroupMemberListModel(m *lbServicegroupMemberDSAPIModel) LbServicegroupMemberListModel {
	port := types.Int64Null()
	if m.Port != nil {
		port = types.Int64Value(*m.Port)
	}
	weight := types.Int64Null()
	if m.Weight != nil {
		weight = types.Int64Value(*m.Weight)
	}
	return LbServicegroupMemberListModel{
		Id:           types.StringValue(m.Id),
		Address:      types.StringValue(m.Address),
		Port:         port,
		Servername:   types.StringValue(m.Servername),
		Weight:       weight,
		Customer:     types.StringPointerValue(m.Customer),
		Loadbalancer: types.StringPointerValue(m.Loadbalancer),
	}
}
