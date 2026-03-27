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

var _ resource.Resource = &MonitorIPResource{}

type MonitorIPResource struct {
	client *apiclient.Client
}

type monitorIPModel struct {
	Id                 types.String `tfsdk:"id"`
	IpAddress          types.String `tfsdk:"ipaddress"`
	Timestamp          types.String `tfsdk:"timestamp"`
	NotifyEmail        types.String `tfsdk:"notify_email"`
	LastCheckTimestamp  types.String `tfsdk:"last_check_timestamp"`
	Customer           types.String `tfsdk:"customer"`
	Comment            types.String `tfsdk:"comment"`
}

type monitorIPAPIModel struct {
	Id                 string  `json:"id,omitempty"`
	IpAddress          string  `json:"ipaddress"`
	Timestamp          string  `json:"timestamp,omitempty"`
	NotifyEmail        *string `json:"notify_email,omitempty"`
	LastCheckTimestamp  string  `json:"last_check_timestamp,omitempty"`
	Customer           string  `json:"customer"`
	Comment            *string `json:"comment,omitempty"`
}

func NewMonitorIPResource() resource.Resource {
	return &MonitorIPResource{}
}

func (r *MonitorIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_ip"
}

func (r *MonitorIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipaddress": schema.StringAttribute{
				Required: true,
			},
			"timestamp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notify_email": schema.StringAttribute{
				Optional: true,
			},
			"last_check_timestamp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"customer": schema.StringAttribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *MonitorIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func monitorIPBodyFromPlan(plan *monitorIPModel) monitorIPAPIModel {
	body := monitorIPAPIModel{
		IpAddress: plan.IpAddress.ValueString(),
		Customer:  plan.Customer.ValueString(),
	}
	if !plan.NotifyEmail.IsNull() && !plan.NotifyEmail.IsUnknown() {
		v := plan.NotifyEmail.ValueString()
		body.NotifyEmail = &v
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		v := plan.Comment.ValueString()
		body.Comment = &v
	}
	return body
}

func monitorIPStateFromAPI(result *monitorIPAPIModel, state *monitorIPModel) {
	state.Id = types.StringValue(result.Id)
	state.IpAddress = types.StringValue(result.IpAddress)
	state.Timestamp = types.StringValue(result.Timestamp)
	if result.NotifyEmail != nil {
		state.NotifyEmail = types.StringValue(*result.NotifyEmail)
	} else {
		state.NotifyEmail = types.StringNull()
	}
	state.LastCheckTimestamp = types.StringValue(result.LastCheckTimestamp)
	state.Customer = types.StringValue(result.Customer)
	if result.Comment != nil {
		state.Comment = types.StringValue(*result.Comment)
	} else {
		state.Comment = types.StringNull()
	}
}

func (r *MonitorIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan monitorIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := monitorIPBodyFromPlan(&plan)

	var result monitorIPAPIModel
	err := r.client.Post(ctx, "/api/dbl/monitorip/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating monitor IP entry", err.Error())
		return
	}

	monitorIPStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MonitorIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state monitorIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result monitorIPAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dbl/monitorip/%s/", state.Id.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading monitor IP entry", err.Error())
		return
	}

	monitorIPStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MonitorIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan monitorIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := monitorIPBodyFromPlan(&plan)

	var result monitorIPAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dbl/monitorip/%s/", plan.Id.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating monitor IP entry", err.Error())
		return
	}

	monitorIPStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MonitorIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state monitorIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/dbl/monitorip/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting monitor IP entry", err.Error())
	}
}
