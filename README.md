# Furiosa System Management Interface (SMI) Go Binding

## Overview

Furiosa System Management Interface Go Binding, `furiosa-smi-go`, is a Go Binding to manage and monitor FuriosaAI NPU devices.

Currently, it supports 4 modules for FuriosaAI NPU devices.
- Device module
    - Provides NPU device discovery and information.
        - e.g. name, serial, uuid, driver version, firmware version
    - Diagnose NPU device status.
        - e.g. liveness, error information
- Topology module
    - Provides the device topology status in the system.
        - e.g. device link type
- System module
    - Provides the system information about NPU devices.
        - e.g. driver information
- Performance module
    - Provides NPU device performance status.
        - e.g. power consumption, temperature, NPU utilization

We provide additional language bindings for furiosa-smi.

| Language | Repository                                                       |
|----------|------------------------------------------------------------------|
| Rust     | [`furiosa-smi-rs`](https://github.com/furiosa-ai/furiosa-smi-rs) |
| C        | [`furiosa-smi`](https://github.com/furiosa-ai/furiosa-smi)       |

For additional information, please refer to [Documentation](https://pkg.go.dev/github.com/furiosa-ai/furiosa-smi-go).

## Install

### Requirements

#### Driver & Firmware

See [following guide](https://furiosa-ai.github.io/docs/latest/en/software/installation.html) to install the driver and firmware.

#### `furiosa-smi` C Library

`furiosa-smi-go` uses `furiosa-smi` C Library.

Please follow [`furiosa-smi`](https://github.com/furiosa-ai/furiosa-smi) installation guide first.

### Using go package manager

Get package via below command in your go project.

```shell
go get github.com/furiosa-si/furiosa-smi-go@latest
```

After installation, you can use it in the other go project.

## Example

You can refer to [the example go programs](example) for more usage.
- [`device_lister`](example/device_lister/device_lister.go)
- [`show_p2p_capability`](example/show_p2p_capability/show_p2p_capability.go)
- [`show_topology`](example/show_topology/show_topology.go)

Below are the outputs of above examples.

### `device_lister`
```text
found 1 device(s)
Device Arch: 1
Device CoreNum: 8
Device NumaNode: 0
Device Name: /dev/rngd/npu0
Device Serial: RNGD000143
Device UUID: 09512C86-0702-4303-8F40-474746474A40
Device BDF: 0000:c7:00.0
Device Major: 509
Device Minor: 0
Device FirmwareVersion  Major: 0
  Minor: 0
  Patch: 20
  Meta: c2e6335
Liveness: true
Core Status:
  Core 6: 0
  Core 7: 0
  Core 0: 0
  Core 1: 0
  Core 2: 0
  Core 3: 0
  Core 4: 0
  Core 5: 0
Device Files:
  Cores: [0]
  Path: /dev/rngd/npu0pe0
  Cores: [0 1]
  Path: /dev/rngd/npu0pe0-1
  Cores: [0 1 2 3]
  Path: /dev/rngd/npu0pe0-3
  Cores: [1]
  Path: /dev/rngd/npu0pe1
  Cores: [2]
  Path: /dev/rngd/npu0pe2
  Cores: [2 3]
  Path: /dev/rngd/npu0pe2-3
  Cores: [3]
  Path: /dev/rngd/npu0pe3
  Cores: [4]
  Path: /dev/rngd/npu0pe4
  Cores: [4 5]
  Path: /dev/rngd/npu0pe4-5
  Cores: [4 5 6 7]
  Path: /dev/rngd/npu0pe4-7
  Cores: [5]
  Path: /dev/rngd/npu0pe5
  Cores: [6]
  Path: /dev/rngd/npu0pe6
  Cores: [6 7]
  Path: /dev/rngd/npu0pe6-7
  Cores: [7]
  Path: /dev/rngd/npu0pe7
Device Error Info:
  AxiPostErrorCount: 0
  AxiFetchErrorCount: 0
  AxiDiscardErrorCount: 0
  AxiDoorbellErrorCount: 0
  PciePostErrorCount: 0
  PcieFetchErrorCount: 0
  PcieDiscardErrorCount: 0
  PcieDoorbellErrorCount: 0
  DeviceErrorCount: 0
Core Utilization:
  PE Utilization:
    Core: 0
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 1
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 2
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 3
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 4
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 5
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 6
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
  PE Utilization:
    Core: 7
    Time Window Mill: 0
    PE Usage Percentage: 0.000000
not supported error
Device Performance Counter:
  Core 0 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 1 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 2 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 3 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 4 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 5 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 6 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
  Core 7 Performance Counter:
    Timestamp: 2024-11-08 17:37:38 +0900 KST
    Cycle Count: 0
    Task Execution Cycle: 0
Device Temperature:
  Soc Peak: 59.018000
  Ambient: 47.000000
Power Consumption: 42.000000
```

### `show_p2p_capability`

```text
+------+------------+
| #    | NPU0       |
+------+------------+
| npu0 | Accessible |
+------+------------+
```

### `show_topology`

```text
+------+------+
| #    | NPU0 |
+------+------+
| npu0 | NoC  |
+------+------+
```