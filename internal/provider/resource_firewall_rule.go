package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &FirewallRuleResource{}

type FirewallRuleResource struct {
	client *apiclient.Client
}

type FirewallRuleModel struct {
	Id               types.String `tfsdk:"id"`
	Domain           types.String `tfsdk:"domain"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Src              types.List   `tfsdk:"src"`
	Dst              types.List   `tfsdk:"dst"`
	SrcIntf          types.List   `tfsdk:"src_intf"`
	DstIntf          types.List   `tfsdk:"dst_intf"`
	Service          types.List   `tfsdk:"service"`
	Action           types.Bool   `tfsdk:"action"`
	Used             types.Bool   `tfsdk:"used"`
	Compliant        types.Bool   `tfsdk:"compliant"`
	Uuid             types.String `tfsdk:"uuid"`
	PolicyId         types.Int64  `tfsdk:"policyid"`
	HitCount         types.Int64  `tfsdk:"hit_count"`
	LastHit          types.String `tfsdk:"last_hit"`
	Group            types.String `tfsdk:"group"`
	Comment          types.String `tfsdk:"comment"`
	CompliancyErrors types.List   `tfsdk:"compliancy_errors"`
}

type firewallRuleAPI struct {
	Enabled          bool     `json:"enabled"`
	Src              []string `json:"src,omitempty"`
	Dst              []string `json:"dst,omitempty"`
	SrcIntf          []string `json:"src_intf,omitempty"`
	DstIntf          []string `json:"dst_intf,omitempty"`
	Service          []string `json:"service,omitempty"`
	Action           bool     `json:"action"`
	Used             bool     `json:"used"`
	Compliant        bool     `json:"compliant"`
	Uuid             string   `json:"uuid,omitempty"`
	PolicyId         int      `json:"policyid"`
	HitCount         int      `json:"hit_count"`
	LastHit          string   `json:"last_hit,omitempty"`
	Group            string   `json:"group,omitempty"`
	Comment          *string  `json:"comment,omitempty"`
	CompliancyErrors []string `json:"compliancy_errors,omitempty"`
}

func NewFirewallRuleResource() resource.Resource {
	return &FirewallRuleResource{}
}

func (r *FirewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *FirewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall rule in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain this rule belongs to.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the rule is enabled.",
			},
			"src": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Source addresses.",
			},
			"dst": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Destination addresses.",
			},
			"src_intf": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Source interfaces.",
			},
			"dst_intf": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Destination interfaces.",
			},
			"service": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Services for the rule.",
			},
			"action": schema.BoolAttribute{
				Required:    true,
				Description: "Action for the rule (true=allow, false=deny).",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the rule is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"compliant": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the rule is compliant.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"policyid": schema.Int64Attribute{
				Computed:      true,
				Description:   "Policy ID assigned by the firewall.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"hit_count": schema.Int64Attribute{
				Computed:      true,
				Description:   "Number of times the rule was hit.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"last_hit": schema.StringAttribute{
				Computed:      true,
				Description:   "Timestamp of the last hit.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"group": schema.StringAttribute{
				Computed:      true,
				Description:   "Group the rule belongs to.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Comment for the rule.",
			},
			"compliancy_errors": schema.ListAttribute{
				ElementType:   types.StringType,
				Computed:      true,
				Description:   "List of compliancy errors.",
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	body := firewallRuleAPI{
		Enabled: plan.Enabled.ValueBool(),
		Action:  plan.Action.ValueBool(),
	}
	if !plan.Src.IsNull() {
		resp.Diagnostics.Append(plan.Src.ElementsAs(ctx, &body.Src, false)...)
	}
	if !plan.Dst.IsNull() {
		resp.Diagnostics.Append(plan.Dst.ElementsAs(ctx, &body.Dst, false)...)
	}
	if !plan.SrcIntf.IsNull() {
		resp.Diagnostics.Append(plan.SrcIntf.ElementsAs(ctx, &body.SrcIntf, false)...)
	}
	if !plan.DstIntf.IsNull() {
		resp.Diagnostics.Append(plan.DstIntf.ElementsAs(ctx, &body.DstIntf, false)...)
	}
	if !plan.Service.IsNull() {
		resp.Diagnostics.Append(plan.Service.ElementsAs(ctx, &body.Service, false)...)
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		body.Comment = &v
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallRuleAPI
	path := fmt.Sprintf("/api/networking/domain/%s/rules/", domain)
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall rule", err.Error())
		return
	}

	mapFirewallRuleToState(ctx, &plan, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(state.Id.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID format", "Expected domain/policyid")
		return
	}
	domain, policyid := parts[0], parts[1]

	var apiResp firewallRuleAPI
	path := fmt.Sprintf("/api/networking/domain/%s/rules/%s/", domain, policyid)
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall rule", err.Error())
		return
	}

	mapFirewallRuleToState(ctx, &state, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"mcs_firewall_rule does not support updates, destroy and recreate the resource.",
	)
}

func (r *FirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(state.Id.ValueString(), "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid ID format", "Expected domain/policyid")
		return
	}
	domain, policyid := parts[0], parts[1]

	path := fmt.Sprintf("/api/networking/domain/%s/rules/%s/", domain, policyid)
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall rule", err.Error())
	}
}

func mapFirewallRuleToState(ctx context.Context, model *FirewallRuleModel, domain string, api *firewallRuleAPI, diagnostics *diag.Diagnostics) {
	model.Id = types.StringValue(domain + "/" + strconv.Itoa(api.PolicyId))
	model.Domain = types.StringValue(domain)
	model.Enabled = types.BoolValue(api.Enabled)
	model.Action = types.BoolValue(api.Action)
	model.Used = types.BoolValue(api.Used)
	model.Compliant = types.BoolValue(api.Compliant)
	model.Uuid = types.StringValue(api.Uuid)
	model.PolicyId = types.Int64Value(int64(api.PolicyId))
	model.HitCount = types.Int64Value(int64(api.HitCount))
	model.LastHit = types.StringValue(api.LastHit)
	model.Group = types.StringValue(api.Group)

	if api.Comment != nil {
		model.Comment = types.StringValue(*api.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	setStringList := func(src []string) types.List {
		if len(src) > 0 {
			listVal, diags := types.ListValueFrom(ctx, types.StringType, src)
			diagnostics.Append(diags...)
			return listVal
		}
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	model.Src = setStringList(api.Src)
	model.Dst = setStringList(api.Dst)
	model.SrcIntf = setStringList(api.SrcIntf)
	model.DstIntf = setStringList(api.DstIntf)
	model.Service = setStringList(api.Service)
	model.CompliancyErrors = setStringList(api.CompliancyErrors)
}
