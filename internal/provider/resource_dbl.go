package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &DblResource{}

type DblResource struct {
	client *apiclient.Client
}

type dblModel struct {
	Id         types.String `tfsdk:"id"`
	ApiId      types.Int64  `tfsdk:"api_id"`
	IpAddress  types.String `tfsdk:"ipaddress"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Hostname   types.String `tfsdk:"hostname"`
}

type dblAPIModel struct {
	Id         int     `json:"id,omitempty"`
	IpAddress  string  `json:"ipaddress"`
	Timestamp  string  `json:"timestamp,omitempty"`
	Source     *string `json:"source,omitempty"`
	Occurrence int     `json:"occurrence,omitempty"`
	Persistent *bool   `json:"persistent,omitempty"`
	Hostname   string  `json:"hostname,omitempty"`
}

func NewDblResource() resource.Resource {
	return &DblResource{}
}

func (r *DblResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dbl"
}

func (r *DblResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
			"source": schema.StringAttribute{
				Optional: true,
			},
			"occurrence": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"persistent": schema.BoolAttribute{
				Optional: true,
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DblResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func dblBodyFromPlan(plan *dblModel) dblAPIModel {
	body := dblAPIModel{
		IpAddress: plan.IpAddress.ValueString(),
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() {
		v := plan.Source.ValueString()
		body.Source = &v
	}
	if !plan.Persistent.IsNull() && !plan.Persistent.IsUnknown() {
		v := plan.Persistent.ValueBool()
		body.Persistent = &v
	}
	return body
}

func dblStateFromAPI(result *dblAPIModel, state *dblModel) {
	state.Id = types.StringValue(result.IpAddress)
	state.ApiId = types.Int64Value(int64(result.Id))
	state.IpAddress = types.StringValue(result.IpAddress)
	state.Timestamp = types.StringValue(result.Timestamp)
	if result.Source != nil {
		state.Source = types.StringValue(*result.Source)
	} else {
		state.Source = types.StringNull()
	}
	state.Occurrence = types.Int64Value(int64(result.Occurrence))
	if result.Persistent != nil {
		state.Persistent = types.BoolValue(*result.Persistent)
	} else {
		state.Persistent = types.BoolNull()
	}
	state.Hostname = types.StringValue(result.Hostname)
}

func (r *DblResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := dblBodyFromPlan(&plan)

	var result dblAPIModel
	err := r.client.Post(ctx, "/api/dbl/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DblResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result dblAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dbl/%s/", state.Id.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DblResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := dblBodyFromPlan(&plan)

	var result dblAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dbl/%s/", plan.Id.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating dbl entry", err.Error())
		return
	}

	dblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DblResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/dbl/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting dbl entry", err.Error())
	}
}
