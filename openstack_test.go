// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"strings"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
)

// wantMap returns an assertion that a (Resource, error) pair is error-free and
// carries map[key] == val. Usage: wantMap(t,"id","srv-1")(c.Server("srv-1")).
func wantMap(t *testing.T, key, val string) func(Resource, error) {
	return func(m Resource, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, _ := m[key].(string); got != val {
			t.Fatalf("map[%q] = %v, want %v", key, m[key], val)
		}
	}
}

// wantList asserts a ([]Resource, error) pair is error-free and has length n.
func wantList(t *testing.T, n int) func([]Resource, error) {
	return func(l []Resource, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(l) != n {
			t.Fatalf("list length = %d, want %d", len(l), n)
		}
	}
}

func noErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompute(t *testing.T) {
	_, conn := newMockCloud(t)
	c, err := conn.Compute()
	noErr(t, err)

	wantList(t, 1)(c.Servers())
	wantMap(t, "id", "srv-1")(c.Server("srv-1"))
	wantMap(t, "id", "srv-1")(c.CreateServer(Resource{"name": "web", "flavorRef": "1", "imageRef": "img-1"}))
	wantMap(t, "id", "srv-1")(c.UpdateServer("srv-1", Resource{"name": "web2"}))
	noErr(t, c.DeleteServer("srv-1"))
	noErr(t, c.StartServer("srv-1"))
	noErr(t, c.StopServer("srv-1"))
	noErr(t, c.RebootServer("srv-1", "SOFT"))
	noErr(t, c.RebootServer("srv-1", "HARD"))
	wantMap(t, "id", "vol-1")(c.AttachVolume("srv-1", Resource{"volumeId": "vol-1"}))
	noErr(t, c.DetachVolume("srv-1", "vol-1"))

	wantList(t, 1)(c.Flavors())
	wantMap(t, "id", "1")(c.Flavor("1"))

	wantList(t, 1)(c.Keypairs())
	wantMap(t, "name", "kp")(c.Keypair("kp"))
	wantMap(t, "name", "kp")(c.CreateKeypair(Resource{"name": "kp"}))
	noErr(t, c.DeleteKeypair("kp"))
}

func TestNetwork(t *testing.T) {
	_, conn := newMockCloud(t)
	n, err := conn.Network()
	noErr(t, err)

	wantList(t, 1)(n.Networks())
	wantMap(t, "id", "net-1")(n.GetNetwork("net-1"))
	wantMap(t, "id", "net-1")(n.CreateNetwork(Resource{"name": "private"}))
	wantMap(t, "id", "net-1")(n.UpdateNetwork("net-1", Resource{"name": "p2"}))
	noErr(t, n.DeleteNetwork("net-1"))

	wantList(t, 1)(n.Subnets())
	wantMap(t, "id", "sub-1")(n.GetSubnet("sub-1"))
	wantMap(t, "id", "sub-1")(n.CreateSubnet(Resource{"network_id": "net-1", "cidr": "10.0.0.0/24", "ip_version": 4}))
	wantMap(t, "id", "sub-1")(n.UpdateSubnet("sub-1", Resource{"name": "s2"}))
	noErr(t, n.DeleteSubnet("sub-1"))

	wantList(t, 1)(n.Ports())
	wantMap(t, "id", "port-1")(n.GetPort("port-1"))
	wantMap(t, "id", "port-1")(n.CreatePort(Resource{"network_id": "net-1"}))
	wantMap(t, "id", "port-1")(n.UpdatePort("port-1", Resource{"name": "p2"}))
	noErr(t, n.DeletePort("port-1"))

	wantList(t, 1)(n.Routers())
	wantMap(t, "id", "rtr-1")(n.GetRouter("rtr-1"))
	wantMap(t, "id", "rtr-1")(n.CreateRouter(Resource{"name": "r"}))
	wantMap(t, "id", "rtr-1")(n.UpdateRouter("rtr-1", Resource{"name": "r2"}))
	noErr(t, n.DeleteRouter("rtr-1"))

	wantList(t, 1)(n.SecurityGroups())
	wantMap(t, "id", "sg-1")(n.GetSecurityGroup("sg-1"))
	wantMap(t, "id", "sg-1")(n.CreateSecurityGroup(Resource{"name": "default"}))
	wantMap(t, "id", "sg-1")(n.UpdateSecurityGroup("sg-1", Resource{"name": "d2"}))
	noErr(t, n.DeleteSecurityGroup("sg-1"))

	wantList(t, 1)(n.SecurityGroupRules())
	wantMap(t, "id", "sgr-1")(n.GetSecurityGroupRule("sgr-1"))
	wantMap(t, "id", "sgr-1")(n.CreateSecurityGroupRule(Resource{"direction": "ingress", "ethertype": "IPv4", "security_group_id": "sg-1"}))
	noErr(t, n.DeleteSecurityGroupRule("sgr-1"))

	wantList(t, 1)(n.FloatingIPs())
	wantMap(t, "id", "fip-1")(n.GetFloatingIP("fip-1"))
	wantMap(t, "id", "fip-1")(n.CreateFloatingIP(Resource{"floating_network_id": "net-1"}))
	wantMap(t, "id", "fip-1")(n.UpdateFloatingIP("fip-1", Resource{"port_id": "port-1"}))
	noErr(t, n.DeleteFloatingIP("fip-1"))
}

