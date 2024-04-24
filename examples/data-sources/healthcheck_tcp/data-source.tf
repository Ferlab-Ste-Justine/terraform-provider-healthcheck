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