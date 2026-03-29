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

var _ datasource.DataSource = &ContactDataSource{}

type ContactDataSource struct {
	client *apiclient.Client
}

type ContactDataSourceModel struct {
	Id        types.String         `tfsdk:"id"`
	Name      types.String         `tfsdk:"name"`
	Company   types.String         `tfsdk:"company"`
	Firstname types.String         `tfsdk:"firstname"`
	Lastname  types.String         `tfsdk:"lastname"`
	Email     types.String         `tfsdk:"email"`
	Phone     types.String         `tfsdk:"phone"`
	Address   types.String         `tfsdk:"address"`
	Contacts  []ContactListModel   `tfsdk:"contacts"`
}

type ContactListModel struct {
	Id        types.String `tfsdk:"id"`
	Company   types.String `tfsdk:"company"`
	Firstname types.String `tfsdk:"firstname"`
	Lastname  types.String `tfsdk:"lastname"`
	Email     types.String `tfsdk:"email"`
	Phone     types.String `tfsdk:"phone"`
	Address   types.String `tfsdk:"address"`
}

func NewContactDataSource() datasource.DataSource {
	return &ContactDataSource{}
}

func (d *ContactDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (d *ContactDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"company": schema.StringAttribute{
			Computed: true,
		},
		"firstname": schema.StringAttribute{Computed: true},
		"lastname":  schema.StringAttribute{Computed: true},
		"email":     schema.StringAttribute{Computed: true},
		"phone":     schema.StringAttribute{Computed: true},
		"address":   schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS contacts by id, by company name, or list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Numeric id of the contact to fetch.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact company name to match (maps to the company field).",
			},
			"company": schema.StringAttribute{Computed: true},
			"firstname": schema.StringAttribute{Computed: true},
			"lastname":  schema.StringAttribute{Computed: true},
			"email":     schema.StringAttribute{Computed: true},
			"phone":     schema.StringAttribute{Computed: true},
			"address":   schema.StringAttribute{Computed: true},
			"contacts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *ContactDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ContactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ContactDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item contactAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/tenant/contacts/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading contact", err.Error())
			return
		}
		setSingleContact(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	var page struct {
		Results []contactAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/tenant/contacts/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading contacts", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *contactAPIModel
		want := config.Name.ValueString()
		for i := range page.Results {
			if page.Results[i].Company == want {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Contact not found",
				fmt.Sprintf("No contact with company %q was found.", want))
			return
		}
		setSingleContact(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := ContactDataSourceModel{
		Id:        types.StringNull(),
		Name:      types.StringNull(),
		Company:   types.StringNull(),
		Firstname: types.StringNull(),
		Lastname:  types.StringNull(),
		Email:     types.StringNull(),
		Phone:     types.StringNull(),
		Address:   types.StringNull(),
		Contacts:  make([]ContactListModel, 0, len(page.Results)),
	}
	for i := range page.Results {
		state.Contacts = append(state.Contacts, contactToListModel(&page.Results[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleContact(state *ContactDataSourceModel, item *contactAPIModel) {
	state.Id = types.StringValue(strconv.Itoa(item.Id))
	state.Name = types.StringNull()
	state.Company = types.StringValue(item.Company)
	state.Firstname = types.StringValue(item.Firstname)
	state.Lastname = types.StringValue(item.Lastname)
	state.Email = types.StringValue(item.Email)
	state.Phone = types.StringValue(item.Phone)
	state.Address = types.StringValue(item.Address)
	state.Contacts = []ContactListModel{}
}

func contactToListModel(item *contactAPIModel) ContactListModel {
	return ContactListModel{
		Id:        types.StringValue(strconv.Itoa(item.Id)),
		Company:   types.StringValue(item.Company),
		Firstname: types.StringValue(item.Firstname),
		Lastname:  types.StringValue(item.Lastname),
		Email:     types.StringValue(item.Email),
		Phone:     types.StringValue(item.Phone),
		Address:   types.StringValue(item.Address),
	}
}
