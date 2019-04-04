package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

type Marshaller interface {
	Marshal(io.Writer) error
}

func Parse(data []byte, into interface{}) error {
	buf := bytes.NewBuffer(data)
	err := ParseBuf(buf, reflect.ValueOf(into), nil, nil)
	if err != nil {
		return err
	}

	if buf.Len() > 0 {
		return fmt.Errorf("unread bytes at the end of the buffer: %d remain", buf.Len())
	}

	return nil
}

var (
	interfaceFactories = map[reflect.Type]func(io.Reader, reflect.Value) error{}
)

func ParseBuf(buf io.Reader, into reflect.Value, parent *reflect.Value, parentField *reflect.StructField) error {
	t := into.Type()
	switch t.Kind() {
	case reflect.Ptr:
		return ParseBuf(buf, into.Elem(), nil, nil)
	case reflect.Struct:
		for i := 0; i < into.NumField(); i++ {
			f, tf := into.Field(i), t.Field(i)
			if !f.IsValid() || !f.CanSet() {
				continue
			}
			if err := ParseBuf(buf, f, &into, &tf); err != nil {
				return err
			}
		}
	case reflect.String:
		str, err := parseString(buf)
		if err != nil {
			return err
		}
		into.SetString(str)
	case reflect.Int, reflect.Int32:
		i32, err := parseInt32(buf)
		if err != nil {
			return err
		}
		into.SetInt(int64(i32))
	case reflect.Uint8:
		i8, err := parseUint8(buf)
		if err != nil {
			return err
		}
		into.SetUint(uint64(i8))
	case reflect.Uint32:
		i32, err := parseUint32(buf)
		if err != nil {
			return err
		}
		into.SetUint(uint64(i32))
	case reflect.Uint64:
		i64, err := parseUint64(buf)
		if err != nil {
			return err
		}
		into.SetUint(i64)
	case reflect.Int64:
		i64, err := parseInt64(buf)
		if err != nil {
			return err
		}
		into.SetInt(i64)
	case reflect.Slice:
		// Get the count field
		tags := parentField.Tag.Get("jdwp")
		countName := findKey(tags, "counter")
		count := int(parent.FieldByName(countName).Int())
		// Make a slice of that length
		slice := reflect.MakeSlice(t, count, count)

		for i := 0; i < count; i++ {
			err := ParseBuf(buf, slice.Index(i), nil, nil)
			if err != nil {
				return err
			}
		}
		into.Set(slice)
	case reflect.Interface:
		logrus.Debug("into is ", into, " and type is ", into.Type())
		if parser, ok := interfaceFactories[into.Type()]; !ok {
			return fmt.Errorf("cannot instantiate type %s", into.Type())
		} else {
			return parser(buf, into)
		}
	default:
		logrus.Warn("warning: field ", into, " has unknown kind ", into.Kind())
	}

	return nil
}

func findKey(tags string, key string) string {
	for _, item := range strings.Split(tags, " ") {
		kv := strings.SplitN(item, ":", 2)
		if kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

func parseString(buf io.Reader) (string, error) {
	l, err := parseInt32(buf)
	if err != nil {
		return "", err
	}
	bs := make([]byte, l)
	_, err = io.ReadFull(buf, bs)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func parseInt64(buf io.Reader) (int64, error) {
	var i int64
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func parseUint64(buf io.Reader) (uint64, error) {
	var i uint64
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func parseInt32(buf io.Reader) (int32, error) {
	var i int32
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func parseUint32(buf io.Reader) (uint32, error) {
	var i uint32
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func parseUint8(buf io.Reader) (uint8, error) {
	var i uint8
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func RegisterFactory(slice interface{}, f func(io.Reader, reflect.Value) error) {
	t := reflect.ValueOf(slice).Type().Elem()
	interfaceFactories[t] = f
}
