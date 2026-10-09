# Rebuilding and relinking

The release source archive contains the preferred-form `hexwish` source and
the exact Go module source ZIPs linked into the distributed binary. It is
provided so recipients can inspect, modify, rebuild, and relink the application
with a modified compatible library.

Use Go 1.25.14, set `GOTOOLCHAIN=local` and `CGO_ENABLED=0`, extract the module
ZIPs into a local module proxy or module cache, then adjust `go.mod` with a
`replace` directive that points to the modified module source. Run:

```text
go mod verify
go test -shuffle=on -count=1 ./...
go vet ./...
go build -trimpath ./
```

The application source is MIT-licensed. Each dependency remains governed by
the license material in the binary archive's `licenses/` directory and by its
included source. `github.com/ethereum/go-ethereum` v1.17.7 library packages are
distributed under LGPL-3.0-or-later; see `COPYING` and `COPYING.LESSER` in its
module source and bundled license directory. This document is operational
rebuild guidance, not legal advice.
