// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"io"

	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/imagedata"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
)

// Image is the Glance (v2) service accessor: image records and their data.
// Glance returns image objects unwrapped, so single results carry no envelope.
type Image struct{ *service }

// Images lists all images.
func (i *Image) Images() ([]Resource, error) {
	return readList(i.ctx, images.List(i.sc, images.ListOpts{}), "images", "")
}

// GetImage fetches an image record by ID.
func (i *Image) GetImage(id string) (Resource, error) {
	return readObject(images.Get(i.ctx, i.sc, id), "")
}

// CreateImage creates an image record (name, disk_format, container_format, ...).
// Upload the binary payload separately with UploadImage.
func (i *Image) CreateImage(opts Resource) (Resource, error) {
	o, err := buildOpts[images.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(images.Create(i.ctx, i.sc, o), "")
}

// UpdateImage updates an image record's mutable attributes. Glance uses a JSON
// patch document, so callers pass a typed images.UpdateOpts (a list of
// Add/Replace/Remove operations).
func (i *Image) UpdateImage(id string, opts images.UpdateOpts) (Resource, error) {
	return readObject(images.Update(i.ctx, i.sc, id, opts), "")
}

// DeleteImage deletes an image by ID.
func (i *Image) DeleteImage(id string) error {
	return done(images.Delete(i.ctx, i.sc, id))
}

// UploadImage streams the binary image payload for a previously-created image.
func (i *Image) UploadImage(id string, data io.Reader) error {
	return done(imagedata.Upload(i.ctx, i.sc, id, data))
}
