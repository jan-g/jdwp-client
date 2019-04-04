package client

import (
	"fmt"
)

const (
	ObjectReference              = CommandSet(9)
	ObjectReferenceReferenceType = Command(1)
	ObjectReferenceClassObject   = Command(11)
)

type ObjectId uint64

func (o ObjectId) Tag() Tag {
	return TagObject
}

type Object struct {
	ObjectId ObjectId
	Class    *Class
}

func (o ObjectId) RecoverValue(c Client) (interface{}, error) {
	t, cid, err := o.ReferenceType(c)
	if err != nil {
		return nil, err
	}
	switch t {
	case TypeTagClass:
		cls, err := cid.RecoverClass(c)
		if err != nil {
			return nil, err
		}
		return &Object{
			ObjectId: o,
			Class:    cls,
		}, nil
	default:
		return nil, fmt.Errorf("unimplemented: RecoverValue on ObjectId doesn't recognise TypeTag %v", t)
	}
}

func (o ObjectId) ReferenceType(c Client) (TypeTag, ClassId, error) {
	res, err := c.Call(ObjectReference, ObjectReferenceReferenceType,
		Seq().ReferenceTypeId(ReferenceTypeId(o)).Marshal())
	if err != nil {
		return 0, 0, err
	}
	if res.ErrCode != 0 {
		return 0, 0, lookupError(res.ErrCode)
	}
	var tv struct {
		RTT TypeTag
		Ref ClassId
	}
	err = Parse(res.Data, &tv)
	return tv.RTT, tv.Ref, err
}

func (o ObjectId) ClassObject(c Client) (ClassId, error) {
	res, err := c.Call(ObjectReference, ObjectReferenceClassObject,
		Seq().ReferenceTypeId(ReferenceTypeId(o)).Marshal())
	if err != nil {
		return 0, err
	}
	if res.ErrCode != 0 {
		return 0, lookupError(res.ErrCode)
	}
	var ref ClassId
	err = Parse(res.Data, &ref)
	return ref, err
}

type Class struct {
	ClassId   ClassId
	Signature string
	Fields    []Field
}

func (id ClassId) RecoverClass(c Client) (*Class, error) {
	sig, err := id.Signature(c)
	if err != nil {
		return nil, err
	}
	fs, err := id.Fields(c)
	if err != nil {
		return nil, err
	}
	return &Class{
		ClassId:   id,
		Signature: sig,
		Fields:    fs,
	}, nil
}

func (id ClassId) Fields(c Client) ([]Field, error) {
	r, err := c.Call(ReferenceType, ReferenceTypeFields, Seq().ReferenceTypeId(ReferenceTypeId(id)).Marshal())
	if err != nil {
		return nil, err
	}
	if r.ErrCode != 0 {
		return nil, lookupError(r.ErrCode)
	}
	var res struct {
		Count  int
		Fields []Field `jdwp:"counter:Count"`
	}
	err = Parse(r.Data, &res)
	return res.Fields, err
}

type FieldId uint64
type Field struct {
	FieldId   FieldId
	Name      string
	Signature string
	ModBits   uint32
}
