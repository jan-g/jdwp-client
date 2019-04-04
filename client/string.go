package client

const (
	StringReference      = CommandSet(10)
	StringReferenceValue = Command(1)
)

type StringId uint64

func (o StringId) Tag() Tag {
	return TagString
}

func (o StringId) RecoverValue(c Client) (interface{}, error) {
	res, err := c.Call(StringReference, StringReferenceValue, Seq().ReferenceTypeId(ReferenceTypeId(o)).Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	var s string
	err = Parse(res.Data, &s)
	return s, err
}
