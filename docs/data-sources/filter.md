---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "healthcheck_filter Data Source - terraform-provider-healthcheck"
subcategory: ""
description: |-
  Filter to perform further processing on the result of a tcp or http health checks. Currently only supports the 'not empty' clause.
---

# healthcheck_filter (Data Source)

Filter to perform further processing on the result of a tcp or http health checks. Currently only supports the 'not empty' clause.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `down` (Attributes List) List of endpoints that will not be returned in the list of effective endpoints by default. Should receive the 'down' output of a tcp or http check. (see [below for nested schema](#nestedatt--down))
- `up` (Attributes List) List of endpoints that will be returned as the effective endpoints by default. Should receive the 'up' output of a tcp or http check. (see [below for nested schema](#nestedatt--up))

### Optional

- `not_empty` (Boolean) If set to true and the list of 'up' endpoints is empty, 'down' endpoints will be returned as the effective endpoints instead. It is used to provide continued availability in the event the terraform node has some kind of network partition

### Read-Only

- `endpoints` (Attributes List) List of effective endpoints computed after processing (see [below for nested schema](#nestedatt--endpoints))

<a id="nestedatt--down"></a>
### Nested Schema for `down`

Optional:

- `address` (String)
- `error` (String)
- `name` (String)
- `port` (Number)


<a id="nestedatt--up"></a>
### Nested Schema for `up`

Optional:

- `address` (String)
- `name` (String)
- `port` (Number)


<a id="nestedatt--endpoints"></a>
### Nested Schema for `endpoints`

Read-Only:

- `address` (String)
- `name` (String)
- `port` (Number)