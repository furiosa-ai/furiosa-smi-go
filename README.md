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

## Installation

`furiosa-smi-go` is available on the [Go Packages](https://pkg.go.dev/).

```shell
go get github.com/furiosa-ai/furiosa-smi-go@latest
```

Once installed, you can import the `furiosa-smi-go` module:

```go
import "github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
```

## Usage

To get started with `furiosa-smi-go`, simply import the `furiosa-smi-go` package and utilize its functions to interact with NPU devices.

The package provides various methods to access the NPU device information and status.

```go
package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

func main() {
	devices, err := smi.ListDevices()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("found %d device(s)\n", len(devices))

	for _, device := range devices {
		deviceInfo, err := device.DeviceInfo()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Device Arch: %v\n", deviceInfo.Arch())
		fmt.Printf("Device CoreNum: %d\n", deviceInfo.CoreNum())
		fmt.Printf("Device NumaNode: %d\n", deviceInfo.NumaNode())
		fmt.Printf("Device Name: %s\n", deviceInfo.Name())
		
		// ... You can use other APIs. Please refer to the documentation.
	}
}
```

The expected output is as below.

```text
found 1 device(s)
Device Arch: 8
Device CoreNum: 0
Device NumaNode: 1986356271
Device Name: /rngd/npu0

...
```

You can refer to [the example go programs](example) for more usage.
- [`device_lister`](example/device_lister/device_lister.go)
- [`show_p2p_capability`](example/show_p2p_capability/show_p2p_capability.go)
- [`show_topology`](example/show_topology/show_topology.go)
