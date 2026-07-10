// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

// Options carries the Keystone v3 authentication parameters plus the pluggable
// HTTP transport. It mirrors the keyword arguments a Ruby caller passes to
// OpenStack::Connection.new(auth_url:, username:, password:, project_name:,
// domain_name:, region:, ...). Three authentication styles are supported:
//
//   - password: Username (or UserID) + Password + a domain;
//   - token: Token (an existing Keystone token, scoped by the project fields);
//   - application credential: ApplicationCredentialID/Name + Secret.
type Options struct {
	// AuthURL is the Keystone v3 identity endpoint ("OS_AUTH_URL"), e.g.
	// https://keystone.example.com/v3. Required.
	AuthURL string

	// Username / UserID + Password select password authentication.
	Username string
	UserID   string
	Password string

	// Token authenticates with an existing Keystone token ID.
	Token string

	// Application-credential authentication.
	ApplicationCredentialID     string
	ApplicationCredentialName   string
	ApplicationCredentialSecret string

	// ProjectName / ProjectID and DomainName / DomainID scope the token.
	ProjectName string
	ProjectID   string
	DomainName  string
	DomainID    string

	// Region selects which catalog endpoint to use for every service. Leave
	// empty when the cloud publishes a single region.
	Region string

	// AllowReauth lets gophercloud transparently re-authenticate when the token
	// expires.
	AllowReauth bool

	// Transport is the injectable HTTP transport seam. When non-nil it becomes
	// the RoundTripper of the underlying gophercloud client, letting tests and
	// rbgo supply a fake so no live cloud is needed. Mutually exclusive with
	// HTTPClient (HTTPClient wins when both are set).
	Transport http.RoundTripper

	// HTTPClient, when non-nil, is used verbatim as the gophercloud HTTP client.
	HTTPClient *http.Client
}

// Connection is an authenticated OpenStack session. It owns a gophercloud
// ProviderClient (carrying the token and service catalog) and hands out the
// per-service accessors below.
type Connection struct {
	provider *gophercloud.ProviderClient
	region   string
	ctx      context.Context
}

// Seams over gophercloud's package-level constructors so tests can exercise the
// error branches without a broken cloud.
var (
	newProviderClient = openstack.NewClient
	authenticate      = openstack.AuthenticateV3
)

// Connect authenticates against Keystone v3 and returns a ready Connection.
// A missing AuthURL, or any authentication failure, yields an *AuthError.
func Connect(ctx context.Context, opts Options) (*Connection, error) {
	if opts.AuthURL == "" {
		return nil, &AuthError{baseError{message: "openstack: auth_url is required"}}
	}

	provider, err := newProviderClient(opts.AuthURL)
	if err != nil {
		return nil, authError(err)
	}

	switch {
	case opts.HTTPClient != nil:
		provider.HTTPClient = *opts.HTTPClient
	case opts.Transport != nil:
		provider.HTTPClient = http.Client{Transport: opts.Transport}
	}

	authOpts := &gophercloud.AuthOptions{
		IdentityEndpoint:            opts.AuthURL,
		Username:                    opts.Username,
		UserID:                      opts.UserID,
		Password:                    opts.Password,
		TokenID:                     opts.Token,
		ApplicationCredentialID:     opts.ApplicationCredentialID,
		ApplicationCredentialName:   opts.ApplicationCredentialName,
		ApplicationCredentialSecret: opts.ApplicationCredentialSecret,
		DomainName:                  opts.DomainName,
		DomainID:                    opts.DomainID,
		TenantName:                  opts.ProjectName,
		TenantID:                    opts.ProjectID,
		AllowReauth:                 opts.AllowReauth,
	}
	if opts.ProjectName != "" || opts.ProjectID != "" {
		authOpts.Scope = &gophercloud.AuthScope{
			ProjectName: opts.ProjectName,
			ProjectID:   opts.ProjectID,
			DomainName:  opts.DomainName,
			DomainID:    opts.DomainID,
		}
	}

	if err := authenticate(ctx, provider, authOpts, gophercloud.EndpointOpts{}); err != nil {
		return nil, authError(err)
	}

	return &Connection{provider: provider, region: opts.region(), ctx: ctx}, nil
}

func (o Options) region() string { return o.Region }

// endpointOpts is the per-service endpoint selector, honouring the connection's
// region.
func (c *Connection) endpointOpts() gophercloud.EndpointOpts {
	return gophercloud.EndpointOpts{Region: c.region}
}

// serviceFactory is the shape of gophercloud's New${Service}V${N} constructors.
type serviceFactory func(*gophercloud.ProviderClient, gophercloud.EndpointOpts) (*gophercloud.ServiceClient, error)

// newService builds a service client through the given gophercloud factory,
// wrapping any catalog/endpoint error in the typed tree. Centralising the error
// branch here keeps every accessor a single line.
func newService(c *Connection, factory serviceFactory) (*service, error) {
	sc, err := factory(c.provider, c.endpointOpts())
	if err != nil {
		return nil, mapError(err)
	}
	return &service{sc: sc, ctx: c.ctx}, nil
}

// service is the shared base embedded by every typed service accessor.
type service struct {
	sc  *gophercloud.ServiceClient
	ctx context.Context
}

// Compute returns the Nova (compute) service accessor.
func (c *Connection) Compute() (*Compute, error) {
	s, err := newService(c, openstack.NewComputeV2)
	if err != nil {
		return nil, err
	}
	return &Compute{s}, nil
}

// Network returns the Neutron (network) service accessor.
func (c *Connection) Network() (*Network, error) {
	s, err := newService(c, openstack.NewNetworkV2)
	if err != nil {
		return nil, err
	}
	return &Network{s}, nil
}

// BlockStorage returns the Cinder (block storage v3) service accessor.
func (c *Connection) BlockStorage() (*BlockStorage, error) {
	s, err := newService(c, openstack.NewBlockStorageV3)
	if err != nil {
		return nil, err
	}
	return &BlockStorage{s}, nil
}

// ObjectStorage returns the Swift (object storage) service accessor.
func (c *Connection) ObjectStorage() (*ObjectStorage, error) {
	s, err := newService(c, openstack.NewObjectStorageV1)
	if err != nil {
		return nil, err
	}
	return &ObjectStorage{s}, nil
}

// Image returns the Glance (image v2) service accessor.
func (c *Connection) Image() (*Image, error) {
	s, err := newService(c, openstack.NewImageV2)
	if err != nil {
		return nil, err
	}
	return &Image{s}, nil
}

// Identity returns the Keystone (identity v3) service accessor.
func (c *Connection) Identity() (*Identity, error) {
	s, err := newService(c, openstack.NewIdentityV3)
	if err != nil {
		return nil, err
	}
	return &Identity{s}, nil
}
