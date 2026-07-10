// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

// Package openstack is a pure-Go (no cgo) Ruby-facing OpenStack client. It
// puts a clean, idiomatic surface on top of the reference Go OpenStack SDK
// github.com/gophercloud/gophercloud/v2 (Apache-2.0 licensed; imported, never
// vendored). gophercloud already implements everything that is OpenStack: the
// Keystone authentication dance, the service catalog, request signing,
// pagination and the per-service HTTP APIs. This package does not reimplement
// any of that; it wraps gophercloud and adds the thin conveniences a consumer
// such as go-embedded-ruby (rbgo) needs to expose an OpenStack client to Ruby:
//
//   - [Connect] performs Keystone v3 authentication (password, token or
//     application-credential) and returns a [Connection] carrying the token
//     and service catalog, mirroring OpenStack::Connection.new;
//   - per-service accessors ([Connection.Compute], [Connection.Network],
//     [Connection.BlockStorage], [Connection.ObjectStorage],
//     [Connection.Image], [Connection.Identity]) each offering list / get /
//     create / update / delete over the core resources, in the spirit of
//     fog-openstack's collections (Fog::OpenStack::Compute#servers, .get,
//     .create, .destroy);
//   - resources returned as [Resource] hashes (map[string]any keyed by the
//     JSON attribute names), the natural shape for a Ruby Hash;
//   - a typed error tree ([Error], [NotFoundError], [AuthError],
//     [ForbiddenError], [ConflictError], [BadRequestError]) mapping
//     gophercloud's HTTP status codes onto the Ruby exception hierarchy
//     OpenStack::Error < StandardError;
//   - an injectable HTTP transport seam ([Options.Transport] /
//     [Options.HTTPClient]) so tests and rbgo can supply a fake and no live
//     cloud is ever required.
//
// Scope: the core CRUD for the six main services (Nova, Neutron, Cinder,
// Swift, Glance, Keystone). The full OpenStack API is enormous; less-common
// services and advanced per-resource operations are intentionally not wrapped
// but remain reachable through gophercloud directly against the authenticated
// [Connection]. See the README for the exact coverage matrix.
//
// The package has no dependency on any Ruby runtime: the surface is Go-typed,
// and a Ruby binding layer marshals Ruby values onto these Go types.
package openstack
