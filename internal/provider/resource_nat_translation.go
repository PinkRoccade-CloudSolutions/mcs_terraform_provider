package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &NATTranslationResource{}

type NATTranslationResource struct {
	client *apiclient.Client
}

type NATTranslationResourceModel struct {
	Id              types.String `tfsdk:"id"`
	PublicIP        types.String `tfsdk:"public_ip"`
	Interface       types.String `tfsdk:"interface"`
	Firewall        types.String `tfsdk:"firewall"`
	Translation     types.String `tfsdk:"translation"`
	PrivateIP       types.String `tfsdk:"private_ip"`
	TranslationType types.String `tfsdk:"translation_type"`
	PublicPort      types.Int64  `tfsdk:"public_port"`
	PrivatePort     types.Int64  `tfsdk:"private_port"`
	Protocol        types.String `tfsdk:"protocol"`
	Customer        types.String `tfsdk:"customer"`
	Description     types.String `tfsdk:"description"`
	State           types.String `tfsdk:"state"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

type natTranslationAPIModel struct {
	Id              string `json:"id,omitempty"`
	PublicIP        string `json:"public_ip"`
	Interface       string `json:"interface"`
	Firewall        string `json:"firewall"`
	Translation     string `json:"translation,omitempty"`
	PrivateIP       string `json:"private_ip,omitempty"`
	TranslationType string `json:"translation_type,omitempty"`
	PublicPort      *int64 `json:"public_port"`
	PrivatePort     *int64 `json:"private_port"`
	Protocol        string `json:"protocol,omitempty"`
	Customer        string `json:"customer"`
	Description     string `json:"description,omitempty"`
	State           string `json:"state,omitempty"`
	Enabled         bool   `json:"enabled"`
}

func NewNATTranslationResource() resource.Resource {
	return &NATTranslationResource{}
}

func (r *NATTranslationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_translation"
}

func (r *NATTranslationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a NAT translation in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the NAT translation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"public_ip": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the public IP address.",
			},
			"interface": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the private interface.",
			},
			"firewall": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the firewall.",
			},
			"translation": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable translation summary (read-only).",
			},
			"private_ip": schema.StringAttribute{
				Computed:    true,
				Description: "Resolved private IP address (read-only).",
			},
			"translation_type": schema.StringAttribute{
				Computed:    true,
				Description: "Translation type: one_to_one or port_forward (read-only, determined by the API).",
			},
			"public_port": schema.Int64Attribute{
				Optional:    true,
				Description: "Public port (required for port_forward).",
			},
			"private_port": schema.Int64Attribute{
				Optional:    true,
				Description: "Private port (required for port_forward).",
			},
			"protocol": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Protocol: tcp or udp.",
			},
			"customer": schema.StringAttribute{
				Required:    true,
				Description: "Customer identifier.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the NAT translation.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Sync state: synced, unsynced, error, or deleted.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the NAT translation is enabled.",
			},
		},
	}
}

func (r *NATTranslationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NATTranslationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NATTranslationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildNATAPIRequest(&plan)

	var apiResp natTranslationAPIModel
	err := r.client.Post(ctx, "/api/networking/nattranslations/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating NAT translation", err.Error())
		return
	}

	mapNATToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NATTranslationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NATTranslationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp natTranslationAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/networking/nattranslations/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading NAT translation", err.Error())
		return
	}

	mapNATToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NATTranslationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NATTranslationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildNATAPIRequest(&plan)

	var apiResp natTranslationAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/networking/nattranslations/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating NAT translation", err.Error())
		return
	}

	mapNATToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NATTranslationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NATTranslationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/nattranslations/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting NAT translation", err.Error())
	}
}

func buildNATAPIRequest(plan *NATTranslationResourceModel) natTranslationAPIModel {
	apiReq := natTranslationAPIModel{
		PublicIP:  plan.PublicIP.ValueString(),
		Interface: plan.Interface.ValueString(),
		Firewall:  plan.Firewall.ValueString(),
		Customer:  plan.Customer.ValueString(),
		Enabled:   true,
	}
	if !plan.PublicPort.IsNull() {
		v := plan.PublicPort.ValueInt64()
		apiReq.PublicPort = &v
	}
	if !plan.PrivatePort.IsNull() {
		v := plan.PrivatePort.ValueInt64()
		apiReq.PrivatePort = &v
	}
	if !plan.Protocol.IsNull() {
		apiReq.Protocol = plan.Protocol.ValueString()
	}
	if !plan.Description.IsNull() {
		apiReq.Description = plan.Description.ValueString()
	}
	if !plan.Enabled.IsNull() {
		apiReq.Enabled = plan.Enabled.ValueBool()
	}
	return apiReq
}

func mapNATToState(state *NATTranslationResourceModel, api *natTranslationAPIModel) {
	state.Id = types.StringValue(api.Id)
	state.PublicIP = types.StringValue(api.PublicIP)
	state.Interface = types.StringValue(api.Interface)
	state.Firewall = types.StringValue(api.Firewall)
	state.Translation = types.StringValue(api.Translation)
	state.PrivateIP = types.StringValue(api.PrivateIP)
	state.TranslationType = types.StringValue(api.TranslationType)
	state.Protocol = types.StringValue(api.Protocol)
	state.Customer = types.StringValue(api.Customer)
	state.Description = types.StringValue(api.Description)
	state.State = types.StringValue(api.State)
	state.Enabled = types.BoolValue(api.Enabled)
	if api.PublicPort != nil {
		state.PublicPort = types.Int64Value(*api.PublicPort)
	} else {
		state.PublicPort = types.Int64Null()
	}
	if api.PrivatePort != nil {
		state.PrivatePort = types.Int64Value(*api.PrivatePort)
	} else {
		state.PrivatePort = types.Int64Null()
	}
}
