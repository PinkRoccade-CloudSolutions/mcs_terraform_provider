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

var _ datasource.DataSource = &CsActionDataSource{}

type CsActionDataSource struct {
	client *apiclient.Client
}

type CsActionDataSourceModel struct {
	Name        types.String      `tfsdk:"name"`
	Id          types.String      `tfsdk:"id"`
	Lbvserver   types.String      `tfsdk:"lbvserver"`
	Customer    types.String      `tfsdk:"customer"`
	Loadbalancer types.String     `tfsdk:"loadbalancer"`
	CsActions   []CsActionModel   `tfsdk:"cs_actions"`
}

type CsActionModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Lbvserver    types.String `tfsdk:"lbvserver"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csActionDSAPIModel struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Lbvserver    *string `json:"lbvserver,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewCsActionDataSource() datasource.DataSource {
	return &CsActionDataSource{}
}

func (d *CsActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cs_action"
}

func (d *CsActionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	actionAttrs := map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true},
		"name":         schema.StringAttribute{Computed: true},
		"lbvserver":    schema.StringAttribute{Computed: true},
		"customer":     schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS CS actions. Set `name` or `id` to fetch a single CS action, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact CS action name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific CS action to look up.",
			},
			"lbvserver": schema.StringAttribute{
				Computed:    true,
				Description: "Associated LB vserver.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"cs_actions": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All CS actions (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: actionAttrs},
			},
		},
	}
}

func (d *CsActionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CsActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CsActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item csActionDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csaction/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading CS action", err.Error())
			return
		}
		setSingleCsAction(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/csaction/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []csActionDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading CS actions", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *csActionDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("CS action not found",
				fmt.Sprintf("No CS action with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCsAction(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CsActionDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Lbvserver:    types.StringNull(),
		Customer:     types.StringNull(),
		Loadbalancer: types.StringNull(),
		CsActions:    make([]CsActionModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.CsActions = append(state.CsActions, csActionItemToModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCsAction(state *CsActionDataSourceModel, item *csActionDSAPIModel) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	if item.Lbvserver != nil {
		state.Lbvserver = types.StringValue(*item.Lbvserver)
	} else {
		state.Lbvserver = types.StringNull()
	}
	if item.Customer != nil {
		state.Customer = types.StringValue(*item.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		state.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		state.Loadbalancer = types.StringNull()
	}
	state.CsActions = []CsActionModel{}
}

func csActionItemToModel(item *csActionDSAPIModel) CsActionModel {
	m := CsActionModel{
		Id:   types.StringValue(item.Id),
		Name: types.StringValue(item.Name),
	}
	if item.Lbvserver != nil {
		m.Lbvserver = types.StringValue(*item.Lbvserver)
	} else {
		m.Lbvserver = types.StringNull()
	}
	if item.Customer != nil {
		m.Customer = types.StringValue(*item.Customer)
	} else {
		m.Customer = types.StringNull()
	}
	if item.Loadbalancer != nil {
		m.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		m.Loadbalancer = types.StringNull()
	}
	return m
}
