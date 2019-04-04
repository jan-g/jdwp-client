package client

import "github.com/sirupsen/logrus"

const (
	Method              = CommandSet(6)
	MethodLineTable     = Command(1)
	MethodVariableTable = Command(2)
)

func (m MethodId) LineTable(c Client) (*LineTableReply, error) {
	res, err := c.Call(Method, MethodLineTable, Seq().MethodId(m).Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	var lt LineTableReply
	err = Parse(res.Data, &lt)
	if err != nil {
		return nil, err
	}
	return &lt, nil
}

type LineTableReply struct {
	Start       int64 // Lowest valid code index for the method, >=0, or -1 if the method is native
	End         int64 // Highest valid code index for the method, >=0, or -1 if the method is native
	Lines       int
	LineEntries []LineEntry `jdwp:"counter:Lines"`
}

type LineEntry struct {
	LineCodeIndex int64 // Initial code index of the line, start <= lineCodeIndex < end
	LineNumber    int
}

func (m MethodId) VariableTable(c Client) (*VariableTableReply, error) {
	res, err := c.Call(Method, MethodVariableTable, Seq().MethodId(m).Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	var vtr VariableTableReply
	err = Parse(res.Data, &vtr)
	if err != nil {
		return nil, err
	}
	return &vtr, nil
}

type VariableTableReply struct {
	ArgCount  int
	Slots     int
	Variables []VariableDef `jdwp:"counter:Slots"`
}

type VariableDef struct {
	CodeIndex uint64
	Name      string
	Signature string
	Length    uint32
	Slot      int
}

func (v VariableDef) Tag() Tag {
	switch v.Signature[0] {
	case 'L':
		return TagObject
	default:
		logrus.Fatalf("don't know how to produce a tag for %+v", v)
		return TagVoid
	}
}
