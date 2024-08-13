// WARNING: This file has automatically been generated
// Code generated by https://git.io/c-for-go. DO NOT EDIT.

package binding

/*
#include "furiosa/furiosa_smi.h"
#include <stdlib.h>
#include "cgo_helpers.h"
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func FuriosaSmiGetDeviceHandles(outHandles *FuriosaSmiDeviceHandles) FuriosaSmiReturnCode {
	coutHandles, coutHandlesAllocMap := (*C.FuriosaSmiDeviceHandles)(unsafe.Pointer(outHandles)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_handles(coutHandles)
	runtime.KeepAlive(coutHandlesAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiCreateObserver(outObserverInstance **FuriosaSmiObserver) FuriosaSmiReturnCode {
	coutObserverInstance, coutObserverInstanceAllocMap := (*C.FuriosaSmiObserverInstance)(unsafe.Pointer(outObserverInstance)), cgoAllocsUnknown
	__ret := C.furiosa_smi_create_observer(coutObserverInstance)
	runtime.KeepAlive(coutObserverInstanceAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiDestroyObserver(pObserverInstance **FuriosaSmiObserver) FuriosaSmiReturnCode {
	cpObserverInstance, cpObserverInstanceAllocMap := (*C.FuriosaSmiObserverInstance)(unsafe.Pointer(pObserverInstance)), cgoAllocsUnknown
	__ret := C.furiosa_smi_destroy_observer(cpObserverInstance)
	runtime.KeepAlive(cpObserverInstanceAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceHandleByUuid(uuid string, outHandle *FuriosaSmiDeviceHandle) FuriosaSmiReturnCode {
	cuuid, cuuidAllocMap := unpackPCharString(uuid)
	coutHandle, coutHandleAllocMap := (*C.FuriosaSmiDeviceHandle)(unsafe.Pointer(outHandle)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_handle_by_uuid(cuuid, coutHandle)
	runtime.KeepAlive(coutHandleAllocMap)
	runtime.KeepAlive(cuuidAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceHandleBySerial(serial string, outHandle *FuriosaSmiDeviceHandle) FuriosaSmiReturnCode {
	cserial, cserialAllocMap := unpackPCharString(serial)
	coutHandle, coutHandleAllocMap := (*C.FuriosaSmiDeviceHandle)(unsafe.Pointer(outHandle)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_handle_by_serial(cserial, coutHandle)
	runtime.KeepAlive(coutHandleAllocMap)
	runtime.KeepAlive(cserialAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceHandleByBdf(bdf string, outHandle *FuriosaSmiDeviceHandle) FuriosaSmiReturnCode {
	cbdf, cbdfAllocMap := unpackPCharString(bdf)
	coutHandle, coutHandleAllocMap := (*C.FuriosaSmiDeviceHandle)(unsafe.Pointer(outHandle)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_handle_by_bdf(cbdf, coutHandle)
	runtime.KeepAlive(coutHandleAllocMap)
	runtime.KeepAlive(cbdfAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceInfo(handle FuriosaSmiDeviceHandle, outDeviceInfo *FuriosaSmiDeviceInfo) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutDeviceInfo, coutDeviceInfoAllocMap := (*C.FuriosaSmiDeviceInfo)(unsafe.Pointer(outDeviceInfo)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_info(chandle, coutDeviceInfo)
	runtime.KeepAlive(coutDeviceInfoAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceFiles(handle FuriosaSmiDeviceHandle, outDeviceFiles *FuriosaSmiDeviceFiles) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutDeviceFiles, coutDeviceFilesAllocMap := (*C.FuriosaSmiDeviceFiles)(unsafe.Pointer(outDeviceFiles)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_files(chandle, coutDeviceFiles)
	runtime.KeepAlive(coutDeviceFilesAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceCoreStatus(handle FuriosaSmiDeviceHandle, outCoreStatus *FuriosaSmiCoreStatuses) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutCoreStatus, coutCoreStatusAllocMap := (*C.FuriosaSmiCoreStatuses)(unsafe.Pointer(outCoreStatus)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_core_status(chandle, coutCoreStatus)
	runtime.KeepAlive(coutCoreStatusAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceLiveness(handle FuriosaSmiDeviceHandle, outLiveness *bool) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutLiveness, coutLivenessAllocMap := (*C._Bool)(unsafe.Pointer(outLiveness)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_liveness(chandle, coutLiveness)
	runtime.KeepAlive(coutLivenessAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceErrorInfo(handle FuriosaSmiDeviceHandle, outErrorInfo *FuriosaSmiDeviceErrorInfo) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutErrorInfo, coutErrorInfoAllocMap := (*C.FuriosaSmiDeviceErrorInfo)(unsafe.Pointer(outErrorInfo)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_error_info(chandle, coutErrorInfo)
	runtime.KeepAlive(coutErrorInfoAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDriverInfo(outDriverInfo *FuriosaSmiDriverInfo) FuriosaSmiReturnCode {
	coutDriverInfo, coutDriverInfoAllocMap := (*C.FuriosaSmiDriverInfo)(unsafe.Pointer(outDriverInfo)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_driver_info(coutDriverInfo)
	runtime.KeepAlive(coutDriverInfoAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceUtilization(observerInstance *FuriosaSmiObserver, handle FuriosaSmiDeviceHandle, outUtilizationInfo *FuriosaSmiDeviceUtilization) FuriosaSmiReturnCode {
	cobserverInstance, cobserverInstanceAllocMap := (*C.FuriosaSmiObserver)(unsafe.Pointer(observerInstance)), cgoAllocsUnknown
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutUtilizationInfo, coutUtilizationInfoAllocMap := (*C.FuriosaSmiDeviceUtilization)(unsafe.Pointer(outUtilizationInfo)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_utilization(cobserverInstance, chandle, coutUtilizationInfo)
	runtime.KeepAlive(coutUtilizationInfoAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	runtime.KeepAlive(cobserverInstanceAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDevicePowerConsumption(handle FuriosaSmiDeviceHandle, outPowerConsumption *FuriosaSmiDevicePowerConsumption) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutPowerConsumption, coutPowerConsumptionAllocMap := (*C.FuriosaSmiDevicePowerConsumption)(unsafe.Pointer(outPowerConsumption)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_power_consumption(chandle, coutPowerConsumption)
	runtime.KeepAlive(coutPowerConsumptionAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceTemperature(handle FuriosaSmiDeviceHandle, outTemperature *FuriosaSmiDeviceTemperature) FuriosaSmiReturnCode {
	chandle, chandleAllocMap := (C.FuriosaSmiDeviceHandle)(handle), cgoAllocsUnknown
	coutTemperature, coutTemperatureAllocMap := (*C.FuriosaSmiDeviceTemperature)(unsafe.Pointer(outTemperature)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_temperature(chandle, coutTemperature)
	runtime.KeepAlive(coutTemperatureAllocMap)
	runtime.KeepAlive(chandleAllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}

func FuriosaSmiGetDeviceToDeviceLinkType(handle1 FuriosaSmiDeviceHandle, handle2 FuriosaSmiDeviceHandle, outLinkType *FuriosaSmiDeviceToDeviceLinkType) FuriosaSmiReturnCode {
	chandle1, chandle1AllocMap := (C.FuriosaSmiDeviceHandle)(handle1), cgoAllocsUnknown
	chandle2, chandle2AllocMap := (C.FuriosaSmiDeviceHandle)(handle2), cgoAllocsUnknown
	coutLinkType, coutLinkTypeAllocMap := (*C.FuriosaSmiDeviceToDeviceLinkType)(unsafe.Pointer(outLinkType)), cgoAllocsUnknown
	__ret := C.furiosa_smi_get_device_to_device_link_type(chandle1, chandle2, coutLinkType)
	runtime.KeepAlive(coutLinkTypeAllocMap)
	runtime.KeepAlive(chandle2AllocMap)
	runtime.KeepAlive(chandle1AllocMap)
	__v := (FuriosaSmiReturnCode)(__ret)
	return __v
}