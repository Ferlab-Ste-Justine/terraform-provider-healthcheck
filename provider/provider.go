package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &HealthCheckProvider{}

type HealthCheckProvider struct {}


func New() func() provider.Provider {
    return func() provider.Provider {
        return &HealthCheckProvider{}
    }
}

func (p *HealthCheckProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *HealthCheckProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "healthcheck"
}

func (p *HealthCheckProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource {
        NewTcpDataSource,
		NewHttpDataSource,
		NewFilterDataSource,
    }
}

func (p *HealthCheckProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *HealthCheckProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{}
}