func TestBlockStorage(t *testing.T) {
	_, conn := newMockCloud(t)
	b, err := conn.BlockStorage()
	noErr(t, err)

	wantList(t, 1)(b.Volumes())
	wantMap(t, "id", "vol-1")(b.GetVolume("vol-1"))
	wantMap(t, "id", "vol-1")(b.CreateVolume(Resource{"size": 10, "name": "data"}))
	wantMap(t, "id", "vol-1")(b.UpdateVolume("vol-1", Resource{"name": "data2"}))
	noErr(t, b.DeleteVolume("vol-1"))

	wantList(t, 1)(b.Snapshots())
	wantMap(t, "id", "snap-1")(b.GetSnapshot("snap-1"))
	wantMap(t, "id", "snap-1")(b.CreateSnapshot(Resource{"volume_id": "vol-1", "name": "snap"}))
	wantMap(t, "id", "snap-1")(b.UpdateSnapshot("snap-1", Resource{"name": "snap2"}))
	noErr(t, b.DeleteSnapshot("snap-1"))

	wantList(t, 1)(b.VolumeTypes())
	wantMap(t, "id", "vt-1")(b.GetVolumeType("vt-1"))
	wantMap(t, "id", "vt-1")(b.CreateVolumeType(Resource{"name": "lvm"}))
	wantMap(t, "id", "vt-1")(b.UpdateVolumeType("vt-1", Resource{"name": "lvm2"}))
	noErr(t, b.DeleteVolumeType("vt-1"))
}

func TestImage(t *testing.T) {
	_, conn := newMockCloud(t)
	i, err := conn.Image()
	noErr(t, err)

	wantList(t, 1)(i.Images())
	wantMap(t, "id", "img-1")(i.GetImage("img-1"))
	wantMap(t, "id", "img-1")(i.CreateImage(Resource{"name": "cirros"}))
	wantMap(t, "id", "img-1")(i.UpdateImage("img-1", images.UpdateOpts{images.ReplaceImageName{NewName: "cirros2"}}))
	noErr(t, i.UploadImage("img-1", strings.NewReader("data")))
	noErr(t, i.DeleteImage("img-1"))
}

func TestIdentity(t *testing.T) {
	_, conn := newMockCloud(t)
	id, err := conn.Identity()
	noErr(t, err)

	wantList(t, 1)(id.Projects())
	wantMap(t, "id", "proj-1")(id.GetProject("proj-1"))
	wantMap(t, "id", "proj-1")(id.CreateProject(Resource{"name": "demo"}))
	wantMap(t, "id", "proj-1")(id.UpdateProject("proj-1", Resource{"description": "d"}))
	noErr(t, id.DeleteProject("proj-1"))

	wantList(t, 1)(id.Users())
	wantMap(t, "id", "user-1")(id.GetUser("user-1"))
	wantMap(t, "id", "user-1")(id.CreateUser(Resource{"name": "alice"}))
	wantMap(t, "id", "user-1")(id.UpdateUser("user-1", Resource{"email": "a@b.c"}))
	noErr(t, id.DeleteUser("user-1"))

	wantList(t, 1)(id.Roles())
	wantMap(t, "id", "role-1")(id.GetRole("role-1"))
	wantMap(t, "id", "role-1")(id.CreateRole(Resource{"name": "admin"}))
	wantMap(t, "id", "role-1")(id.UpdateRole("role-1", Resource{"description": "d"}))
	noErr(t, id.DeleteRole("role-1"))

	wantList(t, 1)(id.Domains())
	wantMap(t, "id", "dom-1")(id.GetDomain("dom-1"))
	wantMap(t, "id", "dom-1")(id.CreateDomain(Resource{"name": "D"}))
	wantMap(t, "id", "dom-1")(id.UpdateDomain("dom-1", Resource{"description": "d"}))
	noErr(t, id.DeleteDomain("dom-1"))
}

func TestObjectStorage(t *testing.T) {
	_, conn := newMockCloud(t)
	o, err := conn.ObjectStorage()
	noErr(t, err)

	wantList(t, 1)(o.Containers())
	noErr(t, o.CreateContainer("logs"))
	noErr(t, o.DeleteContainer("logs"))

	wantList(t, 1)(o.Objects("logs"))
	noErr(t, o.PutObject("logs", "a.txt", strings.NewReader("hello world")))
	content, err := o.GetObject("logs", "a.txt")
	noErr(t, err)
	if string(content) != objBody {
		t.Fatalf("GetObject content = %q, want %q", content, objBody)
	}
	noErr(t, o.DeleteObject("logs", "a.txt"))
}
