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
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	CertAuth     *ClientCertAuthModel     `tfsdk:"cert_auth"`
	PasswordAuth *ClientPasswordAuthModel `tfsdk:"password_auth"`
}

type HttpDataSourceModel struct {
	StatusCodes   []types.Int64       `tfsdk:"status_codes"`
	Endpoints     []EndpointModel     `tfsdk:"endpoints"`
	Maintenance   []EndpointModel     `tfsdk:"maintenance"`
	Path          types.String        `tfsdk:"path"`
	Tls           types.Bool          `tfsdk:"tls"`
	ServerAuth    *ServerAuthModel     `tfsdk:"server_auth"`
	ClientAuth    *ClientHttpAuthModel `tfsdk:"client_auth"`
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
				Optional: true,
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
						Optional: true,
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
						Optional: true,
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
			"down": schema.ListNestedAttribute{
				Computed: true,
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
						"error": schema.StringAttribute{
							Computed: true,
						},
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

	state.Up = []EndpointModel{}
	state.Down = []EndpointDownModel{}

	ctx = tflog.SetField(ctx, "type", "http")

	urlPath := state.Path.ValueString()
	ctx = tflog.SetField(ctx, "Path", urlPath)

	timeout := "10s"
	if !state.Timeout.IsNull() {
		timeout = state.Timeout.ValueString()
	}
	ctx = tflog.SetField(ctx, "timeout", timeout)

	isTls := true
	if !state.Tls.IsNull() {
		isTls = state.Tls.ValueBool()
	}
	ctx = tflog.SetField(ctx, "use_tls", isTls)

	retries := int64(3)
	if !state.Retries.IsNull() {
		retries = state.Retries.ValueInt64()
	}
	ctx = tflog.SetField(ctx, "max_retries", retries)

	statusCodes := []int64{200, 204}
	if len(state.StatusCodes) > 0 {
		statusCodes = []int64{}
		for _, code := range state.StatusCodes {
			statusCodes = append(statusCodes, code.ValueInt64())
		}
	}
	ctx = tflog.SetField(ctx, "status_codes", statusCodes)

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

	if state.ServerAuth != nil && (!state.ServerAuth.CaCert.IsNull()) {
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

	if state.ServerAuth != nil && (!state.ServerAuth.OverrideServerName.IsNull()) {
		tlsConf.ServerName = state.ServerAuth.OverrideServerName.ValueString()
	}

	if state.ClientAuth != nil && state.ClientAuth.CertAuth != nil && (!state.ClientAuth.CertAuth.Cert.IsNull()) && (!state.ClientAuth.CertAuth.Key.IsNull()) {
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

	endptCh := func() <-chan EndpointDownModel {
		ch := make(chan EndpointDownModel)

		go func() {
			var wg sync.WaitGroup

			for _, endpoint := range state.Endpoints {
				if endpoint.IsInMaintenace(state.Maintenance) {
					continue
				}
		
				wg.Add(1)
				go func(endpoint EndpointModel) {
					defer wg.Done()

					address := endpoint.Address.ValueString()
					port :=  endpoint.Port.ValueInt64()

					tflog.Info(ctx, "Checking Endpoint", map[string]interface{}{
						"address": address,
						"port": port,
					})

					var reqUrl url.URL
					reqUrl.Path = urlPath
					reqUrl.Host = fmt.Sprintf("%s:%d", address, port)
					if isTls {
						reqUrl.Scheme = "https"
					} else {
						reqUrl.Scheme = "http"
					}

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
		
						if state.ClientAuth != nil && state.ClientAuth.PasswordAuth != nil && (!state.ClientAuth.PasswordAuth.Username.IsNull()) && (!state.ClientAuth.PasswordAuth.Password.IsNull()) {
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
				}(endpoint)
			}

			wg.Wait()
			close(ch)
		}()

		return ch
	}()

	resCh := func(endptCh <-chan EndpointDownModel) <-chan ResultModel {
		resCh := make(chan ResultModel)

		go func() {
			res := ResultModel{
				Up: []EndpointModel{},
				Down: []EndpointDownModel{},
			}
	
			for endpt := range endptCh {
				if endpt.Error.ValueString() == "" {
					tflog.Debug(ctx, "Setting endpoint as up", map[string]interface{}{
						"address": endpt.Address.ValueString(),
						"port": endpt.Port.ValueInt64(),
					})
					res.Up = append(res.Up, EndpointModel{
						Name: endpt.Name,
						Address: endpt.Address,
						Port: endpt.Port,
					})
				} else {
					tflog.Debug(ctx, "Setting endpoint as down", map[string]interface{}{
						"address": endpt.Address.ValueString(),
						"port": endpt.Port.ValueInt64(),
					})
					res.Down = append(res.Down, endpt)
				}
			}
	
			resCh <- res
		}()

		return resCh
	}(endptCh)

	res := <-resCh
	state.Up = res.Up
	state.Down = res.Down

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
	  return
	}
}
