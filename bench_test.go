// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-openstack/openstack authors

package openstack

import (
	"strings"
	"testing"
)

// The benchmarks measure the adapter overhead over gophercloud on the request
// build / response parse paths, driven against the in-process mock cloud (no
// network, no live OpenStack). They isolate the cost this package adds: option
// hash -> typed opts marshalling, and raw-body -> Resource hash conversion.

func BenchmarkServerGet(b *testing.B) {
	_, conn := newMockCloud(b)
	c, _ := conn.Compute()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Server("srv-1"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServersList(b *testing.B) {
	_, conn := newMockCloud(b)
	c, _ := conn.Compute()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Servers(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServerCreate(b *testing.B) {
	_, conn := newMockCloud(b)
	c, _ := conn.Compute()
	opts := Resource{"name": "web", "flavorRef": "1", "imageRef": "img-1"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.CreateServer(opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNetworkList(b *testing.B) {
	_, conn := newMockCloud(b)
	n, _ := conn.Network()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := n.Networks(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObjectDownload(b *testing.B) {
	_, conn := newMockCloud(b)
	o, _ := conn.ObjectStorage()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := o.GetObject("logs", "a.txt")
		if err != nil {
			b.Fatal(err)
		}
		if !strings.HasPrefix(string(got), "hello") {
			b.Fatalf("bad content: %s", got)
		}
	}
}
