package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &JobDataSource{}

type JobDataSource struct {
	client *apiclient.Client
}

type JobDataSourceModel struct {
	Id                 types.Int64  `tfsdk:"id"`
	JobName            types.String `tfsdk:"jobname"`
	Timestamp          types.String `tfsdk:"timestamp"`
	EndTime            types.String `tfsdk:"endtime"`
	Message            types.String `tfsdk:"message"`
	DryRun             types.Bool   `tfsdk:"dryrun"`
	ContinueOnFailure types.Bool `tfsdk:"continue_on_failure"`
}

type jobAPIModel struct {
	Id                 int    `json:"id"`
	JobName            string `json:"jobname"`
	Timestamp          string `json:"timestamp"`
	EndTime            string `json:"endtime"`
	Message            string `json:"message"`
	DryRun             bool   `json:"dryrun"`
	ContinueOnFailure bool `json:"continue_on_failure"`
}

func NewJobDataSource() datasource.DataSource {
	return &JobDataSource{}
}

func (d *JobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *JobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required: true,
			},
			"jobname": schema.StringAttribute{
				Computed: true,
			},
			"timestamp": schema.StringAttribute{
				Computed: true,
			},
			"endtime": schema.StringAttribute{
				Computed: true,
			},
			"message": schema.StringAttribute{
				Computed: true,
			},
			"dryrun": schema.BoolAttribute{
				Computed: true,
			},
			"continue_on_failure": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *JobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *JobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state JobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp jobAPIModel
	err := d.client.Get(ctx, fmt.Sprintf("/api/jobs/job/%d/", state.Id.ValueInt64()), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error reading job", err.Error())
		return
	}

	state.Id = types.Int64Value(int64(apiResp.Id))
	state.JobName = types.StringValue(apiResp.JobName)
	state.Timestamp = types.StringValue(apiResp.Timestamp)
	state.EndTime = types.StringValue(apiResp.EndTime)
	state.Message = types.StringValue(apiResp.Message)
	state.DryRun = types.BoolValue(apiResp.DryRun)
	state.ContinueOnFailure = types.BoolValue(apiResp.ContinueOnFailure)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
