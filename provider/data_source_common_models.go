package provider

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EndpointInt interface {
	GetName() string
	GetAddress() string
	GetPort() int64
}

func CmpEndpoints(endpt1 EndpointInt, endpt2 EndpointInt) bool {
	if endpt1.GetName() != endpt2.GetName() {
		return endpt1.GetName() < endpt2.GetName()
	}

	if endpt1.GetAddress() != endpt2.GetAddress() {
		return endpt1.GetAddress() < endpt2.GetAddress()
	}

	return endpt1.GetPort() < endpt2.GetPort()
}

func SortEndpoints[V EndpointInt](endpoints []V) {
	sort.SliceStable(endpoints, func(i, j int) bool { return CmpEndpoints(endpoints[i], endpoints[j])})
}

type EndpointModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}

func (endpoint EndpointModel) GetName() string {
	return endpoint.Name.ValueString()
}

func (endpoint EndpointModel) GetAddress() string {
	return endpoint.Address.ValueString()
}

func (endpoint EndpointModel) GetPort() int64 {
	return endpoint.Port.ValueInt64()
}

func (endpoint *EndpointModel) IsInMaintenace(maintenance []EndpointModel) bool {
	for _, maint := range maintenance {
		if (!maint.Name.IsNull()) && (!endpoint.Name.IsNull()) && maint.Name.ValueString() == endpoint.Name.ValueString() {
			return true
		}

		if (!maint.Address.IsNull()) && maint.Address.ValueString() == endpoint.Address.ValueString() && (!maint.Port.IsNull()) && maint.Port.ValueInt64() == endpoint.Port.ValueInt64() {
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

func (endpoint EndpointDownModel) GetName() string {
	return endpoint.Name.ValueString()
}

func (endpoint EndpointDownModel) GetAddress() string {
	return endpoint.Address.ValueString()
}

func (endpoint EndpointDownModel) GetPort() int64 {
	return endpoint.Port.ValueInt64()
}

type ResultModel struct {
	Up []EndpointModel
	Down []EndpointDownModel
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