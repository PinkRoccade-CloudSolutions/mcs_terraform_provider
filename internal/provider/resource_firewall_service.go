package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &FirewallServiceResource{}

type FirewallServiceResource struct {
	client *apiclient.Client
}

type FirewallServiceModel struct {
	Id           types.String `tfsdk:"id"`
	Domain       types.String `tfsdk:"domain"`
	Name         types.String `tfsdk:"name"`
	Uuid         types.String `tfsdk:"uuid"`
	Protocol     types.String `tfsdk:"protocol"`
	Comment      types.String `tfsdk:"comment"`
	TcpPortrange types.List   `tfsdk:"tcp_portrange"`
	UdpPortrange types.List   `tfsdk:"udp_portrange"`
	Used         types.Bool   `tfsdk:"used"`
}

type firewallServiceAPI struct {
	Name         string   `json:"name,omitempty"`
	Uuid         string   `json:"uuid,omitempty"`
	Protocol     *string  `json:"protocol,omitempty"`
	Comment      *string  `json:"comment,omitempty"`
	TcpPortrange []string `json:"tcp_portrange,omitempty"`
	UdpPortrange []string `json:"udp_portrange,omitempty"`
	Used         bool     `json:"used"`
}

func NewFirewallServiceResource() resource.Resource {
	return &FirewallServiceResource{}
}

func (r *FirewallServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_service"
}

func (r *FirewallServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall service in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain this service belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall service.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"protocol": schema.StringAttribute{
				Optional:    true,
				Description: "Protocol type for the service.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Comment for the service.",
			},
			"tcp_portrange": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of TCP port ranges.",
			},
			"udp_portrange": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of UDP port ranges.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the service is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *FirewallServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	body := firewallServiceAPI{
		Name: plan.Name.ValueString(),
	}
	if !plan.Protocol.IsNull() {
		v := plan.Protocol.ValueString()
		body.Protocol = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		body.Comment = &v
	}
	if !plan.TcpPortrange.IsNull() {
		resp.Diagnostics.Append(plan.TcpPortrange.ElementsAs(ctx, &body.TcpPortrange, false)...)
	}
	if !plan.UdpPortrange.IsNull() {
		resp.Diagnostics.Append(plan.UdpPortrange.ElementsAs(ctx, &body.UdpPortrange, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallServiceAPI
	path := fmt.Sprintf("/api/networking/domain/%s/services/", domain)
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall service", err.Error())
		return
	}

	mapFirewallServiceToState(ctx, &plan, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(state.Id.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID format", "Expected domain/name")
		return
	}
	domain, name := parts[0], parts[1]

	var apiResp firewallServiceAPI
	path := fmt.Sprintf("/api/networking/domain/%s/services/%s/", domain, name)
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall service", err.Error())
		return
	}

	mapFirewallServiceToState(ctx, &state, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallServiceResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"mcs_firewall_service does not support updates, destroy and recreate the resource.",
	)
}

func (r *FirewallServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(state.Id.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID format", "Expected domain/name")
		return
	}
	domain, name := parts[0], parts[1]

	path := fmt.Sprintf("/api/networking/domain/%s/services/%s/", domain, name)
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall service", err.Error())
	}
}

func mapFirewallServiceToState(ctx context.Context, model *FirewallServiceModel, domain string, api *firewallServiceAPI, diagnostics *diag.Diagnostics) {
	model.Id = types.StringValue(domain + "/" + api.Name)
	model.Domain = types.StringValue(domain)
	model.Name = types.StringValue(api.Name)
	model.Uuid = types.StringValue(api.Uuid)
	model.Used = types.BoolValue(api.Used)

	if api.Protocol != nil {
		model.Protocol = types.StringValue(*api.Protocol)
	} else {
		model.Protocol = types.StringNull()
	}
	if api.Comment != nil {
		model.Comment = types.StringValue(*api.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	if len(api.TcpPortrange) > 0 {
		listVal, diags := types.ListValueFrom(ctx, types.StringType, api.TcpPortrange)
		diagnostics.Append(diags...)
		model.TcpPortrange = listVal
	} else {
		model.TcpPortrange = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if len(api.UdpPortrange) > 0 {
		listVal, diags := types.ListValueFrom(ctx, types.StringType, api.UdpPortrange)
		diagnostics.Append(diags...)
		model.UdpPortrange = listVal
	} else {
		model.UdpPortrange = types.ListValueMust(types.StringType, []attr.Value{})
	}
}
