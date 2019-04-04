package client

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sirupsen/logrus"
)

type Location struct {
	TypeTag  TypeTag
	ClassId  ClassId
	MethodId MethodId
	Index    uint64
}

type ClassId ReferenceTypeId
type TypeTag uint8

const (
	TypeTagClass     = TypeTag(1) // ReferenceType is a class.
	TypeTagInterface = TypeTag(2) // 	ReferenceType is an interface.
	TypeTagArray     = TypeTag(3) // 	ReferenceType is an array.
)

type Tag uint8

const (
	TagArray       = Tag('[')
	TagByte        = Tag('B')
	TagChar        = Tag('C')
	TagObject      = Tag('L')
	TagFload       = Tag('F')
	TagDouble      = Tag('D')
	TagInt         = Tag('I')
	TagLong        = Tag('J')
	TagShort       = Tag('S')
	TagVoid        = Tag('V')
	TagBoolean     = Tag('Z')
	TagString      = Tag('s')
	TagThread      = Tag('t')
	TagThreadGroup = Tag('g')
	TagClassLoader = Tag('l')
	TagClassObject = Tag('c')
)

func NewLocation(id MethodId, index int64) Location {
	return Location{
		TypeTag:  TypeTagClass,
		ClassId:  id.ref,
		MethodId: id,
		Index:    uint64(index),
	}
}

func (s *s) Location(l Location) S {
	if err := l.Write((*bytes.Buffer)(s)); err != nil {
		logrus.WithError(err).Error("trouble writing out Location")
	}
	return s
}

func (l *Location) Write(out io.Writer) error {
	if err := binary.Write(out, binary.BigEndian, l.TypeTag); err != nil {
		return err
	}
	if err := binary.Write(out, binary.BigEndian, l.ClassId); err != nil {
		return err
	}
	if err := binary.Write(out, binary.BigEndian, l.MethodId.MethodId); err != nil {
		return err
	}
	if err := binary.Write(out, binary.BigEndian, l.Index); err != nil {
		return err
	}
	return nil
}
