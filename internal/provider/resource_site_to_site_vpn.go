package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &SiteToSiteVPNResource{}

type SiteToSiteVPNResource struct {
	client *apiclient.Client
}

type SiteToSiteVPNResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Uuid               types.String `tfsdk:"uuid"`
	Name               types.String `tfsdk:"name"`
	State              types.String `tfsdk:"state"`
	LastStatus         types.String `tfsdk:"last_status"`
	Resets             types.Int64  `tfsdk:"resets"`
	LastCheck          types.String `tfsdk:"last_check"`
	LastReset types.String `tfsdk:"last_reset"`
}

type siteToSiteVPNAPIModel struct {
	Id                 int    `json:"id,omitempty"`
	Uuid               string `json:"uuid,omitempty"`
	Name               string `json:"name"`
	State              string `json:"state,omitempty"`
	LastStatus         string `json:"last_status,omitempty"`
	Resets             int64  `json:"resets,omitempty"`
	LastCheck          string `json:"last_check,omitempty"`
	LastReset string `json:"last_reset,omitempty"`
}

func NewSiteToSiteVPNResource() resource.Resource {
	return &SiteToSiteVPNResource{}
}

func (r *SiteToSiteVPNResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_to_site_vpn"
}

func (r *SiteToSiteVPNResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"state": schema.StringAttribute{
				Optional: true,
			},
			"last_status": schema.StringAttribute{
				Optional: true,
			},
			"resets": schema.Int64Attribute{
				Optional: true,
			},
			"last_check": schema.StringAttribute{
				Optional: true,
			},
			"last_reset": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *SiteToSiteVPNResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SiteToSiteVPNResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := siteToSiteVPNAPIModel{
		Name:       plan.Name.ValueString(),
		State:      plan.State.ValueString(),
		LastStatus: plan.LastStatus.ValueString(),
		Resets:     plan.Resets.ValueInt64(),
		LastCheck:  plan.LastCheck.ValueString(),
		LastReset:  plan.LastReset.ValueString(),
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Post(ctx, "/api/vpn/site_to_site/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating site-to-site VPN", err.Error())
		return
	}

	plan.Id = types.StringValue(strconv.Itoa(apiResp.Id))
	plan.Uuid = types.StringValue(apiResp.Uuid)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteToSiteVPNResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading site-to-site VPN", err.Error())
		return
	}

	state.Id = types.StringValue(strconv.Itoa(apiResp.Id))
	state.Uuid = types.StringValue(apiResp.Uuid)
	state.Name = types.StringValue(apiResp.Name)
	state.State = types.StringValue(apiResp.State)
	state.LastStatus = types.StringValue(apiResp.LastStatus)
	state.Resets = types.Int64Value(apiResp.Resets)
	state.LastCheck = types.StringValue(apiResp.LastCheck)
	state.LastReset = types.StringValue(apiResp.LastReset)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SiteToSiteVPNResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := siteToSiteVPNAPIModel{
		Name:       plan.Name.ValueString(),
		State:      plan.State.ValueString(),
		LastStatus: plan.LastStatus.ValueString(),
		Resets:     plan.Resets.ValueInt64(),
		LastCheck:  plan.LastCheck.ValueString(),
		LastReset:  plan.LastReset.ValueString(),
	}

	var apiResp siteToSiteVPNAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", plan.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating site-to-site VPN", err.Error())
		return
	}

	plan.Uuid = types.StringValue(apiResp.Uuid)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SiteToSiteVPNResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SiteToSiteVPNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting site-to-site VPN", err.Error())
	}
}
