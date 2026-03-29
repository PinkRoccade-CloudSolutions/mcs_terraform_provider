package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &CustomerResource{}

type CustomerResource struct {
	client *apiclient.Client
}

type CustomerResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ContractId         types.String `tfsdk:"contractid"`
	Sdm                types.Int64  `tfsdk:"sdm"`
	TechContacts  types.List `tfsdk:"tech_contacts"`
	AdminContacts types.List `tfsdk:"admin_contacts"`
}

type customerAPIModel struct {
	Id                 string  `json:"id,omitempty"`
	Name               string  `json:"name"`
	ContractId         string  `json:"contractid,omitempty"`
	Sdm                int64   `json:"sdm,omitempty"`
	TechContacts  []int64 `json:"tech_contacts"`
	AdminContacts []int64 `json:"admin_contacts"`
}

func NewCustomerResource() resource.Resource {
	return &CustomerResource{}
}

func (r *CustomerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customer"
}

func (r *CustomerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"contractid": schema.StringAttribute{
				Optional: true,
			},
			"sdm": schema.Int64Attribute{
				Optional: true,
			},
			"tech_contacts": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"admin_contacts": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
		},
	}
}

func (r *CustomerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *CustomerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := customerAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.ContractId.IsNull() {
		apiReq.ContractId = plan.ContractId.ValueString()
	}
	if !plan.Sdm.IsNull() {
		apiReq.Sdm = plan.Sdm.ValueInt64()
	}
	if !plan.TechContacts.IsNull() {
		resp.Diagnostics.Append(plan.TechContacts.ElementsAs(ctx, &apiReq.TechContacts, false)...)
	}
	if !plan.AdminContacts.IsNull() {
		resp.Diagnostics.Append(plan.AdminContacts.ElementsAs(ctx, &apiReq.AdminContacts, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Post(ctx, "/api/tenant/customers/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating customer", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)

	techList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.TechContacts)
	resp.Diagnostics.Append(diags...)
	plan.TechContacts = techList

	adminList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.AdminContacts)
	resp.Diagnostics.Append(diags...)
	plan.AdminContacts = adminList

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/tenant/customers/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading customer", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.ContractId = types.StringValue(apiResp.ContractId)
	state.Sdm = types.Int64Value(apiResp.Sdm)

	techList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.TechContacts)
	resp.Diagnostics.Append(diags...)
	state.TechContacts = techList

	adminList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.AdminContacts)
	resp.Diagnostics.Append(diags...)
	state.AdminContacts = adminList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CustomerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := customerAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.ContractId.IsNull() {
		apiReq.ContractId = plan.ContractId.ValueString()
	}
	if !plan.Sdm.IsNull() {
		apiReq.Sdm = plan.Sdm.ValueInt64()
	}
	if !plan.TechContacts.IsNull() {
		resp.Diagnostics.Append(plan.TechContacts.ElementsAs(ctx, &apiReq.TechContacts, false)...)
	}
	if !plan.AdminContacts.IsNull() {
		resp.Diagnostics.Append(plan.AdminContacts.ElementsAs(ctx, &apiReq.AdminContacts, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp customerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/tenant/customers/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating customer", err.Error())
		return
	}

	techList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.TechContacts)
	resp.Diagnostics.Append(diags...)
	plan.TechContacts = techList

	adminList, diags := types.ListValueFrom(ctx, types.Int64Type, apiResp.AdminContacts)
	resp.Diagnostics.Append(diags...)
	plan.AdminContacts = adminList

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CustomerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/tenant/customers/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting customer", err.Error())
	}
}
