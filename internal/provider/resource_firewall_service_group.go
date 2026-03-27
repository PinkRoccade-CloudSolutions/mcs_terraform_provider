package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &FirewallServiceGroupResource{}

type FirewallServiceGroupResource struct {
	client *apiclient.Client
}

type FirewallServiceGroupModel struct {
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Uuid    types.String `tfsdk:"uuid"`
	Comment types.String `tfsdk:"comment"`
	Member  types.List   `tfsdk:"member"`
	Used    types.Bool   `tfsdk:"used"`
}

type firewallServiceGroupAPI struct {
	Name    string   `json:"name,omitempty"`
	Uuid    string   `json:"uuid,omitempty"`
	Comment *string  `json:"comment,omitempty"`
	Member  []string `json:"member,omitempty"`
	Used    bool     `json:"used"`
}

func NewFirewallServiceGroupResource() resource.Resource {
	return &FirewallServiceGroupResource{}
}

func (r *FirewallServiceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_service_group"
}

func (r *FirewallServiceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall service group in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain this service group belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the service group.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Comment for the service group.",
			},
			"member": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of member service names.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the service group is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallServiceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallServiceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallServiceGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	body := firewallServiceGroupAPI{
		Name: plan.Name.ValueString(),
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		body.Comment = &v
	}
	if !plan.Member.IsNull() {
		var members []string
		resp.Diagnostics.Append(plan.Member.ElementsAs(ctx, &members, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Member = members
	}

	var apiResp firewallServiceGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/", domain)
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall service group", err.Error())
		return
	}

	plan.Id = types.StringValue(domain + "/" + apiResp.Name)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Uuid = types.StringValue(apiResp.Uuid)
	plan.Used = types.BoolValue(apiResp.Used)
	if apiResp.Comment != nil {
		plan.Comment = types.StringValue(*apiResp.Comment)
	}
	if len(apiResp.Member) > 0 {
		listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Member)
		resp.Diagnostics.Append(diags...)
		plan.Member = listVal
	} else {
		plan.Member = types.ListValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallServiceGroupModel
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

	var apiResp firewallServiceGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/%s/", domain, name)
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall service group", err.Error())
		return
	}

	state.Domain = types.StringValue(domain)
	state.Name = types.StringValue(apiResp.Name)
	state.Uuid = types.StringValue(apiResp.Uuid)
	state.Used = types.BoolValue(apiResp.Used)
	if apiResp.Comment != nil {
		state.Comment = types.StringValue(*apiResp.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if len(apiResp.Member) > 0 {
		listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Member)
		resp.Diagnostics.Append(diags...)
		state.Member = listVal
	} else {
		state.Member = types.ListValueMust(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallServiceGroupResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"mcs_firewall_service_group does not support updates, destroy and recreate the resource.",
	)
}

func (r *FirewallServiceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallServiceGroupModel
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

	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/%s/", domain, name)
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall service group", err.Error())
	}
}
