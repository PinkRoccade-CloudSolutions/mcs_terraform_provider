package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &DomainDblDataSource{}

type DomainDblDataSource struct {
	client *apiclient.Client
}

type DomainDblDataSourceModel struct {
	Id          types.String         `tfsdk:"id"`
	DomainName  types.String         `tfsdk:"domainname"`
	Timestamp   types.String         `tfsdk:"timestamp"`
	Source      types.String         `tfsdk:"source"`
	Persistent  types.Bool           `tfsdk:"persistent"`
	Occurrence  types.Int64          `tfsdk:"occurrence"`
	DomainDbls  []DomainDblListModel `tfsdk:"domain_dbls"`
}

type DomainDblListModel struct {
	Id         types.String `tfsdk:"id"`
	DomainName types.String `tfsdk:"domainname"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Source     types.String `tfsdk:"source"`
	Persistent types.Bool   `tfsdk:"persistent"`
	Occurrence types.Int64  `tfsdk:"occurrence"`
}

func NewDomainDblDataSource() datasource.DataSource {
	return &DomainDblDataSource{}
}

func (d *DomainDblDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dbl"
}

func (d *DomainDblDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"domainname": schema.StringAttribute{
			Computed: true,
		},
		"timestamp":  schema.StringAttribute{Computed: true},
		"source":     schema.StringAttribute{Computed: true},
		"persistent": schema.BoolAttribute{Computed: true},
		"occurrence": schema.Int64Attribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS domain DBL entries by id or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Numeric id to fetch a single entry; omit to list all.",
			},
			"domainname": schema.StringAttribute{Computed: true},
			"timestamp":  schema.StringAttribute{Computed: true},
			"source":     schema.StringAttribute{Computed: true},
			"persistent": schema.BoolAttribute{Computed: true},
			"occurrence": schema.Int64Attribute{Computed: true},
			"domain_dbls": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *DomainDblDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDblDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainDblDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item domainDblAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/dbl/domaindbl/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading domain dbl entry", err.Error())
			return
		}
		setSingleDomainDbl(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	var page struct {
		Results []domainDblAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/dbl/domaindbl/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading domain dbl entries", err.Error())
		return
	}

	state := DomainDblDataSourceModel{
		Id:         types.StringNull(),
		DomainName: types.StringNull(),
		Timestamp:  types.StringNull(),
		Source:     types.StringNull(),
		Persistent: types.BoolNull(),
		Occurrence: types.Int64Null(),
		DomainDbls: make([]DomainDblListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.DomainDbls = append(state.DomainDbls, domainDblToListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleDomainDbl(state *DomainDblDataSourceModel, item *domainDblAPIModel) {
	state.Id = types.StringValue(strconv.Itoa(item.Id))
	state.DomainName = types.StringValue(item.DomainName)
	state.Timestamp = types.StringValue(item.Timestamp)
	state.Source = types.StringValue(item.Source)
	state.Occurrence = types.Int64Value(int64(item.Occurrence))
	if item.Persistent != nil {
		state.Persistent = types.BoolValue(*item.Persistent)
	} else {
		state.Persistent = types.BoolNull()
	}
	state.DomainDbls = []DomainDblListModel{}
}

func domainDblToListModel(item *domainDblAPIModel) DomainDblListModel {
	m := DomainDblListModel{
		Id:         types.StringValue(strconv.Itoa(item.Id)),
		DomainName: types.StringValue(item.DomainName),
		Timestamp:  types.StringValue(item.Timestamp),
		Source:     types.StringValue(item.Source),
		Occurrence: types.Int64Value(int64(item.Occurrence)),
	}
	if item.Persistent != nil {
		m.Persistent = types.BoolValue(*item.Persistent)
	} else {
		m.Persistent = types.BoolNull()
	}
	return m
}
