module "ca" {
  source = "./ca"
  common_name = "test"
}

resource "local_file" "ca_cert" {
  content         = module.ca.certificate
  file_permission = "0600"
  filename        = "${path.module}/output/ca.crt"
}

resource "local_file" "ca_key" {
  content         = module.ca.key
  file_permission = "0600"
  filename        = "${path.module}/output/ca.key"
}

resource "tls_private_key" "server_client" {
  algorithm   = "RSA"
  rsa_bits    = 4096
}

resource "tls_cert_request" "server_client" {
  private_key_pem = tls_private_key.server_client.private_key_pem
  ip_addresses    = ["127.0.0.1"]
  subject {
    common_name  = "test"
    organization = "test"
  }
}

resource "tls_locally_signed_cert" "server_client" {
  cert_request_pem   = tls_cert_request.server_client.cert_request_pem
  ca_private_key_pem = module.ca.key
  ca_cert_pem        = module.ca.certificate

  validity_period_hours = 100*365*24
  early_renewal_hours = 365*24

  allowed_uses = [
    "client_auth",
    "server_auth",
  ]

  is_ca_certificate = false
}

resource "local_file" "server_client_cert" {
  content         = tls_locally_signed_cert.server_client.cert_pem
  file_permission = "0600"
  filename        = "${path.module}/output/server_client.crt"
}

resource "local_file" "server_client_key" {
  content         = tls_private_key.server_client.private_key_pem
  file_permission = "0600"
  filename        = "${path.module}/output/server_client.key"
}