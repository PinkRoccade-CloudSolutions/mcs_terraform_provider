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

var _ resource.Resource = &PublicIPAddressResource{}

type PublicIPAddressResource struct {
	client *apiclient.Client
}

type PublicIPAddressResourceModel struct {
	Id          types.String `tfsdk:"id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Pool        types.String `tfsdk:"pool"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Customer    types.String `tfsdk:"customer"`
}

type publicIPAddressAPIModel struct {
	Id          string  `json:"id,omitempty"`
	IPAddress   string  `json:"ip_address,omitempty"`
	Pool        *string `json:"pool"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status,omitempty"`
	Type        string  `json:"type,omitempty"`
	Customer    *string `json:"customer"`
}

func NewPublicIPAddressResource() resource.Resource {
	return &PublicIPAddressResource{}
}

func (r *PublicIPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip_address"
}

func (r *PublicIPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a public IP address in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the public IP address.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "The public IP address (read-only, assigned by the API).",
			},
			"pool": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the IP pool this address belongs to.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the public IP address.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type: nat, vip, or loadbalancer.",
			},
			"customer": schema.StringAttribute{
				Optional:    true,
				Description: "Customer identifier.",
			},
		},
	}
}

func (r *PublicIPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PublicIPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildPublicIPAPIRequest(&plan)

	var apiResp publicIPAddressAPIModel
	err := r.client.Post(ctx, "/api/networking/publicipaddresss/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating public IP address", err.Error())
		return
	}

	mapPublicIPToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PublicIPAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp publicIPAddressAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading public IP address", err.Error())
		return
	}

	mapPublicIPToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PublicIPAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildPublicIPAPIRequest(&plan)

	var apiResp publicIPAddressAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating public IP address", err.Error())
		return
	}

	mapPublicIPToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PublicIPAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting public IP address", err.Error())
	}
}

func buildPublicIPAPIRequest(plan *PublicIPAddressResourceModel) publicIPAddressAPIModel {
	var apiReq publicIPAddressAPIModel
	if !plan.Pool.IsNull() {
		v := plan.Pool.ValueString()
		apiReq.Pool = &v
	}
	if !plan.Description.IsNull() {
		apiReq.Description = plan.Description.ValueString()
	}
	if !plan.Type.IsNull() {
		apiReq.Type = plan.Type.ValueString()
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiReq.Customer = &v
	}
	return apiReq
}

func mapPublicIPToState(state *PublicIPAddressResourceModel, api *publicIPAddressAPIModel) {
	state.Id = types.StringValue(api.Id)
	state.IPAddress = types.StringValue(api.IPAddress)
	state.Description = types.StringValue(api.Description)
	state.Type = types.StringValue(api.Type)
	if api.Pool != nil {
		state.Pool = types.StringValue(*api.Pool)
	} else {
		state.Pool = types.StringNull()
	}
	if api.Customer != nil {
		state.Customer = types.StringValue(*api.Customer)
	} else {
		state.Customer = types.StringNull()
	}
}
