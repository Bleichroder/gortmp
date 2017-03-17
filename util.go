package rtmp

import (
    "fmt"
    "io"
    "encoding/binary"
    "bytes"
)

// 从r中读取n个byte
func readBuffer(r io.Reader,n int32) ([]byte,error) {
    b := make([]byte,n)
    _,err := r.Read(b)
    return b,err
}

//  向w中写入[]byte
func writeBuffer(w io.Writer,buf []byte) error {
    _,err := w.Write(buf)
    return err
}

func get_byte(r io.Reader) byte {
    data, err := readBuffer(r, 1)
    if err != nil{
        fmt.Println(err)
        panic(err)
    }

    return data[0]
}

func get_two_byte(r io.Reader) uint16 {
    data, err := readBuffer(r, 2)
    if err != nil{
        fmt.Println(err)
    }

    return uint16(data[0])<<8 | uint16(data[1])
}

func get_three_byte(r io.Reader) uint32 {
    data, err := readBuffer(r, 3)
    if err != nil{
        fmt.Println(err)
    }

    return uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
}

func  get_four_byte(r io.Reader) uint32 {
    data, err := readBuffer(r, 4)
    if err != nil{
        fmt.Println(err)
    }

    var v uint32
    v = binary.BigEndian.Uint32(data)
    return v
}

func get_four_byte_LE(r io.Reader) uint32 {
    data, err := readBuffer(r, 4)
    if err != nil{
        fmt.Println(err)
    }

    var v uint32
    v = binary.LittleEndian.Uint32(data)
    return v
}

func change_two_byte(data uint16) []byte {
    val := new(bytes.Buffer)
    binary.Write(val,binary.BigEndian,data)
    return val.Bytes()
}

func change_three_byte(data uint32) []byte {
    val := make([]byte, 3)
    val[0] = byte((data & 0x00ff0000)>>16)
    val[1] = byte((data & 0x0000ff00)>>8)
    val[2] = byte(data & 0xff)
    return val
}

func change_four_byte(data uint32) []byte {
    val := new(bytes.Buffer)
    binary.Write(val,binary.BigEndian,data)
    return val.Bytes()
}

func change_four_byte_LE(data uint32) []byte {
    val := new(bytes.Buffer)
    binary.Write(val,binary.LittleEndian,data)
    return val.Bytes()
}