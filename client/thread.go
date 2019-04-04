package client

import (
	"fmt"
	"io"
	"reflect"

	"github.com/sirupsen/logrus"
)

const (
	Thread        = CommandSet(11)
	ThreadName    = Command(1)
	ThreadSuspend = Command(2)
	ThreadResume  = Command(3)
	ThreadFrames  = Command(6)
)

type ThreadId ReferenceTypeId

func (id ThreadId) Frames(c Client, startFrame int, length int) ([]Frame, error) {
	res, err := c.Call(Thread, ThreadFrames,
		Seq().ThreadId(id).Int(startFrame).Int(length).Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	ms := struct {
		Count  int
		Frames []Frame `jdwp:"counter:Count"`
	}{}
	err = Parse(res.Data, &ms)
	if err != nil {
		return nil, err
	}
	fs := make([]Frame, ms.Count, ms.Count)
	for i, f := range ms.Frames {
		fs[i] = f
		fs[i].thr = id
	}
	return fs, nil
}

type FrameId uint64

type Frame struct {
	thr      ThreadId
	FrameId  FrameId
	Location Location
}

const (
	StackFrame          = CommandSet(16)
	StackFrameGetValues = Command(1)
)

func (f Frame) GetValues(c Client, vars ...VariableDef) (map[string]TaggedValue, error) {
	valid := []VariableDef{}
	for _, v := range vars {
		if v.CodeIndex <= f.Location.Index && f.Location.Index < v.CodeIndex+uint64(v.Length) {
			logrus.WithField("vName", v.Name).WithField("tag", v.Tag()).WithField("offset", v.Slot).Debug("requesting variable")
			valid = append(valid, v)
		} else {
			logrus.WithField("vName", v.Name).Warn("skipping variable - not in legal scope")
		}
	}
	s := Seq().FrameId(f).Int(len(valid))
	for _, v := range valid {
		s.Int(v.Slot).Octet(uint8(v.Tag()))
	}
	res, err := c.Call(StackFrame, StackFrameGetValues, s.Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	ms := struct {
		Count  int
		Values []TaggedValue `jdwp:"counter:Count"`
	}{}
	err = Parse(res.Data, &ms)
	if err != nil {
		return nil, err
	}
	result := map[string]TaggedValue{}
	for i, v := range valid {
		result[v.Name] = ms.Values[i]
	}
	return result, err
}

type TaggedValue interface {
	Tag() Tag
	RecoverValue(c Client) (interface{}, error)
}

func init() {
	RegisterFactory([]TaggedValue{}, TaggedValueFactory)
}

func TaggedValueFactory(buf io.Reader, into reflect.Value) error {
	var tag Tag
	if t, err := parseUint8(buf); err != nil {
		return err
	} else {
		tag = Tag(t)
	}
	var val TaggedValue
	switch tag {
	case TagObject:
		v := ObjectId(0)
		val = &v
	case TagString:
		v := StringId(0)
		val = &v
	default:
		return fmt.Errorf("unimplemented factory for tag %v", tag)
	}
	vv := reflect.ValueOf(val) //.Elem()
	if err := ParseBuf(buf, vv, nil, nil); err != nil {
		return err
	}
	into.Set(vv)
	return nil
}
