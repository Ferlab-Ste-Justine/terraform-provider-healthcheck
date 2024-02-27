package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &HttpDataSource{}
)

type HttpDataSource struct{}

func NewHttpDataSource() datasource.DataSource {
    return &HttpDataSource{}
}

func (d *HttpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_http"
}

func (d *HttpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"expected_codes": schema.ListAttribute{
				Required: true,
                ElementType: types.ListType{
                    ElemType: types.Int64Type,
                },
			},
			"endpoints": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"address": schema.StringAttribute{
							Required: true,
						},
						"port": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"tls": schema.BoolAttribute{
				Optional: true,
			},
			"server_auth": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"ca_cert": schema.StringAttribute{
						Required: true,
					},
					"override_hostname": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"client_auth": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"cert_auth": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"cert": schema.StringAttribute{
								Required: true,
							},
							"key": schema.StringAttribute{
								Required: true,
								Sensitive: true,
							},
						},
					},
					"password_auth": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Required: true,
							},
							"password": schema.StringAttribute{
								Required: true,
								Sensitive: true,
							},
						},
					},
				},
			},
			"timeout": schema.StringAttribute{
				Optional: true,
			},
			"retries": schema.Int64Attribute{
				Optional: true,
			},
			"up": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{},
						"address": schema.StringAttribute{},
						"port": schema.Int64Attribute{},
					},
				},
			},
			"down": schema.ListNestedAttribute{
				Computed: true,
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

func (d *HttpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
}
