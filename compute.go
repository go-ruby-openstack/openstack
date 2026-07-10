// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/volumeattach"
)

// Compute is the Nova service accessor, exposing the core CRUD over servers,
// flavors and keypairs plus server power actions and volume attachment. The
// shape mirrors fog-openstack's Fog::OpenStack::Compute collections.
type Compute struct{ *service }

// Servers lists all servers (detailed).
func (c *Compute) Servers() ([]Resource, error) {
	return readList(c.ctx, servers.List(c.sc, servers.ListOpts{}), "servers", "")
}

// Server fetches a single server by ID.
func (c *Compute) Server(id string) (Resource, error) {
	return readObject(servers.Get(c.ctx, c.sc, id), "server")
}

// CreateServer boots a server from the supplied option hash (name, flavorRef,
// imageRef, networks, ...).
func (c *Compute) CreateServer(opts Resource) (Resource, error) {
	o, err := buildOpts[servers.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(servers.Create(c.ctx, c.sc, o, nil), "server")
}

// UpdateServer updates a server's mutable attributes (name, ...).
func (c *Compute) UpdateServer(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[servers.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(servers.Update(c.ctx, c.sc, id, o), "server")
}

// DeleteServer deletes a server by ID.
func (c *Compute) DeleteServer(id string) error {
	return done(servers.Delete(c.ctx, c.sc, id))
}

// StartServer powers on a stopped server.
func (c *Compute) StartServer(id string) error {
	return done(servers.Start(c.ctx, c.sc, id))
}

// StopServer powers off a running server.
func (c *Compute) StopServer(id string) error {
	return done(servers.Stop(c.ctx, c.sc, id))
}

// RebootServer reboots a server. method is "SOFT" (default) or "HARD".
func (c *Compute) RebootServer(id, method string) error {
	rm := servers.SoftReboot
	if method == "HARD" {
		rm = servers.HardReboot
	}
	return done(servers.Reboot(c.ctx, c.sc, id, servers.RebootOpts{Type: rm}))
}

// AttachVolume attaches a volume to a server; opts carries at least
// {"volumeId" => "..."}.
func (c *Compute) AttachVolume(serverID string, opts Resource) (Resource, error) {
	o, err := buildOpts[volumeattach.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(volumeattach.Create(c.ctx, c.sc, serverID, o), "volumeAttachment")
}

// DetachVolume detaches a volume from a server.
func (c *Compute) DetachVolume(serverID, volumeID string) error {
	return done(volumeattach.Delete(c.ctx, c.sc, serverID, volumeID))
}

// Flavors lists all flavors (detailed).
func (c *Compute) Flavors() ([]Resource, error) {
	return readList(c.ctx, flavors.ListDetail(c.sc, flavors.ListOpts{}), "flavors", "")
}

// Flavor fetches a single flavor by ID.
func (c *Compute) Flavor(id string) (Resource, error) {
	return readObject(flavors.Get(c.ctx, c.sc, id), "flavor")
}

// Keypairs lists all keypairs.
func (c *Compute) Keypairs() ([]Resource, error) {
	return readList(c.ctx, keypairs.List(c.sc, keypairs.ListOpts{}), "keypairs", "keypair")
}

// Keypair fetches a single keypair by name.
func (c *Compute) Keypair(name string) (Resource, error) {
	return readObject(keypairs.Get(c.ctx, c.sc, name, nil), "keypair")
}

// CreateKeypair creates (or imports) a keypair from the option hash
// (name, public_key).
func (c *Compute) CreateKeypair(opts Resource) (Resource, error) {
	o, err := buildOpts[keypairs.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(keypairs.Create(c.ctx, c.sc, o), "keypair")
}

// DeleteKeypair deletes a keypair by name.
func (c *Compute) DeleteKeypair(name string) error {
	return done(keypairs.Delete(c.ctx, c.sc, name, nil))
}
