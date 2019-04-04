package client

import (
	"bytes"
	"encoding/binary"
)

const (
	EventRequest        = CommandSet(15)
	Set                 = Command(1)
	Clear               = Command(2)
	ClearAllBreakPoints = Command(3)
)

type EventRequestSet struct {
	EventKind     EventKind     `jdwp:"Event kind to request. See JDWP.EventKind for a complete list of events that can be requested; some events may require a capability in order to be requested."`
	SuspendPolicy SuspendPolicy `jdwp:"What threads are suspended when this event occurs? Note that the order of events and command replies accurately reflects the order in which threads are suspended and resumed. For example, if a VM-wide resume is processed before an event occurs which suspends the VM, the reply to the resume command will be written to the transport before the suspending event."`
	Modifiers     int32         `jdwp:"Constraints used to control the number of generated events.Modifiers specify additional tests that an event must satisfy before it is placed in the event queue. Events are filtered by applying each modifier to an event in the order they are specified in this collection Only events that satisfy all modifiers are reported. A value of 0 means there are no modifiers in the request."`
	Data          *bytes.Buffer
}

func NewEventRequestSet(kind EventKind, policy SuspendPolicy) *EventRequestSet {
	return &EventRequestSet{
		EventKind:     kind,
		SuspendPolicy: policy,
		Data:          new(bytes.Buffer),
	}
}

func (e *EventRequestSet) WithMod(kind ModKind) *EventRequestSet {
	e.Modifiers++
	if err := binary.Write(e.Data, binary.BigEndian, kind); err != nil {
		panic(err)
	}
	return e
}

func (e *EventRequestSet) WithInt(i32 int32) *EventRequestSet {
	if err := binary.Write(e.Data, binary.BigEndian, i32); err != nil {
		panic(err)
	}
	return e
}

func (e *EventRequestSet) WithReferenceTypeId(ref ReferenceTypeId) *EventRequestSet {
	if err := binary.Write(e.Data, binary.BigEndian, ref); err != nil {
		panic(err)
	}
	return e
}

func (e *EventRequestSet) WithLocation(l Location) *EventRequestSet {
	if err := l.Write(e.Data); err != nil {
		panic(err)
	}
	return e
}

func (e *EventRequestSet) Marshal() []byte {
	out := new(bytes.Buffer)
	err := writeBytes(nil, out, e.EventKind)
	err = writeBytes(err, out, e.SuspendPolicy)
	err = writeBytes(err, out, e.Modifiers)
	if err != nil {
		panic(err)
	}
	_, err = out.Write(e.Data.Bytes())
	if err != nil {
		panic(err)
	}
	return out.Bytes()
}

type SuspendPolicy uint8

const (
	SuspendPolicyNone        = SuspendPolicy(0)
	SuspendPolicyEventThread = SuspendPolicy(1)
	SuspendPolicyAll         = SuspendPolicy(2)
)

type Mod struct {
	ModKind ModKind
	Data    []byte
}

type ModKind uint8

const (
	ModKindCount = ModKind(1)
	// int count. =1 for one-off
	ModKindConditional = ModKind(2)
	// int exprId
	ModKindThreadOnly = ModKind(3)
	// threadId
	ModKindClass = ModKind(4)
	// referenceTypeID	clazz	Required class
	ModKindClassMatch = ModKind(5)
	// string	classPattern	Required class pattern. Matches are limited to exact matches of the given class pattern and matches of patterns that begin or end with '*'; for example, "*.Foo" or "java.*".
	ModKindClassExclude = ModKind(6)
	// string classPattern
	ModKindLocation = ModKind(7)
	// location	loc	Required location
	INVALID_LOCATION = uint16(24) // Invalid location
	ModKindException = ModKind(8)
	// referenceTypeID	exceptionOrNull	Exception to report. Null (0) means report exceptions of all types. A non-null type restricts the reported exception events to exceptions of the given type or any of its subtypes.
	// boolean	caught	Report caught exceptions
	// boolean	uncaught	Report uncaught exceptions. Note that it is not always possible to determine whether an exception is caught or uncaught at the time it is thrown. See the exception event catch location under composite events for more information.
	ModKindField = ModKind(9)
	// referenceTypeID	declaring	Type in which field is declared.
	// fieldID	fieldID	Required field
	ModKindStep = ModKind(10)
	// threadID	thread	Thread in which to step
	// int	size	size of each step. See JDWP.StepSize
	// int	depth	relative call stack limit. See JDWP.StepDepth
	ModKindInstanceOnly = ModKind(11)
	// objectID	instance	Required 'this' object
	ModKindSourceNameMatch = ModKind(12)
	// string	sourceNamePattern	Required source name pattern. Matches are limited to exact matches of the given pattern and matches of patterns that begin or end with '*'; for example, "*.Foo" or "java.*".
)

type EventRequestSetReply struct {
	RequestId int // ID of created request
}

type EventRequestClear struct {
	EventKind EventKind
	RequestId int
}
