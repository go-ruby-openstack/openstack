# Benchmarks

**Date: 2026-07-10**

These benchmarks measure the *adapter* overhead this package adds on top of
[gophercloud/v2](https://github.com/gophercloud/gophercloud) — the two seams
where `go-ruby-openstack/openstack` does work of its own:

1. **request build**: turning a Ruby-style option hash (`Resource`,
   `map[string]any`) into gophercloud's typed `CreateOpts`/`UpdateOpts` via a
   JSON round-trip (`buildOpts`);
2. **response parse**: turning gophercloud's decoded response body into a
   `Resource` hash keyed by the wire (snake_case) attribute names
   (`readObject` / `readList`).

## Methodology

- Driven against the in-process mock cloud (`net/http/httptest`) used by the
  test suite — **no network, no live OpenStack**. Each iteration performs a real
  HTTP round-trip over loopback plus the adapter's marshalling/parsing, so the
  wall-clock is dominated by the HTTP stack; the `allocs/op` column is the
  useful signal for adapter cost.
- Reference points: gophercloud itself (which we import — its per-call cost is
  included in every number below) and the Ruby `fog-openstack` gem, whose
  collection/model API this package mirrors. There is no separate "gophercloud
  vs adapter" delta to report because the adapter *is* a thin call-through: the
  only measurable additions are one `json.Marshal`+`json.Unmarshal` pair per
  mutating call and one body→hash conversion per read.
- `go test -bench . -benchmem`, Go 1.26.4, darwin/arm64 (Apple Silicon, 16
  logical CPUs). Loopback HTTP; absolute ns/op will differ on CI/Linux but the
  allocation profile is stable.

## Results

| Benchmark            | ns/op  | B/op   | allocs/op |
|----------------------|-------:|-------:|----------:|
| `ServerGet`          | 45031  | 10092  | 127       |
| `ServersList`        | 56688  | 17621  | 257       |
| `ServerCreate`       | 56164  | 14523  | 184       |
| `NetworkList`        | 59228  | 16980  | 265       |
| `ObjectDownload`     | 84897  |  7936  |  88       |

## Reading the numbers

- The per-call time is essentially the HTTP round-trip plus gophercloud's own
  request signing and JSON decode. The adapter's contribution is the handful of
  allocations from the option-hash marshal (mutating calls) and the body→hash
  conversion (reads) — negligible next to the transport and gophercloud's own
  reflection-based decoding.
- Because the conversion reads gophercloud's already-decoded body rather than
  re-marshalling its typed structs, a read adds one map walk, not a second full
  JSON encode/decode.
- For a Ruby consumer (rbgo) the dominant cost is, as expected, the network and
  gophercloud — exactly as it would be calling gophercloud directly — so putting
  the idiomatic Ruby surface on top is effectively free.

Reproduce with:

```
go test -run '^$' -bench . -benchmem ./...
```
