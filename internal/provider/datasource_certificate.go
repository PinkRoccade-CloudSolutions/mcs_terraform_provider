package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &CertificateDataSource{}

type CertificateDataSource struct {
	client *apiclient.Client
}

type CertificateDataSourceModel struct {
	Name               types.String       `tfsdk:"name"`
	Id                 types.String       `tfsdk:"id"`
	Ca                 types.Bool         `tfsdk:"ca"`
	ValidToTimestamp   types.String       `tfsdk:"valid_to_timestamp"`
	Loadbalancer       types.String       `tfsdk:"loadbalancer"`
	Certificates       []CertificateModel `tfsdk:"certificates"`
}

type CertificateModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Ca               types.Bool   `tfsdk:"ca"`
	ValidToTimestamp types.String `tfsdk:"valid_to_timestamp"`
	Loadbalancer     types.String `tfsdk:"loadbalancer"`
}

type certificateDSAPIModel struct {
	Id               string  `json:"id"`
	Name             *string `json:"name,omitempty"`
	Ca               *bool   `json:"ca,omitempty"`
	ValidToTimestamp *string `json:"valid_to_timestamp,omitempty"`
	Loadbalancer     *string `json:"loadbalancer,omitempty"`
}

func NewCertificateDataSource() datasource.DataSource {
	return &CertificateDataSource{}
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	certAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Computed: true},
		"ca": schema.BoolAttribute{Computed: true},
		"valid_to_timestamp": schema.StringAttribute{Computed: true},
		"loadbalancer": schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancing certificates. Set `name` or `id` to fetch a single certificate, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact certificate name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific certificate to look up.",
			},
			"ca": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the certificate is a CA certificate.",
			},
			"valid_to_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity end timestamp.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"certificates": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All certificates (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: certAttrs},
			},
		},
	}
}

func (d *CertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item certificateDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading certificate", err.Error())
			return
		}
		setSingleCertificate(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/loadbalancing/certificate/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []certificateDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading certificates", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *certificateDSAPIModel
		for i := range page.Results {
			if page.Results[i].Name != nil && *page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Certificate not found",
				fmt.Sprintf("No certificate with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCertificate(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CertificateDataSourceModel{
		Name:             types.StringNull(),
		Id:               types.StringNull(),
		Ca:               types.BoolNull(),
		ValidToTimestamp: types.StringNull(),
		Loadbalancer:     types.StringNull(),
		Certificates:     make([]CertificateModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.Certificates = append(state.Certificates, certificateItemToModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCertificate(state *CertificateDataSourceModel, item *certificateDSAPIModel) {
	state.Id = types.StringValue(item.Id)
	if item.Name != nil {
		state.Name = types.StringValue(*item.Name)
	} else {
		state.Name = types.StringNull()
	}
	if item.Ca != nil {
		state.Ca = types.BoolValue(*item.Ca)
	} else {
		state.Ca = types.BoolNull()
	}
	if item.ValidToTimestamp != nil {
		state.ValidToTimestamp = types.StringValue(*item.ValidToTimestamp)
	} else {
		state.ValidToTimestamp = types.StringNull()
	}
	if item.Loadbalancer != nil {
		state.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		state.Loadbalancer = types.StringNull()
	}
	state.Certificates = []CertificateModel{}
}

func certificateItemToModel(item *certificateDSAPIModel) CertificateModel {
	m := CertificateModel{
		Id: types.StringValue(item.Id),
	}
	if item.Name != nil {
		m.Name = types.StringValue(*item.Name)
	} else {
		m.Name = types.StringNull()
	}
	if item.Ca != nil {
		m.Ca = types.BoolValue(*item.Ca)
	} else {
		m.Ca = types.BoolNull()
	}
	if item.ValidToTimestamp != nil {
		m.ValidToTimestamp = types.StringValue(*item.ValidToTimestamp)
	} else {
		m.ValidToTimestamp = types.StringNull()
	}
	if item.Loadbalancer != nil {
		m.Loadbalancer = types.StringValue(*item.Loadbalancer)
	} else {
		m.Loadbalancer = types.StringNull()
	}
	return m
}
