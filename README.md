# Furiosa-smi-go

## Overview

This repository is Golang bindings for [Furiosa-smi](https://github.com/furiosa-ai/furiosa-smi). The goal of the `furiosa-smi-go` is to provide wrappers around the C furiosa-smi API.

### Documentation

TBD

## Requirements
### Driver & Firmware
See [following guide](https://furiosa-ai.github.io/docs/latest/en/software/installation.html) to install the driver and firmware.


## Usage example

To manage or monitor your NPU devices, the current implementation mainly offers an API, namely `list_devices`, which returns `([]smi.Device, error)`

```go
import (
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

devices, err := smi.ListDevices()
```

For example, if you want to monitor the temperature of NPU devices,

```go
for _, device := range devices {
    temperature, err := device.DeviceTemperature()
    ...
}
```

You can refer to [the example go programs](example) for more usage.