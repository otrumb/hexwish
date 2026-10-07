# hexwish

`hexwish` v0.1 is an offline EVM vanity-address generator. It creates every candidate with the OS CSPRNG through go-ethereum, matches lowercase address hex, and writes only password-encrypted Web3 V3 keystores.

## Build

Use Go 1.25.14 with `GOTOOLCHAIN=local` and `CGO_ENABLED=0`:

```text
go test -shuffle=on -count=1 ./...
go vet ./...
go build -trimpath -o hexwish ./
```

## Use

```text
hexwish estimate --prefix dead --suffix beef --rate 30000
hexwish benchmark --duration 2s
hexwish generate --prefix dead --suffix beef --output key.json --confirm
hexwish verify --file key.json --prefix dead --suffix beef
```

Repeat `--prefix` or `--suffix` for OR within that side. Prefix and suffix sides combine with AND. Patterns contain 1-40 lowercase hexadecimal characters. Passwords are requested through hidden terminal input only after a winner exists. No password, private key, or seed option exists.

Search time is random and unbounded. `estimate` reports exact probability plus geometric p50/p90/p99 trial counts; ETA depends on supplied measured rate. `benchmark` includes OS randomness, secp256k1 public-key derivation, Keccak address derivation, and matching.

No network, RPC, signing, broadcasting, mnemonic, GPU, regex, plugin, telemetry, or raw-key output exists in v0.1.

## Security

Read `THREAT_MODEL.md` and `SECURITY.md`. Verify downloaded release checksums and keep keystores backed up. This software does not guarantee completion time, perfect zeroization, filesystem atomicity across crashes, ACL hardening, or audit coverage.
