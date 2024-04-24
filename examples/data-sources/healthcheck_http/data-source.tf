data "healthcheck_http" "alertmanager" {
    server_auth = {
        ca_cert = file("my_ca.crt")
        override_server_name = "alertmanager.ferlab.lan"
    }
    client_auth = {
        cert_auth = {
            cert =  file("my_client.crt")
            key = file("my_client.key")
        }
    }
    path = "/-/healthy"
    status_codes = [200]
    endpoints = [
        {
            name = "alertmanager-1"
            address = "192.168.10.10"
            port = 9093
        },
        {
            name = "alertmanager-2"
            address = "192.168.10.11"
            port = 9093
        }
    ]
}

data "healthcheck_filter" "alertmanager" {
    up = data.healthcheck_tcp.alertmanager.up
    down = data.healthcheck_tcp.alertmanager.down
}

module "alertmanager_domain" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-etcd-zonefile.git"
  domain = "alertmanager.ferlab.lan"
  key_prefix = "/ferlab/coredns/"
  dns_server_name = "ns.ferlab.lan."
  a_records = [for endpoint in data.healthcheck_filter.alertmanager.endpoints: {
      prefix = "available"
      ip = endpoint.address
  }],
}