# Furiosa System Management Interface Go Binding

## Overview
Furiosa System Management Interface, is a programmatic interface for managing and monitoring FuriosaAI NPUs.

The interface provides the following API modules, each designed to offer distinct functionalities for managing and monitoring NPU devices.
These modules enable developers to access essential hardware information, topology details, system-wide information, and performance metrics.

#### Device Module
Provides NPU device discovery and information.

- **Features:**
    - Device Specifications
    - Liveness
    - Error Status

#### Topology Module
Provides the device topology status within the system.

- **Features:**
    - Device-to-Device Link Type

#### System Module
Provides system-wide information about NPU devices.

- **Features:**
    - Driver Information

#### Performance Module
Provides NPU device performance status and metrics.

- **Features:**
    - Power Consumption
    - Temperature
    - Core Utilization
    - Memory Utilization
    - Performance Counter

## Prerequisites

`furiosa-smi-go` uses `furiosa-smi` C Library.

Please follow [`furiosa-smi`](https://github.com/furiosa-ai/furiosa-smi) installation guide first.

## Example
The following example demonstrates how to use the Furiosa System Management Interface Go Binding to access and monitor NPU device information.

Start by getting the necessary package.

```shell
go get github.com/furiosa-ai/furiosa-smi-go@latest
```

After that, import package to use go binding.

```go
import "github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
```

You can refer to [the example go programs](example) for more usage.
- [`device_lister`](example/device_lister/device_lister.go)
- [`show_p2p_capability`](example/show_p2p_capability/show_p2p_capability.go)
- [`show_topology`](example/show_topology/show_topology.go)
