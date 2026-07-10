// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/v2"
)

// TestErrorMapping drives every HTTP status onto its typed error and checks the
// matching Is* predicate, plus the Error/Status/Unwrap accessors.
func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status  int
		is      func(error) bool
		status0 int
	}{
		{http.StatusNotFound, IsNotFound, http.StatusNotFound},
		{http.StatusUnauthorized, IsAuth, http.StatusUnauthorized},
		{http.StatusForbidden, IsForbidden, http.StatusForbidden},
		{http.StatusConflict, IsConflict, http.StatusConflict},
		{http.StatusBadRequest, IsBadRequest, http.StatusBadRequest},
	}
	for _, tc := range cases {
		mc, conn := newMockCloud(t)
		c, err := conn.Compute()
		noErr(t, err)
		mc.statusOverride = tc.status
		mc.bodyOverride = `{"error":"boom"}`

		_, gotErr := c.Server("srv-1")
		if gotErr == nil {
			t.Fatalf("status %d: expected error", tc.status)
		}
		if !tc.is(gotErr) {
			t.Fatalf("status %d: predicate did not match: %v", tc.status, gotErr)
		}
		var apiErr APIError
		if !errors.As(gotErr, &apiErr) {
			t.Fatalf("status %d: not an APIError", tc.status)
		}
		if apiErr.Status() != tc.status0 {
			t.Fatalf("Status() = %d, want %d", apiErr.Status(), tc.status0)
		}
		if apiErr.Error() == "" {
			t.Fatalf("empty Error() string")
		}
		if apiErr.Unwrap() == nil {
			t.Fatalf("expected wrapped cause")
		}
	}
}

// TestGenericError checks that an unmapped status (500) becomes a generic
// *Error carrying the status code.
func TestGenericError(t *testing.T) {
	mc, conn := newMockCloud(t)
	c, err := conn.Compute()
	noErr(t, err)
	mc.statusOverride = http.StatusInternalServerError
	mc.bodyOverride = "boom"

	_, gotErr := c.Server("srv-1")
	if IsNotFound(gotErr) || IsAuth(gotErr) || IsConflict(gotErr) || IsBadRequest(gotErr) || IsForbidden(gotErr) {
		t.Fatalf("500 should be generic, got %v", gotErr)
	}
	var e *Error
	if !errors.As(gotErr, &e) {
		t.Fatalf("expected *Error, got %T", gotErr)
	}
	if e.Status() != http.StatusInternalServerError {
		t.Fatalf("Status() = %d, want 500", e.Status())
	}
}

func TestMapErrorNil(t *testing.T) {
	if mapError(nil) != nil {
		t.Fatal("mapError(nil) should be nil")
	}
	if authError(nil) != nil {
		t.Fatal("authError(nil) should be nil")
	}
}

// TestListErrorBranches covers readList's AllPages error and readObject's
// missing-key branch.
func TestListErrorBranches(t *testing.T) {
	mc, conn := newMockCloud(t)
	c, err := conn.Compute()
	noErr(t, err)

	// AllPages error.
	mc.statusOverride = http.StatusInternalServerError
	if _, err := c.Servers(); err == nil {
		t.Fatal("expected list error")
	}

	// Object body missing the expected envelope key.
	mc.statusOverride = http.StatusOK
	mc.bodyOverride = `{"unexpected":{}}`
	if _, err := c.Server("srv-1"); err == nil {
		t.Fatal("expected unwrap error")
	}
}

// TestCollectDirect covers collect's []byte-error and default branches directly.
func TestCollectDirect(t *testing.T) {
	if _, err := collect([]byte("{not-json"), "", ""); err == nil {
		t.Fatal("expected JSON error")
	}
	if _, err := collect(42, "x", ""); err == nil {
		t.Fatal("expected unexpected-body error")
	}
	// []any body path.
	got, err := collect([]any{map[string]any{"id": "x"}}, "", "")
	noErr(t, err)
	if len(got) != 1 || got[0]["id"] != "x" {
		t.Fatalf("collect []any = %v", got)
	}
}

