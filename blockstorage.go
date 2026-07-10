// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumetypes"
)

// BlockStorage is the Cinder (v3) service accessor: volumes, snapshots and
// volume types.
type BlockStorage struct{ *service }

// Volumes lists all volumes (detailed).
func (b *BlockStorage) Volumes() ([]Resource, error) {
	return readList(b.ctx, volumes.List(b.sc, volumes.ListOpts{}), "volumes", "")
}

// GetVolume fetches a volume by ID.
func (b *BlockStorage) GetVolume(id string) (Resource, error) {
	return readObject(volumes.Get(b.ctx, b.sc, id), "volume")
}

// CreateVolume creates a volume (size, name, ...).
func (b *BlockStorage) CreateVolume(opts Resource) (Resource, error) {
	o, err := buildOpts[volumes.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(volumes.Create(b.ctx, b.sc, o, nil), "volume")
}

// UpdateVolume updates a volume (name, description, ...).
func (b *BlockStorage) UpdateVolume(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[volumes.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(volumes.Update(b.ctx, b.sc, id, o), "volume")
}

// DeleteVolume deletes a volume by ID.
func (b *BlockStorage) DeleteVolume(id string) error {
	return done(volumes.Delete(b.ctx, b.sc, id, nil))
}

// Snapshots lists all volume snapshots (detailed).
func (b *BlockStorage) Snapshots() ([]Resource, error) {
	return readList(b.ctx, snapshots.ListDetail(b.sc, snapshots.ListOpts{}), "snapshots", "")
}

// GetSnapshot fetches a snapshot by ID.
func (b *BlockStorage) GetSnapshot(id string) (Resource, error) {
	return readObject(snapshots.Get(b.ctx, b.sc, id), "snapshot")
}

// CreateSnapshot creates a snapshot (volume_id, name, ...).
func (b *BlockStorage) CreateSnapshot(opts Resource) (Resource, error) {
	o, err := buildOpts[snapshots.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(snapshots.Create(b.ctx, b.sc, o), "snapshot")
}

// UpdateSnapshot updates a snapshot (name, description).
func (b *BlockStorage) UpdateSnapshot(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[snapshots.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(snapshots.Update(b.ctx, b.sc, id, o), "snapshot")
}

// DeleteSnapshot deletes a snapshot by ID.
func (b *BlockStorage) DeleteSnapshot(id string) error {
	return done(snapshots.Delete(b.ctx, b.sc, id))
}

// VolumeTypes lists all volume types.
func (b *BlockStorage) VolumeTypes() ([]Resource, error) {
	return readList(b.ctx, volumetypes.List(b.sc, volumetypes.ListOpts{}), "volume_types", "")
}

// GetVolumeType fetches a volume type by ID.
func (b *BlockStorage) GetVolumeType(id string) (Resource, error) {
	return readObject(volumetypes.Get(b.ctx, b.sc, id), "volume_type")
}

// CreateVolumeType creates a volume type (name, description, ...).
func (b *BlockStorage) CreateVolumeType(opts Resource) (Resource, error) {
	o, err := buildOpts[volumetypes.CreateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(volumetypes.Create(b.ctx, b.sc, o), "volume_type")
}

// UpdateVolumeType updates a volume type.
func (b *BlockStorage) UpdateVolumeType(id string, opts Resource) (Resource, error) {
	o, err := buildOpts[volumetypes.UpdateOpts](opts)
	if err != nil {
		return nil, err
	}
	return readObject(volumetypes.Update(b.ctx, b.sc, id, o), "volume_type")
}

// DeleteVolumeType deletes a volume type by ID.
func (b *BlockStorage) DeleteVolumeType(id string) error {
	return done(volumetypes.Delete(b.ctx, b.sc, id))
}
