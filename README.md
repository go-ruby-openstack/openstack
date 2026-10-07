# go-ruby-openstack/openstack

A pure-Go (CGO=0) **Ruby-facing OpenStack client**. It puts a clean, idiomatic
API on top of the reference Go OpenStack SDK
[`gophercloud/v2`](https://github.com/gophercloud/gophercloud) — the mature,
de-facto pure-Go OpenStack client — the same way
[`go-ruby-confd`](https://github.com/go-ruby-confd/confd) reuses `confd`.

**gophercloud does all the OpenStack work** (Keystone auth, the service catalog,
request signing, pagination, the per-service HTTP APIs); this package adds only
the importable Ruby-facing surface: a `Connection`, per-service accessors with
list/get/create/update/delete over the core resources, resources as Ruby-style
hashes, a typed error tree, and an injectable HTTP transport seam so a consumer
such as [go-embedded-ruby](https://github.com/go-embedded-ruby) (rbgo) and its
tests can plug in a fake — no live cloud required. The Ruby semantics mirror the
`fog-openstack` gem's collection/model shape.

## Install

```
go get github.com/go-ruby-openstack/openstack@latest
```

Module path `github.com/go-ruby-openstack/openstack`; go.mod floor `go 1.27.1`, which is above
gophercloud/v2's own `go 1.25.0` rather than matching it. CGO is not used.

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-ruby-openstack/openstack"
)

func main() {
	conn, err := openstack.Connect(context.Background(), openstack.Options{
		AuthURL:     "https://keystone.example.com/v3", // OS_AUTH_URL
		Username:    "admin",
		Password:    "secret",
		ProjectName: "demo",
		DomainName:  "Default",
		Region:      "RegionOne",
	})
	if err != nil {
		panic(err) // openstack.AuthError on bad credentials
	}

	compute, err := conn.Compute()
	if err != nil {
		panic(err)
	}

	servers, err := compute.Servers() // []openstack.Resource (Ruby hashes)
	if err != nil {
		if openstack.IsNotFound(err) { /* ... */ }
		panic(err)
	}
	for _, s := range servers {
		fmt.Println(s["id"], s["name"], s["status"])
	}

	srv, err := compute.CreateServer(openstack.Resource{
		"name":      "web-1",
		"flavorRef": "2",
		"imageRef":  "cirros-uuid",
	})
	_ = srv
	_ = err
}
```

A `Resource` is `map[string]any` keyed by the resource's wire (snake_case) JSON
attributes — the natural shape for a Ruby `Hash`.

### Authentication styles

`Options` supports password (`Username`/`UserID` + `Password`), token (`Token`)
and application-credential (`ApplicationCredentialID`/`Name` + `Secret`)
authentication, each scoped by the project/domain fields.

### Injectable transport

Set `Options.Transport` (an `http.RoundTripper`) or `Options.HTTPClient` to feed
requests through a fake; the whole test suite runs against an in-process
`net/http/httptest` cloud with no network.

## Errors

`Connect` and every operation return the typed tree, mirroring the Ruby
exception hierarchy a binding exposes:

| Go type            | Ruby class              | gophercloud source     |
|--------------------|-------------------------|------------------------|
| `*Error`           | `OpenStack::Error`      | transport / 5xx / other |
| `*NotFoundError`   | `OpenStack::NotFound`   | HTTP 404               |
| `*AuthError`       | `OpenStack::AuthError`  | HTTP 401 / auth failure |
| `*ForbiddenError`  | `OpenStack::Forbidden`  | HTTP 403               |
| `*ConflictError`   | `OpenStack::Conflict`   | HTTP 409               |
| `*BadRequestError` | `OpenStack::BadRequest` | HTTP 400               |

All satisfy the `APIError` interface (`Error()`, `Status()`, `Unwrap()`). Use
the `IsNotFound` / `IsAuth` / `IsForbidden` / `IsConflict` / `IsBadRequest`
predicates.

## Coverage

Core CRUD for the six main services, tested against mocked OpenStack APIs.
Advanced / less-common services are **not yet wrapped** but remain reachable via
gophercloud directly against the authenticated `Connection`.

| Service | Accessor | Resources (list / get / create / update / delete unless noted) |
|---------|----------|-----------------------------------------------------------------|
| Identity (Keystone v3) | `conn.Identity()` | projects, users, roles, domains |
| Compute (Nova) | `conn.Compute()` | servers (+ start/stop/reboot, attach/detach volume), flavors (list/get), keypairs (list/get/create/delete) |
| Network (Neutron) | `conn.Network()` | networks, subnets, ports, routers, security groups, security-group rules (no update), floating IPs |
| Block Storage (Cinder v3) | `conn.BlockStorage()` | volumes, snapshots, volume types |
| Object Storage (Swift) | `conn.ObjectStorage()` | containers (list/create/delete), objects (list / put / get / delete) |
| Image (Glance v2) | `conn.Image()` | images (list/get/create/update/delete) + binary upload |

Deferred (use gophercloud directly if needed): Heat/orchestration, LBaaS/Octavia,
DNS/Designate, bare-metal/Ironic, shared-file-systems/Manila, key-manager/Barbican,
container-infra/Magnum, placement, quotas, per-resource metadata/tags/extra-specs,
and the many advanced actions each service exposes.

## Testing

100% statement coverage of this package's adapter code (gophercloud is a
dependency, not counted), every service's CRUD, authentication and each error
branch exercised against an in-process mocked OpenStack API — no live cloud. The
suite builds and runs on all six 64-bit architectures
(amd64/arm64/riscv64/loong64/ppc64le/s390x). See [BENCHMARKS.md](BENCHMARKS.md).

## Licensing

This package is **BSD-3-Clause** (`the go-ruby-openstack/openstack authors`).

It **imports** `github.com/gophercloud/gophercloud/v2`, which is licensed under
**Apache-2.0**. gophercloud is a normal Go module dependency — imported, never
vendored — so its Apache-2.0 license applies to that dependency, not to this
code. Downstreams redistributing a binary that links gophercloud should comply
with Apache-2.0 for that component.
