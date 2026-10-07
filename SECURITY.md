# Security Policy

No public reporting channel exists before publication approval. Keep reports private and include version, OS, reproduction steps, and impact. Do not include real private keys, passwords, or funded keystores; use fixtures labeled `TEST ONLY NEVER FUND`.

Release readiness requires Go 1.25.14 or later patched compatible toolchain, `CGO_ENABLED=0`, passing tests/vet/build, and zero vulnerabilities in called symbols from both source and final-binary `govulncheck` scans. Reachable standard-library, geth, crypto, keystore, randomness, or filesystem findings block release.

No claim of perfect zeroization, atomic crash durability, hardened ACLs, formal audit, or guaranteed search time is made.
