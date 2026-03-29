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

var _ datasource.DataSource = &LbvServerDataSource{}

type LbvServerDataSource struct {
	client *apiclient.Client
}

type LbvServerDataSourceModel struct {
	Name         types.String       `tfsdk:"name"`
	Id           types.String       `tfsdk:"id"`
	Ipv46        types.String       `tfsdk:"ipv46"`
	Port         types.Int64        `tfsdk:"port"`
	Type         types.String       `tfsdk:"type"`
	Servicegroup types.List         `tfsdk:"servicegroup"`
	Certificate  types.List         `tfsdk:"certificate"`
	Customer     types.String       `tfsdk:"customer"`
	Loadbalancer types.String       `tfsdk:"loadbalancer"`
	LbvServers   []LbvServerListModel `tfsdk:"lbv_servers"`
}

type LbvServerListModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Ipv46        types.String `tfsdk:"ipv46"`
	Port         types.Int64  `tfsdk:"port"`
	Type         types.String `tfsdk:"type"`
	Servicegroup types.List   `tfsdk:"servicegroup"`
	Certificate  types.List   `tfsdk:"certificate"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type lbvServerDSAPIModel struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Ipv46        *string  `json:"ipv46,omitempty"`
	Port         *int64   `json:"port,omitempty"`
	Type         *string  `json:"type,omitempty"`
	Servicegroup []string `json:"servicegroup"`
	Certificate  []string `json:"certificate,omitempty"`
	Customer     *string  `json:"customer,omitempty"`
	Loadbalancer *string  `json:"loadbalancer,omitempty"`
}

func NewLbvServerDataSource() datasource.DataSource {
	return &LbvServerDataSource{}
}

func (d *LbvServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lbv_server"
}

func (d *LbvServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	srvAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"ipv46": schema.StringAttribute{Computed: true},
		"port": schema.Int64Attribute{
			Computed: true,
		},
		"type": schema.StringAttribute{Computed: true},
		"servicegroup": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"certificate": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
		"customer":     schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancer virtual servers. Set `name` or `id` to fetch a single server, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact virtual server name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific lbv_server to look up.",
			},
			"ipv46": schema.StringAttribute{Computed: true},
			"port": schema.Int64Attribute{
				Computed: true,
			},
			"type": schema.StringAttribute{Computed: true},
			"servicegroup": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"customer":     schema.StringAttribute{Computed: true},
			"loadbalancer": schema.StringAttribute{Computed: true},
			"lbv_servers": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All lbv_servers (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: srvAttrs},
			},
		},
	}
}

func (d *LbvServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LbvServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LbvServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var s lbvServerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", config.Id.ValueString()), &s)
		if err != nil {
			resp.Diagnostics.AddError("Error reading lbv_server", err.Error())
			return
		}
		setSingleLbvServer(ctx, &config, &s, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/lbvserver/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []lbvServerDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading lbv_servers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *lbvServerDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("LBV server not found",
				fmt.Sprintf("No lbv_server with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleLbvServer(ctx, &config, match, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := LbvServerDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Ipv46:        types.StringNull(),
		Port:         types.Int64Null(),
		Type:         types.StringNull(),
		Servicegroup: types.ListNull(types.StringType),
		Certificate:  types.ListNull(types.StringType),
		Customer:     types.StringNull(),
		Loadbalancer: types.StringNull(),
		LbvServers:   make([]LbvServerListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		lm, diags := toLbvServerListModel(ctx, &page.Results[i])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.LbvServers = append(state.LbvServers, lm)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleLbvServer(ctx context.Context, state *LbvServerDataSourceModel, s *lbvServerDSAPIModel, diags *diag.Diagnostics) {
	state.Id = types.StringValue(s.Id)
	state.Name = types.StringValue(s.Name)
	state.Ipv46 = types.StringPointerValue(s.Ipv46)
	if s.Port != nil {
		state.Port = types.Int64Value(*s.Port)
	} else {
		state.Port = types.Int64Null()
	}
	state.Type = types.StringPointerValue(s.Type)
	sg := s.Servicegroup
	if sg == nil {
		sg = []string{}
	}
	sgList, d := types.ListValueFrom(ctx, types.StringType, sg)
	diags.Append(d...)
	state.Servicegroup = sgList
	cert := s.Certificate
	if cert == nil {
		cert = []string{}
	}
	certList, d2 := types.ListValueFrom(ctx, types.StringType, cert)
	diags.Append(d2...)
	state.Certificate = certList
	state.Customer = types.StringPointerValue(s.Customer)
	state.Loadbalancer = types.StringPointerValue(s.Loadbalancer)
	state.LbvServers = []LbvServerListModel{}
}

func toLbvServerListModel(ctx context.Context, s *lbvServerDSAPIModel) (LbvServerListModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	port := types.Int64Null()
	if s.Port != nil {
		port = types.Int64Value(*s.Port)
	}
	sg := s.Servicegroup
	if sg == nil {
		sg = []string{}
	}
	sgList, d := types.ListValueFrom(ctx, types.StringType, sg)
	diags.Append(d...)
	cert := s.Certificate
	if cert == nil {
		cert = []string{}
	}
	certList, d2 := types.ListValueFrom(ctx, types.StringType, cert)
	diags.Append(d2...)
	return LbvServerListModel{
		Id:           types.StringValue(s.Id),
		Name:         types.StringValue(s.Name),
		Ipv46:        types.StringPointerValue(s.Ipv46),
		Port:         port,
		Type:         types.StringPointerValue(s.Type),
		Servicegroup: sgList,
		Certificate:  certList,
		Customer:     types.StringPointerValue(s.Customer),
		Loadbalancer: types.StringPointerValue(s.Loadbalancer),
	}, diags
}
