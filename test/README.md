# SMI Parity Tests

This directory contains parity tests that compare:

- legacy package behavior (`pkg/legacy`)
- current Go API behavior (`pkg/smi`)

The goal is TDD for the pure-Go migration while preserving external behavior.

## Run

These tests are integration-style and require cgo + Furiosa SMI C headers/libraries.

```bash
go test -tags=smi_parity ./test -v
```

## Mutating Operations

Mutating API parity checks are disabled by default.

- `EnableDevice` / `DisableDevice`
- `SetGovernorProfile`

To enable them, set:

```bash
FURIOSA_SMI_PARITY_MUTATING=1 go test -tags=smi_parity ./test -v
```
