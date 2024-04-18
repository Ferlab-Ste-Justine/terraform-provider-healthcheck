data "healthcheck_tcp" "test" {
    server_auth = {
        ca_cert = file("${path.module}/../credentials/output/ca.crt")
    }
    client_auth = {
        cert_auth = {
            cert = file("${path.module}/../credentials/output/server_client.crt")
            key = file("${path.module}/../credentials/output/server_client.key")
        }
    }
    endpoints = [
        {
            name = "server-1"
            address = "127.0.0.1"
            port = 8443
        }
    ]
}

resource "local_file" "status" {
  content         = templatefile(
    "${path.module}/templates/status.md.tpl",
    {
      up = data.healthcheck_tcp.test.up
      down = data.healthcheck_tcp.test.down
    }
  )
  file_permission = "0600"
  filename        = "${path.module}/output/tcp_status.md"
}