package rtmp

import (
	"net"
	"fmt"
	"crypto/hmac"
	"crypto/sha256"
	"bytes"
	"math/rand"
)

var (
	clientKey = []byte{
	'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
	'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
	'0', '0', '1',

	0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
	0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	serverKey = []byte{
    'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
    'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
    'S', 'e', 'r', 'v', 'e', 'r', ' ',
    '0', '0', '1',

    0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
    0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
    0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}

	serverVersion = []byte{0x0D, 0x0E, 0x0A, 0x0D}

	schema int
)

func handShake(conn net.Conn) int {
	//receive c0
	c0 := make([]byte, 1)
	size, err := conn.Read(c0)
	if err != nil {
		fmt.Println("Read c0 err : ", err)
		return -1
	}
	fmt.Println("reveive c0")

	//reveive c1
	c1 := make([]byte, 1536)
	size, err = conn.Read(c1)
	if err != nil {
		fmt.Println("Read c1 err : ", err)
		return -1
	}
	fmt.Println("reveive c1")

	if size <= 0 {
		return -1
	}

	//send s0
	s0 := c0
	conn.Write(s0)
	fmt.Println("send s0")

	//get schema
	var off int
	if off = findDigest(clientKey[:30], c1, 8); off == -1 {
		if off = findDigest(clientKey[:30], c1, 772); off == -1 {
			fmt.Println("handshake: You need simple handshake")
			ok := simpHandshake(conn, c1)
			if ok {
				return 0
			} else {
				fmt.Println("HandShake failed")
				return -1
			}
		} else {
			schema = 0
		}
	} else {
		schema = 1
	}
	dig := makeDigest(serverKey, c1[off:off+32], -1)

	//send s1
	s1 := make([]byte, 1536)
	var s1dig []byte
	copy(s1[4:8], serverVersion)
	for i := 8; i < 1536; i++ {
		s1[i] = byte(rand.Int() % 256)
	}
	if (schema == 0) {
		s1dig = writeDigest(serverKey[:36], s1, 772)
	} else {
		s1dig = writeDigest(serverKey[:36], s1, 8)
	}
	conn.Write(s1)
	fmt.Println("send s1")

	//send s2
	s2 := make([]byte, 1536)
	for i:= 0 ; i < 1536; i++ {
		s2[i] = byte(rand.Int() % 256)
	}
	s2dig := makeDigest(dig, s2[0:(1536 - 32)], -1)
	copy(s2[1536-32:], s2dig)
	conn.Write(s2)
	fmt.Println("send s2")

	//receive c2
	c2 := make([]byte, 1536)
	size, err = conn.Read(c2)
	tempd := makeDigest(clientKey, s1dig, -1)
	c2dig := makeDigest(tempd, c2[0:(1536 - 32)], -1)
	if bytes.Compare(c2dig, c2[(1536-32):]) == 0 {
		fmt.Println("reveive c2")
	}
	if err != nil {
		fmt.Println("Read c2 err : ", err)
		return -1
	}

	return 0
}

//生成一个32位的Digest
func makeDigest(key []byte, source []byte, offset int) []byte {
	h := hmac.New(sha256.New, key)
	if offset > 0 && offset <= len(source) - 32 {
		h.Write(source[:offset])
		h.Write(source[offset+32:])
	} else if offset == -1 {
		h.Write(source)
	}
	return h.Sum(nil)
}

//查找一个Digest的位置offset
func findDigest(key []byte, b []byte, base int) int {
	off := 0
	for n := 0; n < 4; n++ {
		off += int(b[n + base])
	}

	off = off % 728 + base + 4
	dig := makeDigest(key, b, off)

	if bytes.Compare(dig, b[off:off+32]) != 0 {
		return -1
	}

	return off
}

//向一个数组写入Digest
func writeDigest(key []byte, b []byte, base int) []byte{
	var off int
	for n := 0; n < 4; n++ {
		off += int(b[base + n])
	}
	off = (off % 728) + base + 4

	dig := makeDigest(key, b, off)

	copy(b[off:off+32], dig)

	return dig
}

//简单握手
func simpHandshake(conn net.Conn, s2 []byte) bool{
	s1 := make([]byte, 1536)
	for i := 0; i < 1536; i++ {
		s1[i] = byte(rand.Int() % 256)
	}
	temp := []byte{0, 0, 0, 0}
	copy(s1[4:8], temp)
	conn.Write(s1)
	fmt.Println("send s1")

	conn.Write(s2)
	fmt.Println("send s2")

	c2 := make([]byte, 1536)
	conn.Read(c2)
	if bytes.Compare(s1, c2) == 0 {
		fmt.Println("reveive c2")
	} else {
		return false
	}

	return true
}