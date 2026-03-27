package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &InterfaceDataSource{}

type InterfaceDataSource struct {
	client *apiclient.Client
}

type InterfaceDataSourceModel struct {
	Name        types.String         `tfsdk:"name"`
	Id          types.String         `tfsdk:"id"`
	IPAddress   types.String         `tfsdk:"ipaddress"`
	IPv6Address types.String         `tfsdk:"ipv6address"`
	Network     types.String         `tfsdk:"network"`
	MACAddress  types.String         `tfsdk:"mac_address"`
	VMName      types.String         `tfsdk:"vm_name"`
	Interfaces  []InterfaceListModel `tfsdk:"interfaces"`
}

type InterfaceListModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	IPAddress   types.String `tfsdk:"ipaddress"`
	IPv6Address types.String `tfsdk:"ipv6address"`
	Network     types.String `tfsdk:"network"`
	MACAddress  types.String `tfsdk:"mac_address"`
	VMName      types.String `tfsdk:"vm_name"`
}

type interfaceAPIModel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	IPAddress   string  `json:"ipaddress"`
	IPv6Address string  `json:"ipv6address"`
	Network     *string `json:"network"`
	MACAddress  string  `json:"macAddress"`
	VMName      string  `json:"vm_name"`
}

func NewInterfaceDataSource() datasource.DataSource {
	return &InterfaceDataSource{}
}

func (d *InterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface"
}

func (d *InterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	ifaceAttrs := map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"name":        schema.StringAttribute{Computed: true},
		"ipaddress":   schema.StringAttribute{Computed: true},
		"ipv6address": schema.StringAttribute{Computed: true},
		"network":     schema.StringAttribute{Computed: true},
		"mac_address": schema.StringAttribute{Computed: true},
		"vm_name":     schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS virtual machine interfaces. Set `name` or `id` to fetch a single interface, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact interface name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific interface to look up, or the UUID of the matched interface when filtering by name.",
			},
			"ipaddress": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 address (set when a single interface is matched).",
			},
			"ipv6address": schema.StringAttribute{
				Computed:    true,
				Description: "IPv6 address (set when a single interface is matched).",
			},
			"network": schema.StringAttribute{
				Computed:    true,
				Description: "Network UUID (set when a single interface is matched).",
			},
			"mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "MAC address (set when a single interface is matched).",
			},
			"vm_name": schema.StringAttribute{
				Computed:    true,
				Description: "Parent VM name (set when a single interface is matched).",
			},
			"interfaces": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All interfaces (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: ifaceAttrs},
			},
		},
	}
}

func (d *InterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config InterfaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Direct ID lookup via /api/virtualization/interface/{id}/
	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var iface interfaceAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/virtualization/interface/%s/", config.Id.ValueString()), &iface)
		if err != nil {
			resp.Diagnostics.AddError("Error reading interface", err.Error())
			return
		}
		setSingleInterface(&config, &iface)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// List all (the API has no name filter parameter for interfaces)
	var page struct {
		Results []interfaceAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/virtualization/interface/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading interfaces", err.Error())
		return
	}

	// Single-match by name (client-side exact match)
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *interfaceAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Interface not found",
				fmt.Sprintf("No interface with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleInterface(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	// List-all mode
	state := InterfaceDataSourceModel{
		Name:        types.StringNull(),
		Id:          types.StringNull(),
		IPAddress:   types.StringNull(),
		IPv6Address: types.StringNull(),
		Network:     types.StringNull(),
		MACAddress:  types.StringNull(),
		VMName:      types.StringNull(),
		Interfaces:  make([]InterfaceListModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		network := types.StringNull()
		if item.Network != nil {
			network = types.StringValue(*item.Network)
		}
		state.Interfaces = append(state.Interfaces, InterfaceListModel{
			Id:          types.StringValue(item.Id),
			Name:        types.StringValue(item.Name),
			IPAddress:   types.StringValue(item.IPAddress),
			IPv6Address: types.StringValue(item.IPv6Address),
			Network:     network,
			MACAddress:  types.StringValue(item.MACAddress),
			VMName:      types.StringValue(item.VMName),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleInterface(state *InterfaceDataSourceModel, iface *interfaceAPIModel) {
	state.Id = types.StringValue(iface.Id)
	state.Name = types.StringValue(iface.Name)
	state.IPAddress = types.StringValue(iface.IPAddress)
	state.IPv6Address = types.StringValue(iface.IPv6Address)
	state.MACAddress = types.StringValue(iface.MACAddress)
	state.VMName = types.StringValue(iface.VMName)
	if iface.Network != nil {
		state.Network = types.StringValue(*iface.Network)
	} else {
		state.Network = types.StringNull()
	}
	state.Interfaces = []InterfaceListModel{}
}
