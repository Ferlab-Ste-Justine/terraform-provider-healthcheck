---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "healthcheck_tcp Data Source - terraform-provider-healthcheck"
subcategory: ""
description: |-
  Returns result for connection checks performed on a set on related tpc endpoints
---

# healthcheck_tcp (Data Source)

Returns result for connection checks performed on a set on related tpc endpoints

## Example Usage

```terraform
data "healthcheck_tcp" "etcd" {
    server_auth = {
        ca_cert = file("my_ca.crt")
    }
    endpoints = [
        {
            name = "etcd-1"
            address = "127.0.1.1"
            port = 2379
        },
        {
            name = "etcd-2"
            address = "127.0.2.1"
            port = 2379
        },
        {
            name = "etcd-3"
            address = "127.0.3.1"
            port = 2379
        }
    ]
    maintenance = [
        {
            name = "etcd-1"
        }
    ]
}

data "healthcheck_filter" "etcd" {
    up = data.healthcheck_tcp.etcd.up
    down = data.healthcheck_tcp.etcd.down
}

module "etcd_domain" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-etcd-zonefile.git"
  domain = "etcd.ferlab.lan"
  key_prefix = "/ferlab/coredns/"
  dns_server_name = "ns.ferlab.lan."
  a_records = [for endpoint in data.healthcheck_filter.etcd.endpoints: {
      prefix = "available"
      ip = endpoint.address
  }],
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `endpoints` (Attributes List) List of endpoints to perform connection check on (see [below for nested schema](#nestedatt--endpoints))

### Optional

- `client_auth` (Attributes) (see [below for nested schema](#nestedatt--client_auth))
- `maintenance` (Attributes List) Optional list of endpoints that are under maintenance. Those endpoints will not be polled or included in the results (see [below for nested schema](#nestedatt--maintenance))
- `retries` (Number) Number of retries to perform on a particular endpoint with a failing connection before determining that it is down
- `server_auth` (Attributes) (see [below for nested schema](#nestedatt--server_auth))
- `timeout` (String) Timeout after which a connection attempt on an endpoint will be aborted
- `tls` (Boolean) Whether a tls connection should be attempted

### Read-Only

- `down` (Attributes List) List of endpoints that could not be connected to (see [below for nested schema](#nestedatt--down))
- `up` (Attributes List) List of endpoints that were successfully connected to (see [below for nested schema](#nestedatt--up))

<a id="nestedatt--endpoints"></a>
### Nested Schema for `endpoints`

Required:

- `address` (String) Address the endpoint is listening on
- `port` (Number) Port the endpoint is listening on

Optional:

- `name` (String) Optional name to provide for the endpoint


<a id="nestedatt--client_auth"></a>
### Nested Schema for `client_auth`

Required:

- `cert_auth` (Attributes) Parameters to perform client certificate authentication during the connection (see [below for nested schema](#nestedatt--client_auth--cert_auth))

<a id="nestedatt--client_auth--cert_auth"></a>
### Nested Schema for `client_auth.cert_auth`

Required:

- `cert` (String) Public certificate to use to authentify the client
- `key` (String, Sensitive) Private key to use to authentify the client



<a id="nestedatt--maintenance"></a>
### Nested Schema for `maintenance`

Optional:

- `address` (String) If provided, endpoint to exclude will be matched by the provided address (in addition to the 'port' field). In such cases, the 'name' field should not be provided
- `name` (String) If provided, endpoint to exclude will be matched by name. In such cases, the 'address' and 'port' fields should not be provided
- `port` (Number) If provided, endpoint to exclude will be matched by the provided port (in addition to the 'address' field). In such cases, the 'name' field should not be provided


<a id="nestedatt--server_auth"></a>
### Nested Schema for `server_auth`

Required:

- `ca_cert` (String) In the case of a tls connection, a CA certificate to check the validity of the server endpoints

Optional:

- `override_server_name` (String) An alternate name to use instead of the passed endpoints address when validating the endpoints' server certificate


<a id="nestedatt--down"></a>
### Nested Schema for `down`

Read-Only:

- `address` (String) Address of the endpoint, corresponding to the entry passed to the 'endpoints' argument
- `error` (String) Error message that was returned during the last attempt to connect
- `name` (String) Name of the endpoint, corresponding to the entry passed to the 'endpoints' argument
- `port` (Number) Port of the endpoint, corresponding to the entry passed to the 'endpoints' argument


<a id="nestedatt--up"></a>
### Nested Schema for `up`

Read-Only:

- `address` (String) Address of the endpoint, corresponding to the entry passed to the 'endpoints' argument
- `name` (String) Name of the endpoint, corresponding to the entry passed to the 'endpoints' argument
- `port` (Number) Port of the endpoint, corresponding to the entry passed to the 'endpoints' argument