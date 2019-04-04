package client

import "fmt"

var (
	Errors                = map[uint16]error{}
	ErrInvalidThread      = err(10, "passed thread is null, is not a valid thread or has exited")
	ErrInvalidThreadGroup = err(11, "thread group invalid")
	ErrInvalidPriority    = err(12, "invalid priority")
	ErrThreadNotSuspended = err(13, "the specified thread has not been suspended by an event")
	ErrThreadSuspended    = err(14, "thread already suspended")
	ErrThreadNotAlive     = err(15, "thread has not been started or is now dead")
	ErrInvalidObject      = err(20, "invalid object") // If this reference type has been unloaded and garbage collected.
	ErrInvalidClass       = err(21, "invalid class")
	ErrClassNotPrepared   = err(22, "class has been loaded but not yet prepared")
	ErrInvalidMethodId    = err(23, "invalid method")
	//INVALID_LOCATION	24	Invalid location.
	//INVALID_FIELDID	25	Invalid field.
	//INVALID_FRAMEID	30	Invalid jframeID.
	//NO_MORE_FRAMES	31	There are no more Java or JNI frames on the call stack.
	//OPAQUE_FRAME	32	Information about the frame is not available.
	//NOT_CURRENT_FRAME	33	Operation can only be performed on current frame.
	//TYPE_MISMATCH	34	The variable is not an appropriate type for the function used.
	//INVALID_SLOT	35	Invalid slot.
	//DUPLICATE	40	Item already set.
	//NOT_FOUND	41	Desired element not found.
	//INVALID_MONITOR	50	Invalid monitor.
	//NOT_MONITOR_OWNER	51	This thread doesn't own the monitor.
	//INTERRUPT	52	The call has been interrupted before completion.
	//INVALID_CLASS_FORMAT	60	The virtual machine attempted to read a class file and determined that the file is malformed or otherwise cannot be interpreted as a class file.
	//CIRCULAR_CLASS_DEFINITION	61	A circularity has been detected while initializing a class.
	//FAILS_VERIFICATION	62	The verifier detected that a class file, though well formed, contained some sort of internal inconsistency or security problem.
	//ADD_METHOD_NOT_IMPLEMENTED	63	Adding methods has not been implemented.
	//SCHEMA_CHANGE_NOT_IMPLEMENTED	64	Schema change has not been implemented.
	//INVALID_TYPESTATE	65	The state of the thread has been modified, and is now inconsistent.
	//HIERARCHY_CHANGE_NOT_IMPLEMENTED	66	A direct superclass is different for the new class version, or the set of directly implemented interfaces is different and canUnrestrictedlyRedefineClasses is false.
	//DELETE_METHOD_NOT_IMPLEMENTED	67	The new class version does not declare a method declared in the old class version and canUnrestrictedlyRedefineClasses is false.
	//UNSUPPORTED_VERSION	68	A class file has a version number not supported by this VM.
	//NAMES_DONT_MATCH	69	The class name defined in the new class file is different from the name in the old class object.
	//CLASS_MODIFIERS_CHANGE_NOT_IMPLEMENTED	70	The new class version has different modifiers and and canUnrestrictedlyRedefineClasses is false.
	//METHOD_MODIFIERS_CHANGE_NOT_IMPLEMENTED	71	A method in the new class version has different modifiers than its counterpart in the old class version and and canUnrestrictedlyRedefineClasses is false.
	//NOT_IMPLEMENTED	99	The functionality is not implemented in this virtual machine.
	//NULL_POINTER	100	Invalid pointer.
	ErrAbsentInformation = err(101, "desired information is not available")
	//INVALID_EVENT_TYPE	102	The specified event type id is not recognized.
	//ILLEGAL_ARGUMENT	103	Illegal argument.
	//OUT_OF_MEMORY	110	The function needed to allocate memory and no more memory was available for allocation.
	//ACCESS_DENIED	111	Debugging has not been enabled in this virtual machine. JVMTI cannot be used.
	ErrVmDead   = err(112, "the virtual machine is not running")
	ErrInternal = err(113, "an unexpected internal error has occurred")
	//UNATTACHED_THREAD	115	The thread being used to call this function is not attached to the virtual machine. Calls must be made from attached threads.
	ErrInvalidTag = err(500, "invalid object type id or class tag")
	//ALREADY_INVOKING	502	Previous invoke not complete.
	//INVALID_INDEX	503	Index is invalid.
	//INVALID_LENGTH	504	The length is invalid.
	ErrInvalidString = err(506, "the string is invalid")
	//INVALID_CLASS_LOADER	507	The class loader is invalid.
	//INVALID_ARRAY	508	The array is invalid.
	//TRANSPORT_LOAD	509	Unable to load the transport.
	//TRANSPORT_INIT	510	Unable to initialize the transport.
	ErrNativeMethod = err(511, "native method")
	ErrInvalidCount = err(512, "the count is invalid")
)

type JdwpError struct {
	Code uint16
	Err  string
}

func err(code uint16, msg string) error {
	e := JdwpError{Code: code, Err: msg}
	Errors[code] = e
	return e
}

func lookupError(code uint16) error {
	if e, ok := Errors[code]; ok {
		return e
	}
	return err(code, fmt.Sprintf("unregistered error %d", code))
}

func (e JdwpError) Error() string {
	return e.Err
}
