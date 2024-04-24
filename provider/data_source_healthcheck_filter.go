package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &FilterDataSource{}
)

type FilterDataSource struct{}

func NewFilterDataSource() datasource.DataSource {
	return &FilterDataSource{}
}

func (d *FilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

type FilterDataSourceModel struct {
	Up        []EndpointModel     `tfsdk:"up"`
	Down      []EndpointDownModel `tfsdk:"down"`
	Endpoints []EndpointModel     `tfsdk:"endpoints"`
	NotEmpty  types.Bool          `tfsdk:"not_empty"`
}

func (d *FilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Filter to perform further processing on the result of a tcp or http health checks. Currently only supports the 'not empty' clause.",
		Attributes: map[string]schema.Attribute{
			"endpoints": schema.ListNestedAttribute{
				Description: "List of effective endpoints computed after processing",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"address": schema.StringAttribute{
							Computed: true,
						},
						"port": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
			"not_empty": schema.BoolAttribute{
				Description: "If set to true and the list of 'up' endpoints is empty, 'down' endpoints will be returned as the effective endpoints instead. It is used to provide continued availability in the event the terraform node has some kind of network partition",
				Optional:    true,
			},
			"up": schema.ListNestedAttribute{
				Description: "List of endpoints that will be returned as the effective endpoints by default. Should receive the 'up' output of a tcp or http check.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"address": schema.StringAttribute{
							Optional: true,
						},
						"port": schema.Int64Attribute{
							Optional: true,
						},
					},
				},
			},
			"down": schema.ListNestedAttribute{
				Description: "List of endpoints that will not be returned in the list of effective endpoints by default. Should receive the 'down' output of a tcp or http check.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"address": schema.StringAttribute{
							Optional: true,
						},
						"port": schema.Int64Attribute{
							Optional: true,
						},
						"error": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (d *FilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state FilterDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notEmpty := true
	if !state.NotEmpty.IsNull() {
		notEmpty = state.NotEmpty.ValueBool()
	}

	state.Endpoints = state.Up
	if notEmpty && len(state.Up) == 0 {
		for _, down := range state.Down {
			state.Endpoints = append(state.Endpoints, EndpointModel{
				Name:    down.Name,
				Address: down.Address,
				Port:    down.Port,
			})
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
