# Dependency Gate: STOP

Date: 2026-10-07

## Candidate

- Local toolchain: `go1.24.0 windows/amd64`
- `GOTOOLCHAIN=local`
- Candidate: `github.com/ethereum/go-ethereum v1.16.8` (`go 1.24.0`)
- Spike imports: `accounts/keystore`, `common`, and `crypto`

## Evidence

- `CGO_ENABLED=0 go test -count=1 ./...`: passed.
- `CGO_ENABLED=0 go build -trimpath`: passed.
- `go version -m` confirmed `CGO_ENABLED=0`, `GOOS=windows`, and `GOARCH=amd64`.
- Linked spike included 13 non-standard modules. Go Ethereum library code is LGPL-3.0; its repository `cmd` code is GPL-3.0. License review did not reach a final all-module conclusion because vulnerability gate stopped work first.
- `GOTOOLCHAIN=local go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 -show verbose ./...` exited 3.

## Kill Condition

Reachable finding `GO-2025-3750`: inconsistent handling of `O_CREATE|O_EXCL` on Windows in `os`/`syscall`.

- Found in local `go1.24.0` standard library.
- Fixed in Go `1.24.4`.
- Reachability trace includes `crypto.GenerateKey` to `os.Open` and `syscall.Open`.
- Required design depends on secure no-overwrite publication on Windows.

This finding directly intersects filesystem safety and the threat contract's overwrite-race kill condition. Automatic toolchain upgrade is forbidden. Product implementation, release workflows, and release claims therefore did not proceed.

Govulncheck also reported `GO-2026-4508` against go-ethereum v1.16.8, fixed in v1.17.0. Its advisory concerns malicious p2p messages, outside this offline CLI, but govulncheck classified symbols in the spike as affected. No exception was assumed because the gate requires stopping on reachable keygen/address/keystore findings.

## Resume Requirement

Coordinator must explicitly select and provide a permitted patched Go 1.24 toolchain (at least 1.24.4, preferably latest patched 1.24.x), then rerun the full dependency, reachability, no-CGO, and license gates from zero. No Rust fallback was attempted.

## Resume Attempt: STOP

Date: 2026-10-07

### Isolated toolchain

- Downloaded official `go1.24.13.windows-amd64.zip` under `C:\Users\trung\AppData\Local\Temp\opencode`.
- Verified SHA-256: `40b16bc8f00540a2cb02dff4de72b73e966fdd8d65f95e33d8e4080b48a2459a`.
- Extracted under `C:\Users\trung\AppData\Local\Temp\opencode\hexwish-go1.24.13`.
- Invoked that portable `go.exe` explicitly with `GOTOOLCHAIN=local`; system Go installation and configuration were not changed.
- Toolchain reported `go version go1.24.13 windows/amd64`.

### Candidate

- Upgraded candidate to `github.com/ethereum/go-ethereum v1.16.9`.
- The downloaded module declares `go 1.24.0`.
- A representative spike exercised `crypto.GenerateKey`, standard-scrypt `keystore.EncryptKey`, and `keystore.DecryptKey`.
- `CGO_ENABLED=0` test passed.
- Representative Windows executable built with `CGO_ENABLED=0`; build metadata reported Go 1.24.13, Windows amd64, and geth v1.16.9.
- `go mod verify` reported `all modules verified`.

### Vulnerability evidence

- Pinned `golang.org/x/vuln/cmd/govulncheck@v1.1.4` no longer reported `GO-2025-3750` with Go 1.24.13.
- Source scan reported `GO-2026-4508` as symbol-reachable and exited 3.
- Binary scan of the representative offline Windows executable also reported `GO-2026-4508`, found in geth v1.16.9 and fixed in v1.17.0, with 182 vulnerable symbols present; it exited 3.
- Binary scan named examples including `accounts.DerivationPath.MarshalJSON`, `accounts.DerivationPath.String`, `accounts.DerivationPath.UnmarshalJSON`, `accounts.ParseDerivationPath`, and `accounts.URL.MarshalJSON`.
- Govulncheck described the advisory as malicious-p2p-message denial of service. The intended CLI has no P2P or network path, but the required exception was conditional on proving no vulnerable symbols reachable from the final offline binary. That proof failed.
- Additional package/module-only findings were reported, including `GO-2025-4087` in `gnark-crypto v0.18.0` and post-Go-1.24 standard-library advisories. Govulncheck did not report these as called symbols in the spike.

### Kill condition

The dependency gate remains stopped. `GO-2026-4508` is still classified as reachable in both source and binary scans. Per the task's explicit condition, implementation, Git initialization, CI/release work, and publication-readiness claims did not proceed.

License review also remains incomplete because the vulnerability gate stopped first. Geth library packages are identified upstream as LGPL-3.0-or-later, but static-link distribution obligations and all transitive licenses require final review before any distribution.

## Final Resume: PASS

Date: 2026-10-07

### Isolated toolchain

- Downloaded official `go1.25.14.windows-amd64.zip` into `C:\Users\trung\AppData\Local\Temp\opencode`.
- Verified exact SHA-256 `119044a92b3987c341cd6aebb256676dd4780d292f7b4e72a3e9976677841697` before extraction.
- Extracted to `C:\Users\trung\AppData\Local\Temp\opencode\hexwish-go1.25.14`; explicit `go.exe` reported `go version go1.25.14 windows/amd64`.
- Every gate Go command used that executable with `GOTOOLCHAIN=local` and `CGO_ENABLED=0`. System Go and configuration were unchanged.

### Candidate and reachability

- Candidate `github.com/ethereum/go-ethereum v1.17.7` declares `go 1.25.0`; module checksum is `h1:jhoGxw/5aYPYUwEIfzfog0RcsiJuLA6SSqsHdhkx1tA=`.
- Representative code called `crypto.GenerateKey`, `crypto.PubkeyToAddress`, standard-scrypt `keystore.EncryptKey`, and `keystore.DecryptKey`, then checked the decrypted address.
- `go test -count=1 ./...`, `go build -trimpath`, and `go mod verify` passed with CGO disabled. Running the executable completed the encrypt/decrypt round trip.
- Build metadata reported Go 1.25.14, Windows amd64, `CGO_ENABLED=0`, and geth v1.17.7.

### Vulnerability result

- `govulncheck v1.7.0` used the 2026-10-01 vulnerability database. Source and representative Windows binary scans both reported zero vulnerabilities in called symbols and zero package findings.
- Both scans reported three module-only findings in `golang.org/x/crypto v0.55.0`: `GO-2026-6355` and `GO-2026-6354` in SSH, and `GO-2026-5932` in OpenPGP. Scanner explicitly stated the code does not call these vulnerabilities; neither package is imported by the offline spike.
- No reachable standard-library, geth, crypto, keystore, or filesystem finding remained. Dependency gate passed; product implementation and local Git initialization may proceed.
