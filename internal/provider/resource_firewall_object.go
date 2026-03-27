package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &FirewallObjectResource{}

type FirewallObjectResource struct {
	client *apiclient.Client
}

type FirewallObjectModel struct {
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Uuid    types.String `tfsdk:"uuid"`
	Address types.String `tfsdk:"address"`
	Subnet  types.String `tfsdk:"subnet"`
	Comment types.String `tfsdk:"comment"`
	Used    types.Bool   `tfsdk:"used"`
}

type firewallObjectAPI struct {
	Name    string  `json:"name,omitempty"`
	Uuid    string  `json:"uuid,omitempty"`
	Address *string `json:"address,omitempty"`
	Subnet  *string `json:"subnet,omitempty"`
	Comment *string `json:"comment,omitempty"`
	Used    bool    `json:"used"`
}

func NewFirewallObjectResource() resource.Resource {
	return &FirewallObjectResource{}
}

func (r *FirewallObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_object"
}

func (r *FirewallObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall address object in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain this object belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall object.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Optional:    true,
				Description: "IP address of the object.",
			},
			"subnet": schema.StringAttribute{
				Optional:    true,
				Description: "Subnet of the object.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Comment for the object.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the object is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	body := firewallObjectAPI{
		Name: plan.Name.ValueString(),
	}
	if !plan.Address.IsNull() {
		v := plan.Address.ValueString()
		body.Address = &v
	}
	if !plan.Subnet.IsNull() {
		v := plan.Subnet.ValueString()
		body.Subnet = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		body.Comment = &v
	}

	var apiResp firewallObjectAPI
	path := fmt.Sprintf("/api/networking/domain/%s/objects/", domain)
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall object", err.Error())
		return
	}

	plan.Id = types.StringValue(domain + "/" + apiResp.Name)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Uuid = types.StringValue(apiResp.Uuid)
	plan.Used = types.BoolValue(apiResp.Used)
	if apiResp.Address != nil {
		plan.Address = types.StringValue(*apiResp.Address)
	}
	if apiResp.Subnet != nil {
		plan.Subnet = types.StringValue(*apiResp.Subnet)
	}
	if apiResp.Comment != nil {
		plan.Comment = types.StringValue(*apiResp.Comment)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallObjectModel
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

	var apiResp firewallObjectAPI
	path := fmt.Sprintf("/api/networking/domain/%s/objects/%s/", domain, name)
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall object", err.Error())
		return
	}

	state.Domain = types.StringValue(domain)
	state.Name = types.StringValue(apiResp.Name)
	state.Uuid = types.StringValue(apiResp.Uuid)
	state.Used = types.BoolValue(apiResp.Used)
	if apiResp.Address != nil {
		state.Address = types.StringValue(*apiResp.Address)
	} else {
		state.Address = types.StringNull()
	}
	if apiResp.Subnet != nil {
		state.Subnet = types.StringValue(*apiResp.Subnet)
	} else {
		state.Subnet = types.StringNull()
	}
	if apiResp.Comment != nil {
		state.Comment = types.StringValue(*apiResp.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallObjectResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"mcs_firewall_object does not support updates, destroy and recreate the resource.",
	)
}

func (r *FirewallObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallObjectModel
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

	path := fmt.Sprintf("/api/networking/domain/%s/objects/%s/", domain, name)
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall object", err.Error())
	}
}