// TestBuildOptsErrors forces the marshalling seams to fail, covering buildOpts'
// two error branches and every mutating method's post-buildOpts error return.
func TestBuildOptsErrors(t *testing.T) {
	_, conn := newMockCloud(t)
	boom := errors.New("boom")
	savedM, savedU := marshalJSON, unmarshalJSON

	// marshal failure (covers buildOpts marshal branch).
	marshalJSON = func(any) ([]byte, error) { return nil, boom }
	if _, err := buildOpts[map[string]any](Resource{}); err == nil {
		t.Fatal("expected marshal error")
	}
	marshalJSON = savedM

	// unmarshal failure, exercised through every wrapper that builds opts.
	unmarshalJSON = func([]byte, any) error { return boom }
	defer func() { unmarshalJSON = savedU }()

	mutators := gatherMutators(t, conn)
	for name, fn := range mutators {
		if err := fn(); err == nil {
			t.Fatalf("%s: expected buildOpts error", name)
		}
	}
}

// gatherMutators returns a closure per Create/Update method (those that build
// opts from a hash) so a single failing-seam run covers them all.
func gatherMutators(t *testing.T, conn *Connection) map[string]func() error {
	t.Helper()
	c, err := conn.Compute()
	noErr(t, err)
	n, err := conn.Network()
	noErr(t, err)
	b, err := conn.BlockStorage()
	noErr(t, err)
	id, err := conn.Identity()
	noErr(t, err)

	errOf := func(_ Resource, e error) error { return e }
	return map[string]func() error{
		"CreateServer":       func() error { return errOf(c.CreateServer(Resource{})) },
		"UpdateServer":       func() error { return errOf(c.UpdateServer("x", Resource{})) },
		"AttachVolume":       func() error { return errOf(c.AttachVolume("x", Resource{})) },
		"CreateKeypair":      func() error { return errOf(c.CreateKeypair(Resource{})) },
		"CreateNetwork":      func() error { return errOf(n.CreateNetwork(Resource{})) },
		"UpdateNetwork":      func() error { return errOf(n.UpdateNetwork("x", Resource{})) },
		"CreateSubnet":       func() error { return errOf(n.CreateSubnet(Resource{})) },
		"UpdateSubnet":       func() error { return errOf(n.UpdateSubnet("x", Resource{})) },
		"CreatePort":         func() error { return errOf(n.CreatePort(Resource{})) },
		"UpdatePort":         func() error { return errOf(n.UpdatePort("x", Resource{})) },
		"CreateRouter":       func() error { return errOf(n.CreateRouter(Resource{})) },
		"UpdateRouter":       func() error { return errOf(n.UpdateRouter("x", Resource{})) },
		"CreateSecGroup":     func() error { return errOf(n.CreateSecurityGroup(Resource{})) },
		"UpdateSecGroup":     func() error { return errOf(n.UpdateSecurityGroup("x", Resource{})) },
		"CreateSecGroupRule": func() error { return errOf(n.CreateSecurityGroupRule(Resource{})) },
		"CreateFloatingIP":   func() error { return errOf(n.CreateFloatingIP(Resource{})) },
		"UpdateFloatingIP":   func() error { return errOf(n.UpdateFloatingIP("x", Resource{})) },
		"CreateVolume":       func() error { return errOf(b.CreateVolume(Resource{})) },
		"UpdateVolume":       func() error { return errOf(b.UpdateVolume("x", Resource{})) },
		"CreateSnapshot":     func() error { return errOf(b.CreateSnapshot(Resource{})) },
		"UpdateSnapshot":     func() error { return errOf(b.UpdateSnapshot("x", Resource{})) },
		"CreateVolumeType":   func() error { return errOf(b.CreateVolumeType(Resource{})) },
		"UpdateVolumeType":   func() error { return errOf(b.UpdateVolumeType("x", Resource{})) },
		"CreateProject":      func() error { return errOf(id.CreateProject(Resource{})) },
		"UpdateProject":      func() error { return errOf(id.UpdateProject("x", Resource{})) },
		"CreateUser":         func() error { return errOf(id.CreateUser(Resource{})) },
		"UpdateUser":         func() error { return errOf(id.UpdateUser("x", Resource{})) },
		"CreateRole":         func() error { return errOf(id.CreateRole(Resource{})) },
		"UpdateRole":         func() error { return errOf(id.UpdateRole("x", Resource{})) },
		"CreateDomain":       func() error { return errOf(id.CreateDomain(Resource{})) },
		"UpdateDomain":       func() error { return errOf(id.UpdateDomain("x", Resource{})) },
		"CreateImage":        func() error { return errOf(conn.mustImage(t).CreateImage(Resource{})) },
	}
}

