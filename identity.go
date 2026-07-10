// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
)

// Identity is the Keystone (v3) service accessor: projects, users, roles and
// domains.
type Identity struct{ *service }

// Projects lists all projects.
func (i *Identity) Projects() ([]Resource, error) {
	return readList(i.ctx, projects.List(i.sc, projects.ListOpts{}), "projects", "")
}

// GetProject fetches a project by ID.
func (i *Identity) GetProject(id string) (Resource, error) {
	return readObject(projects.Get(i.ctx, i.sc, id), "project")
}

// CreateProject creates a project (name, domain_id, description, ...).
func (i *Identity) CreateProject(opts Resource) (Resource, error) {
	o, err := buildOpts[projects.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(projects.Create(i.ctx, i.sc, o), "project")
}

// UpdateProject updates a project.
func (i *Identity) UpdateProject(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[projects.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(projects.Update(i.ctx, i.sc, id, o), "project")
}

// DeleteProject deletes a project by ID.
func (i *Identity) DeleteProject(id string) error {
	return done(projects.Delete(i.ctx, i.sc, id))
}

// Users lists all users.
func (i *Identity) Users() ([]Resource, error) {
	return readList(i.ctx, users.List(i.sc, users.ListOpts{}), "users", "")
}

// GetUser fetches a user by ID.
func (i *Identity) GetUser(id string) (Resource, error) {
	return readObject(users.Get(i.ctx, i.sc, id), "user")
}

// CreateUser creates a user (name, domain_id, password, ...).
func (i *Identity) CreateUser(opts Resource) (Resource, error) {
	o, err := buildOpts[users.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(users.Create(i.ctx, i.sc, o), "user")
}

// UpdateUser updates a user.
func (i *Identity) UpdateUser(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[users.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(users.Update(i.ctx, i.sc, id, o), "user")
}

// DeleteUser deletes a user by ID.
func (i *Identity) DeleteUser(id string) error {
	return done(users.Delete(i.ctx, i.sc, id))
}

// Roles lists all roles.
func (i *Identity) Roles() ([]Resource, error) {
	return readList(i.ctx, roles.List(i.sc, roles.ListOpts{}), "roles", "")
}

// GetRole fetches a role by ID.
func (i *Identity) GetRole(id string) (Resource, error) {
	return readObject(roles.Get(i.ctx, i.sc, id), "role")
}

// CreateRole creates a role (name, domain_id, ...).
func (i *Identity) CreateRole(opts Resource) (Resource, error) {
	o, err := buildOpts[roles.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(roles.Create(i.ctx, i.sc, o), "role")
}

// UpdateRole updates a role.
func (i *Identity) UpdateRole(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[roles.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(roles.Update(i.ctx, i.sc, id, o), "role")
}

// DeleteRole deletes a role by ID.
func (i *Identity) DeleteRole(id string) error {
	return done(roles.Delete(i.ctx, i.sc, id))
}

// Domains lists all domains.
func (i *Identity) Domains() ([]Resource, error) {
	return readList(i.ctx, domains.List(i.sc, domains.ListOpts{}), "domains", "")
}

// GetDomain fetches a domain by ID.
func (i *Identity) GetDomain(id string) (Resource, error) {
	return readObject(domains.Get(i.ctx, i.sc, id), "domain")
}

// CreateDomain creates a domain (name, description, enabled).
func (i *Identity) CreateDomain(opts Resource) (Resource, error) {
	o, err := buildOpts[domains.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(domains.Create(i.ctx, i.sc, o), "domain")
}

// UpdateDomain updates a domain.
func (i *Identity) UpdateDomain(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[domains.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(domains.Update(i.ctx, i.sc, id, o), "domain")
}

// DeleteDomain deletes a domain by ID.
func (i *Identity) DeleteDomain(id string) error {
	return done(domains.Delete(i.ctx, i.sc, id))
}
