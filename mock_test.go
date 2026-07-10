// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockCloud is an in-process fake OpenStack cloud built on net/http/httptest.
// It serves a Keystone v3 token with a service catalog pointing every service
// back at itself, plus canned JSON for the core resources of the six wrapped
// services. No live cloud is ever required.
//
// When statusOverride is non-zero every *resource* response (not the auth
// exchange) is replaced by that status with bodyOverride, letting tests drive
// every error branch of the adapter without bespoke servers.
type mockCloud struct {
	*httptest.Server
	statusOverride int
	bodyOverride   string
	authStatus     int  // when non-zero, the auth exchange returns this status
	emptyCatalog   bool // when true, the issued token carries an empty catalog
	logf           func(string, ...any)
}

// canned single-object and list bodies keyed by the JSON wrappers gophercloud
// expects for each service.
const (
	serverObj  = `{"server":{"id":"srv-1","name":"web","status":"ACTIVE"}}`
	serverList = `{"servers":[{"id":"srv-1","name":"web","status":"ACTIVE"}]}`
	flavorObj  = `{"flavor":{"id":"1","name":"m1.tiny","ram":512,"vcpus":1,"disk":1}}`
	flavorList = `{"flavors":[{"id":"1","name":"m1.tiny","ram":512,"vcpus":1,"disk":1}]}`
	kpObj      = `{"keypair":{"name":"kp","fingerprint":"aa:bb","public_key":"ssh-rsa AAAA"}}`
	kpList     = `{"keypairs":[{"keypair":{"name":"kp","public_key":"ssh-rsa AAAA"}}]}`
	attachObj  = `{"volumeAttachment":{"id":"vol-1","volumeId":"vol-1","serverId":"srv-1","device":"/dev/vdb"}}`

	volObj   = `{"volume":{"id":"vol-1","name":"data","size":10,"status":"available"}}`
	volList  = `{"volumes":[{"id":"vol-1","name":"data","size":10,"status":"available"}]}`
	snapObj  = `{"snapshot":{"id":"snap-1","name":"snap","volume_id":"vol-1","size":10,"status":"available"}}`
	snapList = `{"snapshots":[{"id":"snap-1","name":"snap","volume_id":"vol-1","size":10,"status":"available"}]}`
	vtObj    = `{"volume_type":{"id":"vt-1","name":"lvm"}}`
	vtList   = `{"volume_types":[{"id":"vt-1","name":"lvm"}]}`

	imgObj  = `{"id":"img-1","name":"cirros","status":"active","visibility":"public"}`
	imgList = `{"images":[{"id":"img-1","name":"cirros","status":"active","visibility":"public"}]}`

	netObj   = `{"network":{"id":"net-1","name":"private","status":"ACTIVE","admin_state_up":true}}`
	netList  = `{"networks":[{"id":"net-1","name":"private","status":"ACTIVE","admin_state_up":true}]}`
	subObj   = `{"subnet":{"id":"sub-1","name":"sub","network_id":"net-1","cidr":"10.0.0.0/24","ip_version":4}}`
	subList  = `{"subnets":[{"id":"sub-1","name":"sub","network_id":"net-1","cidr":"10.0.0.0/24","ip_version":4}]}`
	portObj  = `{"port":{"id":"port-1","name":"p","network_id":"net-1","admin_state_up":true}}`
	portList = `{"ports":[{"id":"port-1","name":"p","network_id":"net-1","admin_state_up":true}]}`
	rtrObj   = `{"router":{"id":"rtr-1","name":"r","admin_state_up":true,"status":"ACTIVE"}}`
	rtrList  = `{"routers":[{"id":"rtr-1","name":"r","admin_state_up":true,"status":"ACTIVE"}]}`
	sgObj    = `{"security_group":{"id":"sg-1","name":"default"}}`
	sgList   = `{"security_groups":[{"id":"sg-1","name":"default"}]}`
	sgrObj   = `{"security_group_rule":{"id":"sgr-1","direction":"ingress","security_group_id":"sg-1"}}`
	sgrList  = `{"security_group_rules":[{"id":"sgr-1","direction":"ingress","security_group_id":"sg-1"}]}`
	fipObj   = `{"floatingip":{"id":"fip-1","floating_ip_address":"1.2.3.4","status":"ACTIVE"}}`
	fipList  = `{"floatingips":[{"id":"fip-1","floating_ip_address":"1.2.3.4","status":"ACTIVE"}]}`

	projObj = `{"project":{"id":"proj-1","name":"demo","domain_id":"default","enabled":true}}`
	prjList = `{"projects":[{"id":"proj-1","name":"demo","domain_id":"default","enabled":true}]}`
	usrObj  = `{"user":{"id":"user-1","name":"alice","domain_id":"default","enabled":true}}`
	usrList = `{"users":[{"id":"user-1","name":"alice","domain_id":"default","enabled":true}]}`
	roleObj = `{"role":{"id":"role-1","name":"admin"}}`
	rolList = `{"roles":[{"id":"role-1","name":"admin"}]}`
	domObj  = `{"domain":{"id":"dom-1","name":"Default","enabled":true}}`
	domList = `{"domains":[{"id":"dom-1","name":"Default","enabled":true}]}`

	contList = `[{"name":"logs","count":2,"bytes":100}]`
	objList  = `[{"name":"a.txt","bytes":11,"content_type":"text/plain","hash":"x","last_modified":"2026-01-01T00:00:00.000000"}]`
	objBody  = "hello world"
)

