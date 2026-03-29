package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiskDataSource{}

type DiskDataSource struct {
	client *apiclient.Client
}

type DiskDataSourceModel struct {
	Name     types.String    `tfsdk:"name"`
	Id       types.String    `tfsdk:"id"`
	Size     types.Int64     `tfsdk:"size"`
	Path     types.String    `tfsdk:"path"`
	DiskType types.String    `tfsdk:"type"`
	Disks    []DiskListModel `tfsdk:"disks"`
}

type DiskListModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Size     types.Int64  `tfsdk:"size"`
	Path     types.String `tfsdk:"path"`
	DiskType types.String `tfsdk:"type"`
}

type diskDSAPIModel struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	DiskType string `json:"type"`
}

func NewDiskDataSource() datasource.DataSource {
	return &DiskDataSource{}
}

func (d *DiskDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}

func (d *DiskDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	diskAttrs := map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Computed: true},
		"size": schema.Int64Attribute{Computed: true, Description: "Size in GB."},
		"path": schema.StringAttribute{Computed: true},
		"type": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS virtual machine disks. Set `name` or `id` to fetch a single disk, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact disk name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific disk to look up.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size in GB (set when a single disk is matched).",
			},
			"path": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Disk provisioning type (set when a single disk is matched).",
			},
			"disks": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All disks (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: diskAttrs},
			},
		},
	}
}

func (d *DiskDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DiskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DiskDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var disk diskDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/virtualization/disk/%s/", config.Id.ValueString()), &disk)
		if err != nil {
			resp.Diagnostics.AddError("Error reading disk", err.Error())
			return
		}
		setSingleDisk(&config, &disk)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/virtualization/disk/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []diskDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading disks", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *diskDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Disk not found",
				fmt.Sprintf("No disk with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleDisk(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := DiskDataSourceModel{
		Name:     types.StringNull(),
		Id:       types.StringNull(),
		Size:     types.Int64Null(),
		Path:     types.StringNull(),
		DiskType: types.StringNull(),
		Disks:    make([]DiskListModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Disks = append(state.Disks, toDiskListModel(&item))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleDisk(state *DiskDataSourceModel, disk *diskDSAPIModel) {
	state.Id = types.StringValue(disk.Id)
	state.Name = types.StringValue(disk.Name)
	state.Size = types.Int64Value(disk.Size)
	state.Path = types.StringValue(disk.Path)
	state.DiskType = types.StringValue(disk.DiskType)
	state.Disks = []DiskListModel{}
}

func toDiskListModel(disk *diskDSAPIModel) DiskListModel {
	return DiskListModel{
		Id:       types.StringValue(disk.Id),
		Name:     types.StringValue(disk.Name),
		Size:     types.Int64Value(disk.Size),
		Path:     types.StringValue(disk.Path),
		DiskType: types.StringValue(disk.DiskType),
	}
}
