package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EndpointModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}

func (endpoint *EndpointModel) IsInMaintenace(maintenance []EndpointModel) bool {
	for _, maint := range maintenance {
		if maint.Name.ValueString() == endpoint.Name.ValueString() {
			return true
		}

		if maint.Address.ValueString() == endpoint.Address.ValueString() && maint.Port.ValueInt64() == endpoint.Port.ValueInt64() {
			return true
		}
	}

	return false
}

type EndpointDownModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	Error   types.String `tfsdk:"error"`
}

type ServerAuthModel struct {
	CaCert           types.String `tfsdk:"ca_cert"`
	OverrideServerName types.String `tfsdk:"override_server_name"`
}

type ClientCertAuthModel struct {
	Cert types.String `tfsdk:"cert"`
	Key  types.String `tfsdk:"key"`
}

type ClientPasswordAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}