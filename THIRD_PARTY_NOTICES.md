# Third-Party Notices

`hexwish` is MIT-licensed. This choice keeps application code permissive; it does not change third-party terms and is not legal advice.

The binary statically links library packages from `github.com/ethereum/go-ethereum` v1.17.7, licensed LGPL-3.0-or-later. Geth repository command packages are GPL-3.0, but hexwish imports library packages, not geth command code. Release archives must include geth's `COPYING` and `COPYING.LESSER`, identify exact source version and URL, preserve notices, and provide relinking/source materials required by LGPL terms. Review obligations before distribution.

Other linked modules retain their own licenses. Release assembly bundles license texts for every module reported by final binary build metadata and fails if a license is missing. Archives also contain application source, `go.mod`, and `go.sum` as corresponding application/relinking material. Recipients can rebuild or modify linked combinations using documented Go toolchain; distributors remain responsible for legal review of LGPL obligations.
