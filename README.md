# furiosa-smi-go

Pure-Go (no cgo) library for the Furiosa System Management Interface (SMI). It reads NPU device
information directly from the host's sysfs/devfs, so consumers build with `CGO_ENABLED=0` and need
no native `furiosa-smi` library installed.

It is shared by the `furiosa-cdi`, `furiosa-device-plugin`, `furiosa-dra-driver`,
`furiosa-feature-discovery`, `furiosa-metrics-exporter` and `furiosa-system-manager` modules in this
workspace, which reference it through a `replace` directive in their `go.mod`.

## Usage

```go
import "github.com/furiosa-ai/furiosa-smi-go/pkg/api"
```

`pkg/api` provides the runtime API and `pkg/api/apitest` provides public test fixtures (see below). Everything under `internal/` is an implementation detail.

Entry points:

- `api.ListDevices()` — all available Furiosa NPU devices with fully resolved information.
- `api.ListPCIDevices()`, `api.GetPCIDevice(bdf)`, `api.ListDisabledDevices()` — PCI-level views.
- `api.DriverInfo()` — NPU driver version.
- `api.ParseBDF`, `api.ParseDeviceCoreRange`, `api.NewDeviceCoreRange` — parsing helpers.

Device, topology, PCIe, performance, temperature, throttle and governor types are re-exported from
`pkg/api` as type aliases (`api.Device`, `api.LinkType`, `api.ThrottleReason`, ...).

See `example/example.go` for a program that calls every read-only function in order, and
`example/list_devices`, `example/show_topology` for smaller samples:

```sh
go run ./example
```

## Testing without hardware

`pkg/api/apitest` builds mock sysfs/devfs fixtures that `pkg/api` reads in place of the real host
filesystem:

- `apitest.SetMockRoot(t, dir)` redirects every filesystem root for the lifetime of `t`.
- `apitest.Build(t, spec)` writes the fixture for a `MockSpec` of `RNGDMock` devices.
- `apitest.StaticRNGDDevices` / `StaticRNGDDevice` install a fixed 8-device, 2-socket RNGD server.

The mock root is process-global state, so tests that use it must not call `t.Parallel()`.

## Real-device cross-check

`test/realdevice` compares `pkg/api` against the Rust `furiosa-smi` CLI on attached hardware and
skips cleanly when there is none. It locates the CLI via:

- `FURIOSA_SMI_CLI_BIN` — path to a built `furiosa-smi` binary, or
- `FURIOSA_SMI_RUST_DIR` — Rust Cargo workspace root to build from.

## Development

```sh
make build    # go build ./...
make vet      # go vet ./...
make test     # go test ./...
make lint     # golangci-lint run
make tidy     # go mod tidy
make example  # build every example into example/bin/ (git-ignored)
make          # build fmt lint vet test tidy
```

Use the Go toolchain pinned in the workspace `go.work` (see the repository root `README.md` for the
`PATH` / `GOTOOLCHAIN=local` setup).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
