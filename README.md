# About

This is a terraform provider to perform health checks, most useful when checking the availability of load balancers before putting them in dns records.

It supports tcp connection checks and http request checks, including optional tls parameters and in the case of http, optional client basic auth.