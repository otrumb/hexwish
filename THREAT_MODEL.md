# Threat Model

## Protected assets

- Generated private scalar and password.
- Integrity and confidentiality of persisted Web3 V3 keystore.
- Exact relation among private key, EIP-55 address, requested pattern, and output file.

## Trust boundaries

- CLI strings and keystore files are untrusted input.
- OS CSPRNG, Go runtime, local terminal, filesystem, and pinned dependencies are trusted within documented limits.
- Host compromise, malicious Go/runtime binaries, terminal capture, memory inspection, and storage rollback are outside scope.

## Controls

- `crypto.GenerateKey` obtains fresh OS randomness and rejection-samples each valid secp256k1 scalar independently.
- `crypto.PubkeyToAddress` performs Ethereum Keccak address derivation; matching uses lowercase 40-hex and display uses EIP-55.
- Workers publish one winner through bounded cancellation. Rejected scalar values receive best-effort clearing and are never logged or persisted.
- Passwords use hidden terminal reads only. Web3 V3 encryption uses geth standard scrypt. Verification bounds input at 1 MiB, validates fixed KDF/cipher parameters before KDF, decrypts, rederives address, and checks requested patterns.
- Persistence writes and syncs a same-directory mode-0600 temporary file, then claims the destination through an atomic no-replace hard link. Existing names, including dangling symlinks, fail the claim.

## Residual risk

Go garbage collection and library/runtime copies prevent guaranteed zeroization. Hard-link publication does not promise crash-durable directory metadata, ACL inheritance quality, or protection from a malicious administrator. Search has no guaranteed completion time. Users must protect passwords, host, terminal, backups, and release provenance.
