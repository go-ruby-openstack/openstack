// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/gophercloud/gophercloud/v2/pagination"
)

// Resource is a Ruby-style hash describing a single OpenStack resource. Keys
// are the resource's JSON attribute names exactly as the OpenStack API returns
// them (id, name, status, admin_state_up, ...) and values are the decoded JSON
// scalars, arrays and nested hashes. This is the natural shape for a Ruby
// binding: it maps directly onto a Ruby Hash.
//
// Conversions read the raw API response rather than re-marshalling gophercloud's
// typed structs, so the keys are always the wire (snake_case) names even for the
// few gophercloud structs that omit output JSON tags.
type Resource = map[string]any

// marshalJSON and unmarshalJSON are indirection seams over encoding/json so
// tests can force the otherwise-unreachable marshalling error branches.
var (
	marshalJSON   = json.Marshal
	unmarshalJSON = json.Unmarshal
)

// buildOpts turns a Ruby-supplied option hash into a typed gophercloud options
// struct by round-tripping through JSON. It lets callers pass the same idiomatic
// hash they would build in Ruby (e.g. {"name" => "web", "flavorRef" => "2"})
// while gophercloud still validates required fields when the request is built.
func buildOpts[T any](h Resource) (T, error) {
	var o T
	b, err := marshalJSON(h)
	if err != nil {
		return o, mapError(err)
	}
	if err := unmarshalJSON(b, &o); err != nil {
		return o, mapError(err)
	}
	return o, nil
}

// readObject extracts a single resource from a gophercloud result, unwrapping
// the JSON envelope key (for example "server"). Pass "" when the object is the
// body itself, as with Glance images.
//
// Every gophercloud single-object result embeds gophercloud.Result, so it
// carries the already-decoded response in a promoted Body field and the request
// error in a promoted Err field. We read those directly (rather than the typed
// Extract, whose ExtractInto is overridden to demand a struct) so the returned
// hash keys are exactly the wire JSON names, even for the few gophercloud
// structs that omit output tags.
func readObject(r any, key string) (Resource, error) {
	v := reflect.ValueOf(r)
	if errField := v.FieldByName("Err"); errField.IsValid() && !errField.IsNil() {
		return nil, mapError(errField.Interface().(error))
	}
	body, _ := v.FieldByName("Body").Interface().(map[string]any)
	return unwrap(body, key)
}

func unwrap(body Resource, key string) (Resource, error) {
	if key == "" {
		return body, nil
	}
	m, ok := body[key].(map[string]any)
	if !ok {
		return nil, &Error{baseError{message: "openstack: response missing key " + strconv.Quote(key)}}
	}
	return m, nil
}

// errExtractor is satisfied by gophercloud results that only report an error
// (DeleteResult, ActionResult, ...).
type errExtractor interface {
	ExtractErr() error
}

// done maps the error of a side-effecting gophercloud result onto the typed
// tree, returning nil on success.
func done(r errExtractor) error {
	return mapError(r.ExtractErr())
}

// readList walks every page of a gophercloud pager and returns the items as
// Resources. key is the envelope holding the array in an object body (for
// example "servers"); itemKey unwraps a further per-item envelope, used by Nova
// keypairs whose list is [{"keypair": {...}}].
func readList(ctx context.Context, pager pagination.Pager, key, itemKey string) ([]Resource, error) {
	pages, err := pager.AllPages(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	return collect(pages.GetBody(), key, itemKey)
}

func collect(body any, key, itemKey string) ([]Resource, error) {
	var raw []any
	switch b := body.(type) {
	case map[string]any:
		// A single, unpaginated page body.
		raw, _ = b[key].([]any)
	case map[string][]any:
		// The body AllPages builds when it concatenates several linked/marker
		// pages into one.
		raw = b[key]
	case []any:
		raw = b
	case []byte:
		if err := unmarshalJSON(b, &raw); err != nil {
			return nil, mapError(err)
		}
	default:
		return nil, &Error{baseError{message: "openstack: unexpected list body"}}
	}
	out := make([]Resource, 0, len(raw))
	for _, it := range raw {
		m, _ := it.(map[string]any)
		if itemKey != "" {
			m, _ = m[itemKey].(map[string]any)
		}
		out = append(out, m)
	}
	return out, nil
}
