package provider

import (
    "context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

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

type ClientHttpAuthModel struct {
	CertAuth     ClientCertAuthModel     `tfsdk:"cert_auth"`
	PasswordAuth ClientPasswordAuthModel `tfsdk:"password_auth"`
}

type HttpDataSourceModel struct {
	StatusCodes   []types.Int64       `tfsdk:"status_codes"`
	Endpoints     []EndpointModel     `tfsdk:"endpoints"`
	Maintenance   []EndpointModel     `tfsdk:"maintenance"`
	Path          types.String        `tfsdk:"path"`
	Tls           types.Bool          `tfsdk:"tls"`
	ServerAuth    ServerAuthModel     `tfsdk:"server_auth"`
	ClientAuth    ClientHttpAuthModel `tfsdk:"client_auth"`
	Timeout       types.String        `tfsdk:"timeout"`
	Retries       types.Int64         `tfsdk:"retries"`
	Up            []EndpointModel     `tfsdk:"up"`
	Down          []EndpointDownModel `tfsdk:"down"`
}

func (d *HttpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"status_codes": schema.ListAttribute{
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
			"maintenance": schema.ListNestedAttribute{
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
			"path": schema.StringAttribute{
				Required: true,
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
	var state HttpDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var reqUrl url.URL
	reqUrl.Path = state.Path.ValueString()

	timeout := "10s"
	if !state.Timeout.IsNull() {
		timeout = state.Timeout.ValueString()
	}

	isTls := true
	if !state.Tls.IsNull() {
		isTls = state.Tls.ValueBool()
	}

	if isTls {
		reqUrl.Scheme = "https"
	} else {
		reqUrl.Scheme = "http"
	}

	retries := int64(3)
	if !state.Retries.IsNull() {
		retries = state.Retries.ValueInt64()
	}

	statusCodes := []int64{200, 204}
	if len(state.StatusCodes) > 0 {
		statusCodes = []int64{}
		for _, code := range state.StatusCodes {
			statusCodes = append(statusCodes, code.ValueInt64())
		}
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
		if endpoint.IsInMaintenace(state.Maintenance) {
			continue
		}

		go func(reqUrl url.URL) {
			wg.Add(1)
			defer wg.Done()

			address := endpoint.Address.ValueString()
			port :=  endpoint.Port.ValueInt64()
			reqUrl.Host = fmt.Sprintf("%s:%d", address, port)

			idx := retries

			for idx >= 0 {
				client := http.Client{Timeout: dur}

				if isTls {
					client.Transport = &http.Transport{
						TLSClientConfig: tlsConf,
					}
				}

				req, reqErr := http.NewRequest(http.MethodGet, reqUrl.String(), http.NoBody)
				if reqErr != nil {
					ch <- EndpointDownModel{
						Name: endpoint.Name,
						Address: endpoint.Address,
						Port: endpoint.Port,
						Error: types.StringValue(reqErr.Error()),
					}
					return
				}

				if (!state.ClientAuth.PasswordAuth.Username.IsNull()) && (!state.ClientAuth.PasswordAuth.Password.IsNull()) {
					req.SetBasicAuth(
						state.ClientAuth.PasswordAuth.Username.ValueString(), 
						state.ClientAuth.PasswordAuth.Password.ValueString(),
					)
				}

				res, resErr := client.Do(req)
				if resErr != nil {
					if idx == 0 {
						ch <- EndpointDownModel{
							Name: endpoint.Name,
							Address: endpoint.Address,
							Port: endpoint.Port,
							Error: types.StringValue(resErr.Error()),
						}
						return
					}

					idx = idx - 1
					continue
				}

				code := int64(res.StatusCode)
				res.Body.Close()

				for _, statusCode := range statusCodes {
					if code == statusCode {
						ch <- EndpointDownModel{
							Name: endpoint.Name,
							Address: endpoint.Address,
							Port: endpoint.Port,
							Error: types.StringValue(""),
						}

						return
					}
				}

				if idx == 0 {
					ch <- EndpointDownModel{
						Name: endpoint.Name,
						Address: endpoint.Address,
						Port: endpoint.Port,
						Error: types.StringValue(fmt.Sprintf("Status code %d did not match expected values", code)),
					}
					return
				}

				idx = idx - 1
			}
		}(reqUrl)
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
