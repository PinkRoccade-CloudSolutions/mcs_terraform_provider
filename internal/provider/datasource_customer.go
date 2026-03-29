package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &CustomerDataSource{}

type CustomerDataSource struct {
	client *apiclient.Client
}

type CustomerDataSourceModel struct {
	Id            types.String        `tfsdk:"id"`
	Name          types.String        `tfsdk:"name"`
	ContractId    types.String        `tfsdk:"contractid"`
	AdminContacts types.List          `tfsdk:"admin_contacts"`
	TechContacts  types.List          `tfsdk:"tech_contacts"`
	Customers     []CustomerListModel `tfsdk:"customers"`
}

type CustomerListModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ContractId    types.String `tfsdk:"contractid"`
	AdminContacts types.List   `tfsdk:"admin_contacts"`
	TechContacts  types.List   `tfsdk:"tech_contacts"`
}

type customerDSAPIModel struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	ContractId    string `json:"contractid,omitempty"`
	AdminContacts []int  `json:"admin_contacts"`
	TechContacts  []int  `json:"tech_contacts"`
}

func NewCustomerDataSource() datasource.DataSource {
	return &CustomerDataSource{}
}

func (d *CustomerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer"
}

func (d *CustomerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"contractid": schema.StringAttribute{Computed: true},
		"admin_contacts": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
		"tech_contacts": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS customers by id, by name, or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Customer id to fetch.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by name (uses name__icontains then exact match).",
			},
			"contractid": schema.StringAttribute{Computed: true},
			"admin_contacts": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"tech_contacts": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"customers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *CustomerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CustomerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CustomerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item customerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/tenant/customers/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading customer", err.Error())
			return
		}
		setSingleCustomer(ctx, &config, &item, resp)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/tenant/customers/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []customerDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading customers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		want := config.Name.ValueString()
		var match *customerDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == want {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Customer not found",
				fmt.Sprintf("No customer with exact name %q was found.", want))
			return
		}
		setSingleCustomer(ctx, &config, match, resp)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CustomerDataSourceModel{
		Id:            types.StringNull(),
		Name:          types.StringNull(),
		ContractId:    types.StringNull(),
		AdminContacts: types.ListNull(types.Int64Type),
		TechContacts:  types.ListNull(types.Int64Type),
		Customers:     make([]CustomerListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		lm, diags := customerToListModel(ctx, &page.Results[i])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Customers = append(state.Customers, lm)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCustomer(ctx context.Context, state *CustomerDataSourceModel, item *customerDSAPIModel, resp *datasource.ReadResponse) {
	state.Id = types.StringValue(item.Id)
	state.Name = types.StringValue(item.Name)
	state.ContractId = types.StringValue(item.ContractId)
	admin64 := intSliceToInt64(item.AdminContacts)
	tech64 := intSliceToInt64(item.TechContacts)
	adminList, diags := types.ListValueFrom(ctx, types.Int64Type, admin64)
	resp.Diagnostics.Append(diags...)
	techList, diags := types.ListValueFrom(ctx, types.Int64Type, tech64)
	resp.Diagnostics.Append(diags...)
	state.AdminContacts = adminList
	state.TechContacts = techList
	state.Customers = []CustomerListModel{}
}

func customerToListModel(ctx context.Context, item *customerDSAPIModel) (CustomerListModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	admin64 := intSliceToInt64(item.AdminContacts)
	tech64 := intSliceToInt64(item.TechContacts)
	adminList, d := types.ListValueFrom(ctx, types.Int64Type, admin64)
	diags.Append(d...)
	techList, d2 := types.ListValueFrom(ctx, types.Int64Type, tech64)
	diags.Append(d2...)
	return CustomerListModel{
		Id:            types.StringValue(item.Id),
		Name:          types.StringValue(item.Name),
		ContractId:    types.StringValue(item.ContractId),
		AdminContacts: adminList,
		TechContacts:  techList,
	}, diags
}

func intSliceToInt64(xs []int) []int64 {
	out := make([]int64, len(xs))
	for i, v := range xs {
		out[i] = int64(v)
	}
	return out
}
