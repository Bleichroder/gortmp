package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

//AMF Object objType
const (
	AMF_NUMBER  = 0x00
	AMF_BOOLEAN = 0x01
	AMF_STRING  = 0x02
	AMF_OBJECT  = 0x03
	AMF_NULL    = 0x05
	AMF_MAP     = 0x08
	AMF_OBJEND  = 0x09
	AMF_ARRAY   = 0x0a
)

type amfObj struct {
	objType byte
	str     string
	i       byte
	objs    map[string]*amfObj
	f64     float64
}

//读取amf块
func amf_read(r io.Reader) *amfObj {
	obj := &amfObj{}
	objType := get_byte(r)
	fmt.Println("objType", objType)

	switch objType {
	case AMF_NUMBER:
		obj.f64 = amf_read_number64(r)
		fmt.Println("corenumber", obj.f64)
	case AMF_BOOLEAN:
		obj.i = get_byte(r)
		fmt.Println("coreboolean: ", obj.i)
	case AMF_STRING:
		n := get_two_byte(r)
		obj.str = amf_read_string(r, int32(n))
		fmt.Println("corestring: ", obj.str)
	case AMF_MAP:
		fmt.Println("decode core map")
		num := get_four_byte(r)
		fmt.Println("map num: ", num)
		fallthrough
	case AMF_OBJECT:
		obj.objs = amf_read_core_object(r)
	case AMF_NULL:
		fmt.Println("null")
	}

	return obj
}

//写amf块
func amf_write(obj *amfObj) []byte {
	val := new(bytes.Buffer)

	binary.Write(val, binary.BigEndian, obj.objType)

	switch obj.objType {
	case AMF_NUMBER:
		binary.Write(val, binary.BigEndian, obj.f64)
	case AMF_BOOLEAN:
		binary.Write(val, binary.BigEndian, obj.i)
	case AMF_STRING:
		str := amf_write_string(obj.str)
		binary.Write(val, binary.BigEndian, str)
	case AMF_MAP:
		temp := change_four_byte(uint32(0))
		binary.Write(val, binary.BigEndian, temp)
		fallthrough
	case AMF_OBJECT:
		o := amf_write_core_object(obj)
		binary.Write(val, binary.BigEndian, o)
	}

    return val.Bytes()
}

func amf_read_number64(r io.Reader) float64 {
	var val float64
	data, err := readBuffer(r, 8)

	if err != nil {
		fmt.Println(err)
	}

	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &val)

	return val
}

func amf_read_string(r io.Reader, n int32) string {
	data, err := readBuffer(r, n)

	if err != nil {
		fmt.Println(err)
	}

	return string(data)
}

func amf_write_string(str string) []byte {
	strlen := uint16(len(str))
	val := new(bytes.Buffer)
	binary.Write(val, binary.BigEndian, strlen)
	val.Write([]byte(str))
	return val.Bytes()
}

func amf_read_core_object(r io.Reader) map[string]*amfObj {
	fmt.Println("decode obj")

	objMap := make(map[string]*amfObj)

	for {
		n := get_two_byte(r)
		if n == 0 {
			get_byte(r)
			break
		}

		key := amf_read_string(r, int32(n))
		fmt.Println("key: ", key)

		objectType := get_byte(r)
		fmt.Println("core_object: ", objectType)

		obj := &amfObj{}
		switch objectType {
		case AMF_NUMBER:
			obj.f64 = amf_read_number64(r)
			fmt.Println("corenumber: ", obj.f64)
		case AMF_BOOLEAN:
			obj.i = get_byte(r)
			fmt.Println("coreboolean: ", obj.i)
		case AMF_STRING:
			n = get_two_byte(r)
			obj.str = amf_read_string(r, int32(n))
			fmt.Println("corestring: ", obj.str)
		case AMF_MAP:
			fmt.Println("decode core map")
			num := get_four_byte(r)
			fmt.Println("map num: ", num)
			fallthrough
		case AMF_OBJECT:
			obj.objs = amf_read_core_object(r)
		case AMF_NULL:
			fmt.Println("null")
		}

		objMap[key] = obj
	}

	return objMap
}

func amf_write_core_object(obj *amfObj) []byte {
	val := new(bytes.Buffer)

	for key, value := range obj.objs {
		val.Write(amf_write_string(key))

		binary.Write(val, binary.BigEndian, value.objType)
		switch value.objType {
		case AMF_NUMBER:
			binary.Write(val, binary.BigEndian, value.f64)
		case AMF_BOOLEAN:
			binary.Write(val, binary.BigEndian, value.i)
		case AMF_STRING:
			str := amf_write_string(value.str)
			binary.Write(val, binary.BigEndian, str)
		case AMF_MAP:
			temp := change_four_byte(uint32(0))
			binary.Write(val, binary.BigEndian, temp)
			fallthrough
		case AMF_OBJECT:
			o := amf_write_core_object(value)
			binary.Write(val, binary.BigEndian, o)
		}
	}

	end := change_three_byte(uint32(9))
	binary.Write(val, binary.BigEndian, end)

	return val.Bytes()
}
