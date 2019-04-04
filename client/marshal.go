package client

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sirupsen/logrus"
)

type S interface {
	Octet(uint8) S
	Int(int) S
	ReferenceTypeId(ReferenceTypeId) S
	ThreadId(ThreadId) S
	ClassId(ClassId) S
	MethodId(MethodId) S
	FrameId(Frame) S
	String(string) S

	Marshal() []byte
}

type s bytes.Buffer

func Seq() S {
	return new(s)
}

func (s *s) Octet(octet uint8) S {
	if err := binary.Write((*bytes.Buffer)(s), binary.BigEndian, octet); err != nil {
		logrus.WithError(err).Error("trouble writing out octet")
	}
	return s
}

func (s *s) Int(i int) S {
	i32 := int32(i)
	if err := binary.Write((*bytes.Buffer)(s), binary.BigEndian, i32); err != nil {
		logrus.WithError(err).Error("trouble writing out integer")
	}
	return s
}

func (s *s) ReferenceTypeId(ref ReferenceTypeId) S {
	if err := ref.Write((*bytes.Buffer)(s)); err != nil {
		logrus.WithError(err).Error("trouble writing out ReferenceTypeId")
	}
	return s
}

func (ref ReferenceTypeId) Write(out io.Writer) error {
	return binary.Write(out, binary.BigEndian, ref)
}

func (s *s) ClassId(id ClassId) S {
	if err := id.Write((*bytes.Buffer)(s)); err != nil {
		logrus.WithError(err).Error("trouble writing out ReferenceTypeId")
	}
	return s
}

func (id ClassId) Write(out io.Writer) error {
	return binary.Write(out, binary.BigEndian, id)
}

func (s *s) ThreadId(id ThreadId) S {
	if err := id.Write((*bytes.Buffer)(s)); err != nil {
		logrus.WithError(err).Error("trouble writing out ReferenceTypeId")
	}
	return s
}

func (id ThreadId) Write(out io.Writer) error {
	return binary.Write(out, binary.BigEndian, id)
}

func (s *s) FrameId(id Frame) S {
	if err := id.Write((*bytes.Buffer)(s)); err != nil {
		logrus.WithError(err).Error("trouble writing out ReferenceTypeId")
	}
	return s
}

func (f Frame) Write(out io.Writer) error {
	err := f.thr.Write(out)
	if err == nil {
		err = binary.Write(out, binary.BigEndian, f.FrameId)
	}
	return err
}

func (s *s) MethodId(m MethodId) S {
	if err := binary.Write((*bytes.Buffer)(s), binary.BigEndian, m.ref); err != nil {
		logrus.WithError(err).Error("trouble writing out MethodId.ReferenceTypeId")
	}
	if err := binary.Write((*bytes.Buffer)(s), binary.BigEndian, m.MethodId); err != nil {
		logrus.WithError(err).Error("trouble writing out MethodId.MethodId")
	}
	return s
}

func (s *s) String(str string) S {
	b := []byte(str)
	l32 := uint32(len(b))
	if err := binary.Write((*bytes.Buffer)(s), binary.BigEndian, l32); err != nil {
		logrus.WithError(err).Error("trouble writing out string length")
		return s
	}

	if _, err := (*bytes.Buffer)(s).Write(b); err != nil {
		logrus.WithError(err).Error("trouble writing out string length")
	}

	return s
}

func (s *s) Marshal() []byte {
	return (*bytes.Buffer)(s).Bytes()
}
