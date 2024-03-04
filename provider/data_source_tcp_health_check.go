package provider

import (
    "context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync"
	"time"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &TcpDataSource{}
)

type TcpDataSource struct{}

func NewTcpDataSource() datasource.DataSource {
    return &TcpDataSource{}
}

func (d *TcpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_tcp"
}

type ClientTcpAuthModel struct {
	CertAuth     ClientCertAuthModel     `tfsdk:"cert_auth"`
}

type TcpDataSourceModel struct {
	Endpoints  []EndpointModel     `tfsdk:"endpoints"`
	Tls        types.Bool          `tfsdk:"tls"`
	ServerAuth ServerAuthModel     `tfsdk:"server_auth"`
	ClientAuth ClientTcpAuthModel  `tfsdk:"client_auth"`
	Timeout    types.String        `tfsdk:"timeout"`
	Retries    types.Int64         `tfsdk:"retries"`
	Up         []EndpointModel     `tfsdk:"up"`
	Down       []EndpointDownModel `tfsdk:"down"`
}

func (d *TcpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
					"override_server_name": schema.StringAttribute{
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

func (d *TcpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout := "10s"
	if !state.Timeout.IsNull() {
		timeout = state.Timeout.ValueString()
	}

	isTls := true
	if !state.Tls.IsNull() {
		isTls = state.Tls.ValueBool()
	}

	retries := int64(3)
	if !state.Retries.IsNull() {
		retries = state.Retries.ValueInt64()
	}

	dur, err := time.ParseDuration(timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Timeout Argument",
			"Could not parse timeout, unexpected error: " + err.Error(),
		)
		return
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: false,
	}

	if !state.ServerAuth.CaCert.IsNull() {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(state.ServerAuth.CaCert.ValueString()))
		if !ok {
			resp.Diagnostics.AddError(
				"Error Parsing Server CA Certificate",
				"Certificate format was not valid",
			)
			return
		}
		tlsConf.RootCAs = roots
	}

	if !state.ServerAuth.OverrideServerName.IsNull() {
		tlsConf.ServerName = state.ServerAuth.OverrideServerName.ValueString()
	}

	if (!state.ClientAuth.CertAuth.Cert.IsNull()) && (!state.ClientAuth.CertAuth.Key.IsNull()) {
		certData, err := tls.X509KeyPair([]byte(state.ClientAuth.CertAuth.Cert.ValueString()), []byte(state.ClientAuth.CertAuth.Key.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Client Tls Credentials",
				"Could not parse client cert or private key, unexpected error: " + err.Error(),
			)
			return
		}
		tlsConf.Certificates = []tls.Certificate{certData}
	}

	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	ch := make(chan EndpointDownModel)

	go func() {
		wg2.Add(1)
		defer wg2.Done()

		for res := range ch {
			if res.Error.ValueString() == "" {
				state.Up = append(state.Up, EndpointModel{
					Name: res.Name,
					Address: res.Address,
					Port: res.Port,
				})
			} else {
				state.Down = append(state.Down, res)
			}
		}
	}()

	for _, endpoint := range state.Endpoints {
		go func() {
			wg.Add(1)
			defer wg.Done()

			address := endpoint.Address.ValueString()
			port :=  endpoint.Port.ValueInt64()

			if !isTls {
				idx := retries

				for idx >= 0 {
					conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), dur)
					if err == nil {
						ch <- EndpointDownModel{
							Name: endpoint.Name,
							Address: endpoint.Address,
							Port: endpoint.Port,
							Error: types.StringValue(""),
						}
						conn.Close()
						return
					} else if idx == 0 {
						ch <- EndpointDownModel{
							Name: endpoint.Name,
							Address: endpoint.Address,
							Port: endpoint.Port,
							Error: types.StringValue(err.Error()),
						}
						return
					}
					idx = idx - 1
				}

				return
			}
		
			idx := retries

			for idx >= 0 {
				dialer := &net.Dialer{
					Timeout: dur,
				}
				conn, err := tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", address, port), tlsConf)
				if err == nil {
					ch <- EndpointDownModel{
						Name: endpoint.Name,
						Address: endpoint.Address,
						Port: endpoint.Port,
						Error: types.StringValue(""),
					}
					conn.Close()
					return
				} else if idx == 0 {
					ch <- EndpointDownModel{
						Name: endpoint.Name,
						Address: endpoint.Address,
						Port: endpoint.Port,
						Error: types.StringValue(err.Error()),
					}
					return
				}
				idx = idx - 1
			}
		}()
	}

	wg.Wait()
	close(ch)
	wg2.Wait()
  
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
	  return
	}
}