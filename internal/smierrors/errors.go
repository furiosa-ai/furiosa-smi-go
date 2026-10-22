package smierrors

// furiosa-smi/src/error.rs

import "fmt"

// ErrorKind identifies the category of an SMIError.
type ErrorKind string

const (
	// ErrorKindUnknownArch means a device architecture is unknown.
	ErrorKindUnknownArch ErrorKind = "unknown_arch"
	// ErrorKindParse means a parsing operation fails.
	ErrorKindParse ErrorKind = "parse"
	// ErrorKindIO means a filesystem operation fails.
	ErrorKindIO ErrorKind = "io"
	// ErrorKindDeviceNotAvailable means a device is not available (driver unbound, no mgmt node, etc.).
	ErrorKindDeviceNotAvailable ErrorKind = "device_not_available"
	// ErrorKindDeviceNotFound means a device is not found with the given identifier.
	ErrorKindDeviceNotFound ErrorKind = "device_not_found"
	// ErrorKindUnexpectedValue means a device reports a value inconsistent with the expected state.
	ErrorKindUnexpectedValue ErrorKind = "unexpected_value"
)

// SMIError represents an error from the Furiosa SMI library.
type SMIError interface {
	error
	Kind() ErrorKind
}

// Unknown Architecture Error
type UnknownArchError struct {
	deviceID uint16
}

func UnknownArch(deviceID uint16) SMIError {
	return &UnknownArchError{deviceID: deviceID}
}

func (e *UnknownArchError) Error() string {
	return fmt.Sprintf("unknown architecture: device_id=0x%04x", e.deviceID)
}

func (e *UnknownArchError) Kind() ErrorKind {
	return ErrorKindUnknownArch
}

// Parse Error
type ParseError struct {
	reason string
}

func Parse(reason string) SMIError {
	return &ParseError{reason: reason}
}

func (e *ParseError) Error() string {
	return "parse error: " + e.reason
}

func (e *ParseError) Kind() ErrorKind {
	return ErrorKindParse
}

// I/O Error
type IOError struct {
	err  error
	path string
}

func IO(path string, err error) SMIError {
	return &IOError{err: err, path: path}
}

func (e *IOError) Error() string {
	return fmt.Sprintf("io error: %v (%s)", e.err, e.path)
}

func (e *IOError) Kind() ErrorKind {
	return ErrorKindIO
}

func (e *IOError) Unwrap() error {
	return e.err
}

// Device Not Available Error
type DeviceNotAvailableError struct {
	bdf string
}

func DeviceNotAvailable(bdf string) SMIError {
	return &DeviceNotAvailableError{bdf: bdf}
}

func (e *DeviceNotAvailableError) Error() string {
	return "device not available: " + e.bdf
}

func (e *DeviceNotAvailableError) Kind() ErrorKind {
	return ErrorKindDeviceNotAvailable
}

// DeviceNotFoundError is returned when a device is not found with the given identifier.
type DeviceNotFoundError struct {
	message string
}

func DeviceNotFound(message string) SMIError {
	return &DeviceNotFoundError{message: message}
}

func (e *DeviceNotFoundError) Error() string {
	return "device not found: " + e.message
}

func (e *DeviceNotFoundError) Kind() ErrorKind {
	return ErrorKindDeviceNotFound
}

// Unexpected Value Error
type UnexpectedValueError struct {
	message string
}

func UnexpectedValue(message string) SMIError {
	return &UnexpectedValueError{message: message}
}

func (e *UnexpectedValueError) Error() string {
	return "unexpected value: " + e.message
}

func (e *UnexpectedValueError) Kind() ErrorKind {
	return ErrorKindUnexpectedValue
}
