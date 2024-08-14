// WARNING: This file has automatically been generated
// Code generated by https://git.io/c-for-go. DO NOT EDIT.

package binding

/*
#include "furiosa/furiosa_smi.h"
#include <stdlib.h>
#include "cgo_helpers.h"
*/
import "C"

const (
	// FuriosaSmiMaxPathSize as defined in smi/furiosa_smi.h:9
	FuriosaSmiMaxPathSize = 256
	// FuriosaSmiMaxDeviceFileSize as defined in smi/furiosa_smi.h:11
	FuriosaSmiMaxDeviceFileSize = 64
	// FuriosaSmiMaxCoreStatusSize as defined in smi/furiosa_smi.h:13
	FuriosaSmiMaxCoreStatusSize = 128
	// FuriosaSmiMaxPeSize as defined in smi/furiosa_smi.h:15
	FuriosaSmiMaxPeSize = 64
	// FuriosaSmiMaxDriverInfoSize as defined in smi/furiosa_smi.h:17
	FuriosaSmiMaxDriverInfoSize = 24
	// FuriosaSmiMaxDeviceHandleSize as defined in smi/furiosa_smi.h:19
	FuriosaSmiMaxDeviceHandleSize = 64
	// FuriosaSmiMaxCstrSize as defined in smi/furiosa_smi.h:21
	FuriosaSmiMaxCstrSize = 96
)

// FuriosaSmiArch as declared in smi/furiosa_smi.h:28
type FuriosaSmiArch int32

// FuriosaSmiArch enumeration from smi/furiosa_smi.h:28
const (
	FuriosaSmiArchWarboy  FuriosaSmiArch = iota
	FuriosaSmiArchRngd    FuriosaSmiArch = 1
	FuriosaSmiArchRngdMax FuriosaSmiArch = 2
	FuriosaSmiArchRngdS   FuriosaSmiArch = 3
)

// FuriosaSmiCoreStatus as declared in smi/furiosa_smi.h:33
type FuriosaSmiCoreStatus int32

// FuriosaSmiCoreStatus enumeration from smi/furiosa_smi.h:33
const (
	FuriosaSmiCoreStatusAvailable FuriosaSmiCoreStatus = iota
	FuriosaSmiCoreStatusOccupied  FuriosaSmiCoreStatus = 1
)

// FuriosaSmiDeviceToDeviceLinkType as declared in smi/furiosa_smi.h:41
type FuriosaSmiDeviceToDeviceLinkType int32

// FuriosaSmiDeviceToDeviceLinkType enumeration from smi/furiosa_smi.h:41
const (
	FuriosaSmiDeviceToDeviceLinkTypeUnknown      FuriosaSmiDeviceToDeviceLinkType = iota
	FuriosaSmiDeviceToDeviceLinkTypeInterconnect FuriosaSmiDeviceToDeviceLinkType = 10
	FuriosaSmiDeviceToDeviceLinkTypeCpu          FuriosaSmiDeviceToDeviceLinkType = 20
	FuriosaSmiDeviceToDeviceLinkTypeBridge       FuriosaSmiDeviceToDeviceLinkType = 30
	FuriosaSmiDeviceToDeviceLinkTypeNoc          FuriosaSmiDeviceToDeviceLinkType = 70
)

// FuriosaSmiReturnCode as declared in smi/furiosa_smi.h:60
type FuriosaSmiReturnCode int32

// FuriosaSmiReturnCode enumeration from smi/furiosa_smi.h:60
const (
	FuriosaSmiReturnCodeOk                       FuriosaSmiReturnCode = iota
	FuriosaSmiReturnCodeInvalidArgumentError     FuriosaSmiReturnCode = 1
	FuriosaSmiReturnCodeNullPointerError         FuriosaSmiReturnCode = 2
	FuriosaSmiReturnCodeMaxBufferSizeExceedError FuriosaSmiReturnCode = 3
	FuriosaSmiReturnCodeDeviceNotFoundError      FuriosaSmiReturnCode = 4
	FuriosaSmiReturnCodeDeviceBusyError          FuriosaSmiReturnCode = 5
	FuriosaSmiReturnCodeIoError                  FuriosaSmiReturnCode = 6
	FuriosaSmiReturnCodePermissionDeniedError    FuriosaSmiReturnCode = 7
	FuriosaSmiReturnCodeUnknownArchError         FuriosaSmiReturnCode = 8
	FuriosaSmiReturnCodeIncompatibleDriverError  FuriosaSmiReturnCode = 9
	FuriosaSmiReturnCodeUnexpectedValueError     FuriosaSmiReturnCode = 10
	FuriosaSmiReturnCodeParseError               FuriosaSmiReturnCode = 11
	FuriosaSmiReturnCodeUnknownError             FuriosaSmiReturnCode = 12
	FuriosaSmiReturnCodeInternalError            FuriosaSmiReturnCode = 13
	FuriosaSmiReturnCodeUninitializedError       FuriosaSmiReturnCode = 14
	FuriosaSmiReturnCodeContextError             FuriosaSmiReturnCode = 15
)