// startMock starts the fake cloud without connecting.
func startMock(t testing.TB) *mockCloud {
	t.Helper()
	mc := &mockCloud{logf: t.Logf}
	mc.Server = httptest.NewServer(mc.handler())
	t.Cleanup(mc.Server.Close)
	return mc
}

// options returns default RegionOne password-auth options for this cloud.
func (mc *mockCloud) options() Options {
	return Options{
		AuthURL:     mc.URL + "/identity/v3",
		Username:    "admin",
		Password:    "secret",
		ProjectName: "demo",
		DomainName:  "Default",
		Region:      "RegionOne",
	}
}

// newMockCloud starts the fake cloud and returns it together with an
// authenticated Connection scoped to RegionOne.
func newMockCloud(t testing.TB) (*mockCloud, *Connection) {
	t.Helper()
	mc := startMock(t)
	conn, err := Connect(context.Background(), mc.options())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	return mc, conn
}

// write emits a resource response, honouring any active override.
func (mc *mockCloud) write(w http.ResponseWriter, status int, body string) {
	if mc.statusOverride != 0 {
		status = mc.statusOverride
		body = mc.bodyOverride
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, body)
}

// collection registers a collection endpoint (GET list, POST create) at path
// and a per-item endpoint (GET/PUT/PATCH item, DELETE) at path/{id}.
func (mc *mockCloud) collection(mux *http.ServeMux, path, listBody, objBody string, createStatus int) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mc.write(w, createStatus, objBody)
			return
		}
		mc.write(w, 200, mc.pageList(r, listBody))
	})
	mux.HandleFunc(path+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mc.write(w, 204, "")
			return
		}
		mc.write(w, 200, objBody)
	})
}

// pageList returns an empty page for a marker follow-up request so marker-based
// pagers (Swift, Keystone, ...) terminate.
func (mc *mockCloud) pageList(r *http.Request, body string) string {
	if r.URL.Query().Get("marker") != "" {
		if strings.HasPrefix(strings.TrimSpace(body), "[") {
			return "[]"
		}
	}
	return body
}

