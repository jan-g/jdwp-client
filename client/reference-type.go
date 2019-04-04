package client

const (
	ReferenceType          = CommandSet(2)
	ReferenceTypeSignature = Command(1)
	ReferenceTypeMethods   = Command(5)
	ReferenceTypeFields    = Command(4)
)

func (ref ClassId) Signature(c Client) (string, error) {
	res, err := c.Call(ReferenceType, ReferenceTypeSignature, Seq().ClassId(ref).Marshal())
	if err != nil {
		return "", err
	}
	if res.ErrCode != 0 {
		return "", lookupError(res.ErrCode)
	}
	var sig string
	err = Parse(res.Data, &sig)
	if err != nil {
		return "", err
	}
	return sig, nil
}

func (ref ClassId) Methods(c Client) ([]MethodDef, error) {
	res, err := c.Call(ReferenceType, ReferenceTypeMethods, Seq().ClassId(ref).Marshal())
	if err != nil {
		return nil, err
	}
	if res.ErrCode != 0 {
		return nil, lookupError(res.ErrCode)
	}
	ms := struct {
		Declared int
		Methods  []MethodDef `jdwp:"counter:Declared"`
	}{}
	err = Parse(res.Data, &ms)
	if err != nil {
		return nil, err
	}
	for i := range ms.Methods {
		ms.Methods[i].MethodId.ref = ref
	}
	return ms.Methods, nil
}

type MethodId struct {
	ref      ClassId
	MethodId uint64
}

type MethodDef struct {
	MethodId  MethodId
	Name      string
	Signature string
	ModBits   uint32
}
