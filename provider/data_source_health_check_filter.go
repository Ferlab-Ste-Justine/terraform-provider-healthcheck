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
		Attributes: map[string]schema.Attribute{
			"endpoints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{},
						"address": schema.StringAttribute{},
						"port": schema.Int64Attribute{},
					},
				},
			},
			"not_empty": schema.BoolAttribute{
				Optional: true,
			},
			"up": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{},
						"address": schema.StringAttribute{},
						"port": schema.Int64Attribute{},
					},
				},
			},
			"down": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{},
						"address": schema.StringAttribute{},
						"port": schema.Int64Attribute{},
						"error": schema.StringAttribute{},
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
				Name: down.Name,
				Address: down.Address,
				Port: down.Port,
			})
		}
	}
  
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
	  return
	}
}