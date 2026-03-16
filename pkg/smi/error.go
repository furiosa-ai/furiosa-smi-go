package smi

import "errors"

var (
	ErrInvalidArgument      = errors.New("invalid argument error")
	ErrNullPointer          = errors.New("null pointer error")
	ErrMaxBufferSizeExceed  = errors.New("max buffer size exceed error")
	ErrDeviceNotFound       = errors.New("device not found error")
	ErrDeviceBusy           = errors.New("device busy error")
	ErrIo                   = errors.New("io error")
	ErrPermissionDenied     = errors.New("permission denied error")
	ErrUnknownArch          = errors.New("unknown arch error")
	ErrIncompatibleDriver   = errors.New("incompatible driver error")
	ErrUnexpectedValue      = errors.New("unexpected value error")
	ErrParse                = errors.New("parse error")
	ErrUnknown              = errors.New("unknown error")
	ErrInternal             = errors.New("internal error")
	ErrUninitialized        = errors.New("uninitialized error")
	ErrContext              = errors.New("context error")
	ErrNotSupported         = errors.New("not supported error")
)
