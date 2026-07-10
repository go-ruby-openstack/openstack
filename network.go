// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
)

// Network is the Neutron service accessor: networks, subnets, ports, routers,
// security groups and their rules, and floating IPs.
type Network struct{ *service }

// Networks lists all networks.
func (n *Network) Networks() ([]Resource, error) {
	return readList(n.ctx, networks.List(n.sc, networks.ListOpts{}), "networks", "")
}

// GetNetwork fetches a network by ID.
func (n *Network) GetNetwork(id string) (Resource, error) {
	return readObject(networks.Get(n.ctx, n.sc, id), "network")
}

// CreateNetwork creates a network from the option hash (name, admin_state_up, ...).
func (n *Network) CreateNetwork(opts Resource) (Resource, error) {
	o, err := buildOpts[networks.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(networks.Create(n.ctx, n.sc, o), "network")
}

// UpdateNetwork updates a network.
func (n *Network) UpdateNetwork(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[networks.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(networks.Update(n.ctx, n.sc, id, o), "network")
}

// DeleteNetwork deletes a network by ID.
func (n *Network) DeleteNetwork(id string) error {
	return done(networks.Delete(n.ctx, n.sc, id))
}

// Subnets lists all subnets.
func (n *Network) Subnets() ([]Resource, error) {
	return readList(n.ctx, subnets.List(n.sc, subnets.ListOpts{}), "subnets", "")
}

// GetSubnet fetches a subnet by ID.
func (n *Network) GetSubnet(id string) (Resource, error) {
	return readObject(subnets.Get(n.ctx, n.sc, id), "subnet")
}

// CreateSubnet creates a subnet (network_id, cidr, ip_version, ...).
func (n *Network) CreateSubnet(opts Resource) (Resource, error) {
	o, err := buildOpts[subnets.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(subnets.Create(n.ctx, n.sc, o), "subnet")
}

// UpdateSubnet updates a subnet.
func (n *Network) UpdateSubnet(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[subnets.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(subnets.Update(n.ctx, n.sc, id, o), "subnet")
}

// DeleteSubnet deletes a subnet by ID.
func (n *Network) DeleteSubnet(id string) error {
	return done(subnets.Delete(n.ctx, n.sc, id))
}

// Ports lists all ports.
func (n *Network) Ports() ([]Resource, error) {
	return readList(n.ctx, ports.List(n.sc, ports.ListOpts{}), "ports", "")
}

// GetPort fetches a port by ID.
func (n *Network) GetPort(id string) (Resource, error) {
	return readObject(ports.Get(n.ctx, n.sc, id), "port")
}

// CreatePort creates a port (network_id, ...).
func (n *Network) CreatePort(opts Resource) (Resource, error) {
	o, err := buildOpts[ports.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(ports.Create(n.ctx, n.sc, o), "port")
}

// UpdatePort updates a port.
func (n *Network) UpdatePort(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[ports.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(ports.Update(n.ctx, n.sc, id, o), "port")
}

// DeletePort deletes a port by ID.
func (n *Network) DeletePort(id string) error {
	return done(ports.Delete(n.ctx, n.sc, id))
}

// Routers lists all routers.
func (n *Network) Routers() ([]Resource, error) {
	return readList(n.ctx, routers.List(n.sc, routers.ListOpts{}), "routers", "")
}

// GetRouter fetches a router by ID.
func (n *Network) GetRouter(id string) (Resource, error) {
	return readObject(routers.Get(n.ctx, n.sc, id), "router")
}

// CreateRouter creates a router (name, admin_state_up, external_gateway_info, ...).
func (n *Network) CreateRouter(opts Resource) (Resource, error) {
	o, err := buildOpts[routers.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(routers.Create(n.ctx, n.sc, o), "router")
}

// UpdateRouter updates a router.
func (n *Network) UpdateRouter(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[routers.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(routers.Update(n.ctx, n.sc, id, o), "router")
}

// DeleteRouter deletes a router by ID.
func (n *Network) DeleteRouter(id string) error {
	return done(routers.Delete(n.ctx, n.sc, id))
}

// SecurityGroups lists all security groups.
func (n *Network) SecurityGroups() ([]Resource, error) {
	return readList(n.ctx, groups.List(n.sc, groups.ListOpts{}), "security_groups", "")
}

// GetSecurityGroup fetches a security group by ID.
func (n *Network) GetSecurityGroup(id string) (Resource, error) {
	return readObject(groups.Get(n.ctx, n.sc, id), "security_group")
}

// CreateSecurityGroup creates a security group (name, description).
func (n *Network) CreateSecurityGroup(opts Resource) (Resource, error) {
	o, err := buildOpts[groups.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(groups.Create(n.ctx, n.sc, o), "security_group")
}

// UpdateSecurityGroup updates a security group.
func (n *Network) UpdateSecurityGroup(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[groups.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(groups.Update(n.ctx, n.sc, id, o), "security_group")
}

// DeleteSecurityGroup deletes a security group by ID.
func (n *Network) DeleteSecurityGroup(id string) error {
	return done(groups.Delete(n.ctx, n.sc, id))
}

// SecurityGroupRules lists all security group rules.
func (n *Network) SecurityGroupRules() ([]Resource, error) {
	return readList(n.ctx, rules.List(n.sc, rules.ListOpts{}), "security_group_rules", "")
}

// GetSecurityGroupRule fetches a security group rule by ID.
func (n *Network) GetSecurityGroupRule(id string) (Resource, error) {
	return readObject(rules.Get(n.ctx, n.sc, id), "security_group_rule")
}

// CreateSecurityGroupRule creates a rule (security_group_id, direction,
// protocol, port_range_min, ...).
func (n *Network) CreateSecurityGroupRule(opts Resource) (Resource, error) {
	o, err := buildOpts[rules.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(rules.Create(n.ctx, n.sc, o), "security_group_rule")
}

// DeleteSecurityGroupRule deletes a rule by ID.
func (n *Network) DeleteSecurityGroupRule(id string) error {
	return done(rules.Delete(n.ctx, n.sc, id))
}

// FloatingIPs lists all floating IPs.
func (n *Network) FloatingIPs() ([]Resource, error) {
	return readList(n.ctx, floatingips.List(n.sc, floatingips.ListOpts{}), "floatingips", "")
}

// GetFloatingIP fetches a floating IP by ID.
func (n *Network) GetFloatingIP(id string) (Resource, error) {
	return readObject(floatingips.Get(n.ctx, n.sc, id), "floatingip")
}

// CreateFloatingIP allocates a floating IP (floating_network_id, ...).
func (n *Network) CreateFloatingIP(opts Resource) (Resource, error) {
	o, err := buildOpts[floatingips.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(floatingips.Create(n.ctx, n.sc, o), "floatingip")
}

// UpdateFloatingIP associates/disassociates a floating IP (port_id).
func (n *Network) UpdateFloatingIP(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[floatingips.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(floatingips.Update(n.ctx, n.sc, id, o), "floatingip")
}

// DeleteFloatingIP releases a floating IP by ID.
func (n *Network) DeleteFloatingIP(id string) error {
	return done(floatingips.Delete(n.ctx, n.sc, id))
}
