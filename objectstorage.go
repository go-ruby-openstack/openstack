// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"io"

	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/objects"
)

// ObjectStorage is the Swift service accessor: containers and objects. Swift
// list responses are bare JSON arrays, so the items come back without an
// envelope key.
type ObjectStorage struct{ *service }

// Containers lists all containers (with metadata).
func (o *ObjectStorage) Containers() ([]Resource, error) {
	return readList(o.ctx, containers.List(o.sc, containers.ListOpts{}), "", "")
}

// CreateContainer creates (or updates) a container by name.
func (o *ObjectStorage) CreateContainer(name string) error {
	return mapError(containers.Create(o.ctx, o.sc, name, containers.CreateOpts{}).Err)
}

// DeleteContainer deletes an empty container by name.
func (o *ObjectStorage) DeleteContainer(name string) error {
	return mapError(containers.Delete(o.ctx, o.sc, name).Err)
}

// Objects lists the objects (with metadata) in a container.
func (o *ObjectStorage) Objects(container string) ([]Resource, error) {
	return readList(o.ctx, objects.List(o.sc, container, objects.ListOpts{}), "", "")
}

// PutObject uploads an object's content into a container.
func (o *ObjectStorage) PutObject(container, name string, content io.Reader) error {
	return mapError(objects.Create(o.ctx, o.sc, container, name, objects.CreateOpts{Content: content}).Err)
}

// GetObject downloads an object's content from a container.
func (o *ObjectStorage) GetObject(container, name string) ([]byte, error) {
	r := objects.Download(o.ctx, o.sc, container, name, nil)
	content, err := r.ExtractContent()
	if err != nil {
		return nil, mapError(err)
	}
	return content, nil
}

// DeleteObject deletes an object from a container.
func (o *ObjectStorage) DeleteObject(container, name string) error {
	return mapError(objects.Delete(o.ctx, o.sc, container, name, nil).Err)
}
