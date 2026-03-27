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

var _ resource.Resource = &DomainDblResource{}

type DomainDblResource struct {
	client *apiclient.Client
}

type domainDblModel struct {
	Id         types.String `tfsdk:"id"`
	DomainName types.String `tfsdk:"domainname"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
}

type domainDblAPIModel struct {
	Id         int    `json:"id,omitempty"`
	DomainName string `json:"domainname"`
	Timestamp  string `json:"timestamp,omitempty"`
	Source     string `json:"source"`
	Persistent *bool  `json:"persistent,omitempty"`
	Occurrence int    `json:"occurrence,omitempty"`
}

func NewDomainDblResource() resource.Resource {
	return &DomainDblResource{}
}

func (r *DomainDblResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dbl"
}

func (r *DomainDblResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domainname": schema.StringAttribute{
				Required: true,
			},
			"timestamp": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.StringAttribute{
				Required: true,
			},
			"persistent": schema.BoolAttribute{
				Optional: true,
			},
			"occurrence": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DomainDblResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func domainDblBodyFromPlan(plan *domainDblModel) domainDblAPIModel {
	body := domainDblAPIModel{
		DomainName: plan.DomainName.ValueString(),
		Source:     plan.Source.ValueString(),
	}
	if !plan.Persistent.IsNull() && !plan.Persistent.IsUnknown() {
		v := plan.Persistent.ValueBool()
		body.Persistent = &v
	}
	return body
}

func domainDblStateFromAPI(result *domainDblAPIModel, state *domainDblModel) {
	state.Id = types.StringValue(strconv.Itoa(result.Id))
	state.DomainName = types.StringValue(result.DomainName)
	state.Timestamp = types.StringValue(result.Timestamp)
	state.Source = types.StringValue(result.Source)
	if result.Persistent != nil {
		state.Persistent = types.BoolValue(*result.Persistent)
	} else {
		state.Persistent = types.BoolNull()
	}
	state.Occurrence = types.Int64Value(int64(result.Occurrence))
}

func (r *DomainDblResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainDblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := domainDblBodyFromPlan(&plan)

	var result domainDblAPIModel
	err := r.client.Post(ctx, "/api/dbl/domaindbl/", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DomainDblResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainDblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result domainDblAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", state.Id.ValueString()), &result)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DomainDblResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainDblModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := domainDblBodyFromPlan(&plan)

	var result domainDblAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", plan.Id.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain dbl entry", err.Error())
		return
	}

	domainDblStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DomainDblResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainDblModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting domain dbl entry", err.Error())
	}
}