func (c *Connection) mustImage(t *testing.T) *Image {
	t.Helper()
	i, err := c.Image()
	noErr(t, err)
	return i
}

// TestGetObjectError covers GetObject's download-error branch.
func TestGetObjectError(t *testing.T) {
	mc, conn := newMockCloud(t)
	o, err := conn.ObjectStorage()
	noErr(t, err)
	mc.statusOverride = http.StatusNotFound
	if _, err := o.GetObject("logs", "missing.txt"); !IsNotFound(err) {
		t.Fatalf("GetObject: want NotFound, got %v", err)
	}
}

func TestConnectErrors(t *testing.T) {
	mc := startMock(t)

	// Missing AuthURL.
	if _, err := Connect(context.Background(), Options{}); !IsAuth(err) {
		t.Fatalf("missing auth_url: want AuthError, got %v", err)
	}

	// newProviderClient failure via the seam.
	saved := newProviderClient
	newProviderClient = func(string) (*gophercloud.ProviderClient, error) { return nil, errors.New("bad url") }
	if _, err := Connect(context.Background(), mc.options()); !IsAuth(err) {
		t.Fatalf("provider error: want AuthError, got %v", err)
	}
	newProviderClient = saved

	// Real 401 from Keystone -> mapError yields AuthError (authError IsAuth branch).
	mc.authStatus = http.StatusUnauthorized
	if _, err := Connect(context.Background(), mc.options()); !IsAuth(err) {
		t.Fatalf("401 auth: want AuthError, got %v", err)
	}

	// A 500 during auth maps to a generic error, which authError re-wraps as an
	// AuthError (authError non-auth branch).
	mc.authStatus = http.StatusInternalServerError
	if _, err := Connect(context.Background(), mc.options()); !IsAuth(err) {
		t.Fatalf("500 auth: want AuthError, got %v", err)
	}
	mc.authStatus = 0

	// HTTPClient and Transport option branches (both must still connect).
	optsHTTP := mc.options()
	optsHTTP.HTTPClient = &http.Client{}
	if _, err := Connect(context.Background(), optsHTTP); err != nil {
		t.Fatalf("HTTPClient option: %v", err)
	}
	optsTr := mc.options()
	optsTr.Transport = http.DefaultTransport
	if _, err := Connect(context.Background(), optsTr); err != nil {
		t.Fatalf("Transport option: %v", err)
	}
}

// TestServiceAccessorErrors covers newService's error branch and each accessor's
// error return by authenticating against a cloud with an empty catalog.
func TestServiceAccessorErrors(t *testing.T) {
	mc := startMock(t)
	mc.emptyCatalog = true
	conn, err := Connect(context.Background(), mc.options())
	noErr(t, err)

	if _, err := conn.Compute(); err == nil {
		t.Fatal("Compute: expected endpoint error")
	}
	if _, err := conn.Network(); err == nil {
		t.Fatal("Network: expected endpoint error")
	}
	if _, err := conn.BlockStorage(); err == nil {
		t.Fatal("BlockStorage: expected endpoint error")
	}
	if _, err := conn.ObjectStorage(); err == nil {
		t.Fatal("ObjectStorage: expected endpoint error")
	}
	if _, err := conn.Image(); err == nil {
		t.Fatal("Image: expected endpoint error")
	}
	if _, err := conn.Identity(); err == nil {
		t.Fatal("Identity: expected endpoint error")
	}
}