func (mc *mockCloud) handler() http.Handler {
	mux := http.NewServeMux()

	// Keystone auth: issue a token whose catalog points every service back at
	// this same server.
	mux.HandleFunc("POST /identity/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		if mc.authStatus != 0 {
			w.WriteHeader(mc.authStatus)
			return
		}
		base := "http://" + r.Host
		w.Header().Set("X-Subject-Token", "faketoken")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, catalog(base, mc.emptyCatalog))
	})

	// Version-discovery documents served at each service root. gophercloud's
	// endpoint locator GETs these (expecting [200 300]) to confirm the catalog
	// endpoint speaks the requested major version. Object storage is skipped by
	// gophercloud, and identity is resolved without discovery.
	versions := func(id string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"versions":[{"id":"`+id+`","status":"CURRENT"}]}`)
		}
	}
	mux.HandleFunc("GET /compute/{$}", versions("v2.1"))
	mux.HandleFunc("GET /network/{$}", versions("v2.0"))
	mux.HandleFunc("GET /volume/{$}", versions("v3.0"))
	mux.HandleFunc("GET /image/{$}", versions("v2.0"))

	// Compute (Nova).
	mux.HandleFunc("GET /compute/servers/detail", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, mc.pageList(r, serverList))
	})
	mux.HandleFunc("POST /compute/servers", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 202, serverObj)
	})
	mux.HandleFunc("/compute/servers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mc.write(w, 204, "")
			return
		}
		mc.write(w, 200, serverObj)
	})
	mux.HandleFunc("POST /compute/servers/{id}/action", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 202, "")
	})
	mux.HandleFunc("POST /compute/servers/{id}/os-volume_attachments", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, attachObj)
	})
	mux.HandleFunc("DELETE /compute/servers/{id}/os-volume_attachments/{aid}", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 204, "")
	})
	mux.HandleFunc("GET /compute/flavors/detail", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, mc.pageList(r, flavorList))
	})
	mux.HandleFunc("GET /compute/flavors/{id}", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, flavorObj)
	})
	mux.HandleFunc("/compute/os-keypairs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mc.write(w, 201, kpObj)
			return
		}
		mc.write(w, 200, mc.pageList(r, kpList))
	})
	mux.HandleFunc("/compute/os-keypairs/{name}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mc.write(w, 204, "")
			return
		}
		mc.write(w, 200, kpObj)
	})

	// Block storage (Cinder).
	mux.HandleFunc("GET /volume/volumes/detail", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, mc.pageList(r, volList))
	})
	mux.HandleFunc("POST /volume/volumes", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 202, volObj)
	})
	mux.HandleFunc("/volume/volumes/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mc.write(w, 204, "")
			return
		}
		mc.write(w, 200, volObj)
	})
	mux.HandleFunc("GET /volume/snapshots/detail", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, mc.pageList(r, snapList))
	})
	mux.HandleFunc("POST /volume/snapshots", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 202, snapObj)
	})
	mux.HandleFunc("/volume/snapshots/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			mc.write(w, 204, "")
			return
		}
		mc.write(w, 200, snapObj)
	})
	mc.collection(mux, "/volume/types", vtList, vtObj, 200)

	// Image (Glance): single objects are unwrapped, list is under "images".
	mux.HandleFunc("/image/v2/images", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mc.write(w, 201, imgObj)
			return
		}
		mc.write(w, 200, imgList)
	})
	mux.HandleFunc("/image/v2/images/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			mc.write(w, 204, "")
		case http.MethodPut: // data upload
			mc.write(w, 204, "")
		default:
			mc.write(w, 200, imgObj)
		}
	})
	mux.HandleFunc("PUT /image/v2/images/{id}/file", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 204, "")
	})

	// Network (Neutron).
	mc.collection(mux, "/network/v2.0/networks", netList, netObj, 201)
	mc.collection(mux, "/network/v2.0/subnets", subList, subObj, 201)
	mc.collection(mux, "/network/v2.0/ports", portList, portObj, 201)
	mc.collection(mux, "/network/v2.0/routers", rtrList, rtrObj, 201)
	mc.collection(mux, "/network/v2.0/security-groups", sgList, sgObj, 201)
	mc.collection(mux, "/network/v2.0/security-group-rules", sgrList, sgrObj, 201)
	mc.collection(mux, "/network/v2.0/floatingips", fipList, fipObj, 201)

	// Identity (Keystone).
	mc.collection(mux, "/identity/v3/projects", prjList, projObj, 201)
	mc.collection(mux, "/identity/v3/users", usrList, usrObj, 201)
	mc.collection(mux, "/identity/v3/roles", rolList, roleObj, 201)
	mc.collection(mux, "/identity/v3/domains", domList, domObj, 201)

	// Object storage (Swift).
	mux.HandleFunc("GET /object/{$}", func(w http.ResponseWriter, r *http.Request) {
		mc.write(w, 200, mc.pageList(r, contList))
	})
	mux.HandleFunc("/object/{container}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			mc.write(w, 201, "")
		case http.MethodDelete:
			mc.write(w, 204, "")
		default:
			mc.write(w, 200, mc.pageList(r, objList))
		}
	})
	mux.HandleFunc("/object/{container}/{object}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			mc.write(w, 201, "")
		case http.MethodDelete:
			mc.write(w, 204, "")
		default:
			mc.write(w, 200, objBody)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if mc.logf != nil {
			mc.logf("UNMATCHED %s %s", r.Method, r.URL.Path)
		}
		http.NotFound(w, r)
	})

	return mux
}

// catalog renders a Keystone v3 token body whose service catalog points every
// wrapped service at base.
func catalog(base string, empty bool) string {
	entry := func(typ, url string) string {
		return `{"type":"` + typ + `","name":"` + typ + `","endpoints":[` +
			`{"id":"1","interface":"public","region":"RegionOne","region_id":"RegionOne","url":"` + url + `"}]}`
	}
	entries := ""
	if !empty {
		entries = strings.Join([]string{
			entry("identity", base+"/identity/v3/"),
			entry("compute", base+"/compute/"),
			entry("network", base+"/network/"),
			entry("block-storage", base+"/volume/"),
			entry("object-store", base+"/object/"),
			entry("image", base+"/image/"),
		}, ",")
	}
	return `{"token":{"methods":["password"],"expires_at":"2035-01-01T00:00:00.000000Z",` +
		`"user":{"id":"user-1","name":"admin","domain":{"id":"default","name":"Default"}},` +
		`"project":{"id":"proj-1","name":"demo","domain":{"id":"default","name":"Default"}},` +
		`"catalog":[` + entries + `]}}`
}
