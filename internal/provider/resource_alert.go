package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &AlertResource{}

type AlertResource struct {
	client *apiclient.Client
}

type alertModel struct {
	Id             types.String `tfsdk:"id"`
	Url            types.String `tfsdk:"url"`
	CreateTime     types.String `tfsdk:"createtime"`
	LastUpdate     types.String `tfsdk:"lastupdate"`
	DuplicateCount types.Int64  `tfsdk:"duplicate_count"`
	Resource       types.String `tfsdk:"resource"`
	Environment    types.String `tfsdk:"environment"`
	Correlate      types.String `tfsdk:"correlate"`
	Event          types.String `tfsdk:"event"`
	Service        types.String `tfsdk:"service"`
	Value          types.String `tfsdk:"value"`
	Status         types.String `tfsdk:"status"`
	Timeout        types.Int64  `tfsdk:"timeout"`
	Text           types.String `tfsdk:"text"`
	AlertType      types.String `tfsdk:"type"`
	Origin         types.String `tfsdk:"origin"`
	Tags           types.String `tfsdk:"tags"`
	Severity       types.String `tfsdk:"severity"`
}

type alertAPIModel struct {
	Id             int     `json:"id,omitempty"`
	Url            *string `json:"url,omitempty"`
	CreateTime     string  `json:"createtime,omitempty"`
	LastUpdate     string  `json:"lastupdate,omitempty"`
	DuplicateCount int     `json:"duplicate_count,omitempty"`
	Resource       string  `json:"resource"`
	Environment    *string `json:"environment,omitempty"`
	Correlate      *string `json:"correlate,omitempty"`
	Event          string  `json:"event"`
	Service        *string `json:"service,omitempty"`
	Value          *string `json:"value,omitempty"`
	Status         *string `json:"status,omitempty"`
	Timeout        int     `json:"timeout,omitempty"`
	Text           *string `json:"text,omitempty"`
	AlertType      *string `json:"type,omitempty"`
	Origin         *string `json:"origin,omitempty"`
	Tags           *string `json:"tags,omitempty"`
	Severity       *string `json:"severity,omitempty"`
}

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

func (r *AlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Optional: true,
			},
			"createtime": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lastupdate": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"duplicate_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"resource": schema.StringAttribute{
				Required: true,
			},
			"environment": schema.StringAttribute{
				Optional: true,
			},
			"correlate": schema.StringAttribute{
				Optional: true,
			},
			"event": schema.StringAttribute{
				Required: true,
			},
			"service": schema.StringAttribute{
				Optional: true,
			},
			"value": schema.StringAttribute{
				Optional: true,
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Valid values: open, closed.",
			},
			"timeout": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"text": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"origin": schema.StringAttribute{
				Optional: true,
			},
			"tags": schema.StringAttribute{
				Optional: true,
			},
			"severity": schema.StringAttribute{
				Optional:    true,
				Description: "Valid values: minor, warning, major, informational, fatal, critical.",
			},
		},
	}
}

func (r *AlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func alertBodyFromPlan(plan *alertModel) alertAPIModel {
	body := alertAPIModel{
		Resource: plan.Resource.ValueString(),
		Event:    plan.Event.ValueString(),
	}
	if !plan.Url.IsNull() && !plan.Url.IsUnknown() {
		v := plan.Url.ValueString()
		body.Url = &v
	}
	if !plan.Environment.IsNull() && !plan.Environment.IsUnknown() {
		v := plan.Environment.ValueString()
		body.Environment = &v
	}
	if !plan.Correlate.IsNull() && !plan.Correlate.IsUnknown() {
		v := plan.Correlate.ValueString()
		body.Correlate = &v
	}
	if !plan.Service.IsNull() && !plan.Service.IsUnknown() {
		v := plan.Service.ValueString()
		body.Service = &v
	}
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		v := plan.Value.ValueString()
		body.Value = &v
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		v := plan.Status.ValueString()
		body.Status = &v
	}
	if !plan.Text.IsNull() && !plan.Text.IsUnknown() {
		v := plan.Text.ValueString()
		body.Text = &v
	}
	if !plan.AlertType.IsNull() && !plan.AlertType.IsUnknown() {
		v := plan.AlertType.ValueString()
		body.AlertType = &v
	}
	if !plan.Origin.IsNull() && !plan.Origin.IsUnknown() {
		v := plan.Origin.ValueString()
		body.Origin = &v
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		v := plan.Tags.ValueString()
		body.Tags = &v
	}
	if !plan.Severity.IsNull() && !plan.Severity.IsUnknown() {
		v := plan.Severity.ValueString()
		body.Severity = &v
	}
	return body
}

func alertStateFromAPI(result *alertAPIModel, state *alertModel) {
	state.Id = types.StringValue(strconv.Itoa(result.Id))
	if result.Url != nil {
		state.Url = types.StringValue(*result.Url)
	} else {
		state.Url = types.StringNull()
	}
	state.CreateTime = types.StringValue(result.CreateTime)
	state.LastUpdate = types.StringValue(result.LastUpdate)
	state.DuplicateCount = types.Int64Value(int64(result.DuplicateCount))
	state.Resource = types.StringValue(result.Resource)
	if result.Environment != nil {
		state.Environment = types.StringValue(*result.Environment)
	} else {
		state.Environment = types.StringNull()
	}
	if result.Correlate != nil {
		state.Correlate = types.StringValue(*result.Correlate)
	} else {
		state.Correlate = types.StringNull()
	}
	state.Event = types.StringValue(result.Event)
	if result.Service != nil {
		state.Service = types.StringValue(*result.Service)
	} else {
		state.Service = types.StringNull()
	}
	if result.Value != nil {
		state.Value = types.StringValue(*result.Value)
	} else {
		state.Value = types.StringNull()
	}
	if result.Status != nil {
		state.Status = types.StringValue(*result.Status)
	} else {
		state.Status = types.StringNull()
	}
	state.Timeout = types.Int64Value(int64(result.Timeout))
	if result.Text != nil {
		state.Text = types.StringValue(*result.Text)
	} else {
		state.Text = types.StringNull()
	}
	if result.AlertType != nil {
		state.AlertType = types.StringValue(*result.AlertType)
	} else {
		state.AlertType = types.StringNull()
	}
	if result.Origin != nil {
		state.Origin = types.StringValue(*result.Origin)
	} else {
		state.Origin = types.StringNull()
	}
	if result.Tags != nil {
		state.Tags = types.StringValue(*result.Tags)
	} else {
		state.Tags = types.StringNull()
	}
	if result.Severity != nil {
		state.Severity = types.StringValue(*result.Severity)
	} else {
		state.Severity = types.StringNull()
	}
}

func (r *AlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := alertBodyFromPlan(&plan)

	var result alertAPIModel
	err := r.client.Post(ctx, "/api/alerts/alert/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}

	alertStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result alertAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/alerts/alert/%s/", state.Id.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	alertStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := alertBodyFromPlan(&plan)

	var result alertAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/alerts/alert/%s/", plan.Id.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}

	alertStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/alerts/alert/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
	}
}
