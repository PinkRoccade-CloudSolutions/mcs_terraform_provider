package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &LbServicegroupMemberResource{}

type LbServicegroupMemberResource struct {
	client *apiclient.Client
}

type LbServicegroupMemberResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	Servername   types.String `tfsdk:"servername"`
	Weight       types.Int64  `tfsdk:"weight"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupMemberAPIModel struct {
	Id           string  `json:"id,omitempty"`
	Address      string  `json:"address"`
	Port         *int64  `json:"port,omitempty"`
	Servername   string  `json:"servername"`
	Weight       *int64  `json:"weight,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupMemberResource() resource.Resource {
	return &LbServicegroupMemberResource{}
}

func (r *LbServicegroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup_member"
}

func (r *LbServicegroupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Required: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"servername": schema.StringAttribute{
				Required: true,
			},
			"weight": schema.Int64Attribute{
				Optional: true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *LbServicegroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *LbServicegroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbServicegroupMemberAPIModel{
		Address:    plan.Address.ValueString(),
		Servername: plan.Servername.ValueString(),
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Weight.IsNull() {
		v := plan.Weight.ValueInt64()
		apiModel.Weight = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbservicegroupmember/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_servicegroup_member", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Address = types.StringValue(apiResp.Address)
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Servername = types.StringValue(apiResp.Servername)
	plan.Weight = types.Int64PointerValue(apiResp.Weight)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lb_servicegroup_member", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Address = types.StringValue(apiResp.Address)
	state.Port = types.Int64PointerValue(apiResp.Port)
	state.Servername = types.StringValue(apiResp.Servername)
	state.Weight = types.Int64PointerValue(apiResp.Weight)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbServicegroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbServicegroupMemberAPIModel{
		Address:    plan.Address.ValueString(),
		Servername: plan.Servername.ValueString(),
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Weight.IsNull() {
		v := plan.Weight.ValueInt64()
		apiModel.Weight = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbServicegroupMemberAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_servicegroup_member", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Address = types.StringValue(apiResp.Address)
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Servername = types.StringValue(apiResp.Servername)
	plan.Weight = types.Int64PointerValue(apiResp.Weight)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbServicegroupMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroupmember/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lb_servicegroup_member", err.Error())
	}
}
