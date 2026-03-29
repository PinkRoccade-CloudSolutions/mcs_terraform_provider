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

var _ datasource.DataSource = &CsPolicyDataSource{}

type CsPolicyDataSource struct {
	client *apiclient.Client
}

type CsPolicyDataSourceModel struct {
	Name         types.String      `tfsdk:"name"`
	Id           types.String      `tfsdk:"id"`
	Action       types.String      `tfsdk:"action"`
	Expression   types.String      `tfsdk:"expression"`
	Customer     types.String      `tfsdk:"customer"`
	Application  types.String      `tfsdk:"application"`
	Loadbalancer types.String      `tfsdk:"loadbalancer"`
	CsPolicies   []CsPolicyModel   `tfsdk:"cs_policies"`
}

type CsPolicyModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Action       types.String `tfsdk:"action"`
	Expression   types.String `tfsdk:"expression"`
	Customer     types.String `tfsdk:"customer"`
	Application  types.String `tfsdk:"application"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csPolicyDSAPIModel struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Action       *string `json:"action,omitempty"`
	Expression   *string `json:"expression,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Application  *string `json:"application,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewCsPolicyDataSource() datasource.DataSource {
	return &CsPolicyDataSource{}
}

func (d *CsPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cs_policy"
}

func (d *CsPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	policyAttrs := map[string]schema.Attribute{
		"id":           schema.StringAttribute{Computed: true},
		"name":         schema.StringAttribute{Computed: true},
		"action":       schema.StringAttribute{Computed: true},
		"expression":   schema.StringAttribute{Computed: true},
		"customer":     schema.StringAttribute{Computed: true},
		"application":  schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS CS policies. Set `name` or `id` to fetch a single CS policy, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact CS policy name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific CS policy to look up.",
			},
			"action": schema.StringAttribute{
				Computed:    true,
				Description: "Associated CS action.",
			},
			"expression": schema.StringAttribute{
				Computed:    true,
				Description: "Policy expression.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"application": schema.StringAttribute{
				Computed:    true,
				Description: "Application identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"cs_policies": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All CS policies (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: policyAttrs},
			},
		},
	}
}

func (d *CsPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CsPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CsPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item csPolicyDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/cspolicy/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading CS policy", err.Error())
			return
		}
		setSingleCsPolicy(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/cspolicy/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []csPolicyDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading CS policies", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *csPolicyDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("CS policy not found",
				fmt.Sprintf("No CS policy with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCsPolicy(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CsPolicyDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Action:       types.StringNull(),
		Expression:   types.StringNull(),
		Customer:     types.StringNull(),
		Application:  types.StringNull(),
		Loadbalancer: types.StringNull(),
		CsPolicies:   make([]CsPolicyModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.CsPolicies = append(state.CsPolicies, csPolicyItemToModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCsPolicy(state *CsPolicyDataSourceModel, item *csPolicyDSAPIModel) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	if item.Action != nil {
		state.Action = types.StringValue(*item.Action)
	} else {
		state.Action = types.StringNull()
	}
	if item.Expression != nil {
		state.Expression = types.StringValue(*item.Expression)
	} else {
		state.Expression = types.StringNull()
	}
	if item.Customer != nil {
		state.Customer = types.StringValue(*item.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	if item.Application != nil {
		state.Application = types.StringValue(*item.Application)
	} else {
		state.Application = types.StringNull()
	}
	if item.Loadbalancer != nil {
		state.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		state.Loadbalancer = types.StringNull()
	}
	state.CsPolicies = []CsPolicyModel{}
}

func csPolicyItemToModel(item *csPolicyDSAPIModel) CsPolicyModel {
	m := CsPolicyModel{
		Id:   types.StringValue(item.Id),
		Name: types.StringValue(item.Name),
	}
	if item.Action != nil {
		m.Action = types.StringValue(*item.Action)
	} else {
		m.Action = types.StringNull()
	}
	if item.Expression != nil {
		m.Expression = types.StringValue(*item.Expression)
	} else {
		m.Expression = types.StringNull()
	}
	if item.Customer != nil {
		m.Customer = types.StringValue(*item.Customer)
	} else {
		m.Customer = types.StringNull()
	}
	if item.Application != nil {
		m.Application = types.StringValue(*item.Application)
	} else {
		m.Application = types.StringNull()
	}
	if item.Loadbalancer != nil {
		m.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		m.Loadbalancer = types.StringNull()
	}
	return m
}
