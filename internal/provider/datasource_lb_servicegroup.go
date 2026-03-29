package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &LbServicegroupDataSource{}

type LbServicegroupDataSource struct {
	client *apiclient.Client
}

type LbServicegroupDataSourceModel struct {
	Name            types.String              `tfsdk:"name"`
	Id              types.String              `tfsdk:"id"`
	Type            types.String              `tfsdk:"type"`
	State           types.String              `tfsdk:"state"`
	Members         types.List                `tfsdk:"members"`
	Healthmonitor   types.String              `tfsdk:"healthmonitor"`
	Customer        types.String              `tfsdk:"customer"`
	Loadbalancer    types.String              `tfsdk:"loadbalancer"`
	LbServicegroups []LbServicegroupListModel `tfsdk:"lb_servicegroups"`
}

type LbServicegroupListModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	State         types.String `tfsdk:"state"`
	Members       types.List   `tfsdk:"members"`
	Healthmonitor types.String `tfsdk:"healthmonitor"`
	Customer      types.String `tfsdk:"customer"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupDSAPIModel struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	State         *string  `json:"state,omitempty"`
	Members       []string `json:"members,omitempty"`
	Healthmonitor *string  `json:"healthmonitor,omitempty"`
	Customer      *string  `json:"customer,omitempty"`
	Loadbalancer  *string  `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupDataSource() datasource.DataSource {
	return &LbServicegroupDataSource{}
}

func (d *LbServicegroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup"
}

func (d *LbServicegroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	sgAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{Computed: true},
		"state": schema.StringAttribute{
			Computed: true,
		},
		"members": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"healthmonitor": schema.StringAttribute{Computed: true},
		"customer":      schema.StringAttribute{Computed: true},
		"loadbalancer":  schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancer service groups. Set `name` or `id` to fetch a single service group, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact service group name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific service group to look up.",
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"members": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"healthmonitor": schema.StringAttribute{Computed: true},
			"customer":      schema.StringAttribute{Computed: true},
			"loadbalancer":  schema.StringAttribute{Computed: true},
			"lb_servicegroups": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All service groups (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: sgAttrs},
			},
		},
	}
}

func (d *LbServicegroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LbServicegroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LbServicegroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var sg lbServicegroupDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", config.Id.ValueString()), &sg)
		if err != nil {
			resp.Diagnostics.AddError("Error reading lb_servicegroup", err.Error())
			return
		}
		setSingleLbServicegroup(ctx, &config, &sg, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/lbservicegroup/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []lbServicegroupDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading lb_servicegroups", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *lbServicegroupDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("LB service group not found",
				fmt.Sprintf("No lb_servicegroup with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleLbServicegroup(ctx, &config, match, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := LbServicegroupDataSourceModel{
		Name:            types.StringNull(),
		Id:              types.StringNull(),
		Type:            types.StringNull(),
		State:           types.StringNull(),
		Members:         types.ListNull(types.StringType),
		Healthmonitor:   types.StringNull(),
		Customer:        types.StringNull(),
		Loadbalancer:    types.StringNull(),
		LbServicegroups: make([]LbServicegroupListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		lm, diags := toLbServicegroupListModel(ctx, &page.Results[i])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.LbServicegroups = append(state.LbServicegroups, lm)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleLbServicegroup(ctx context.Context, state *LbServicegroupDataSourceModel, sg *lbServicegroupDSAPIModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(sg.Id)
	state.Name = types.StringValue(sg.Name)
	state.Type = types.StringValue(sg.Type)
	state.State = types.StringPointerValue(sg.State)
	members := sg.Members
	if members == nil {
		members = []string{}
	}
	listVal, d := types.ListValueFrom(ctx, types.StringType, members)
	diags.Append(d...)
	state.Members = listVal
	state.Healthmonitor = types.StringPointerValue(sg.Healthmonitor)
	state.Customer = types.StringPointerValue(sg.Customer)
	state.Loadbalancer = types.StringPointerValue(sg.Loadbalancer)
	state.LbServicegroups = []LbServicegroupListModel{}
}

func toLbServicegroupListModel(ctx context.Context, sg *lbServicegroupDSAPIModel) (LbServicegroupListModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	members := sg.Members
	if members == nil {
		members = []string{}
	}
	listVal, d := types.ListValueFrom(ctx, types.StringType, members)
	diags.Append(d...)
	return LbServicegroupListModel{
		Id:            types.StringValue(sg.Id),
		Name:          types.StringValue(sg.Name),
		Type:          types.StringValue(sg.Type),
		State:         types.StringPointerValue(sg.State),
		Members:       listVal,
		Healthmonitor: types.StringPointerValue(sg.Healthmonitor),
		Customer:      types.StringPointerValue(sg.Customer),
		Loadbalancer:  types.StringPointerValue(sg.Loadbalancer),
	}, diags
}
