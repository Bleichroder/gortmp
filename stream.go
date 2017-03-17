package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// RTMP Chunk Header
//
// The header is broken down into three parts:
//
// | Basic header|Chunk Msg Header|Extended Time Stamp|   Chunk Data |
//
// Chunk basic header: 1 to 3 bytes
//
// This field encodes the chunk stream ID and the chunk type. Chunk
// type determines the format of the encoded message header. The
// length depends entirely on the chunk stream ID, which is a
// variable-length field.
//
// Chunk message header: 0, 3, 7, or 11 bytes
//
// This field encodes information about the message being sent
// (whether in whole or in part). The length can be determined using
// the chunk type specified in the chunk header.
//
// Extended timestamp: 0 or 4 bytes
//
// This field MUST be sent when the normal timsestamp is set to
// 0xffffff, it MUST NOT be sent if the normal timestamp is set to
// anything else. So for values less than 0xffffff the normal
// timestamp field SHOULD be used in which case the extended timestamp
// MUST NOT be present. For values greater than or equal to 0xffffff
// the normal timestamp field MUST NOT be used and MUST be set to
// 0xffffff and the extended timestamp MUST be sent.

type chunkHeader struct {
	//Basic header
	chunkfmt uint8
	csid     uint32

	//Chunk message header
	timestamp uint32
	msglen    uint32
	typeid    uint8
	streamid  uint32

	//Extended timestamp
	extendedts uint32
}

//rtmp包, 包含rtmp header和rtmp body数据
type rtmpPacket struct {
	chunkHeader
	currts uint32
	data   *bytes.Buffer
}

//rtmp消息, 由rtmp包组合而成
type rtmpMessage struct {
	app  string
	ts   uint32
	msgs map[uint32]*rtmpPacket
	conn net.Conn
}

//解析chunk header
func (r *rtmpMessage) readchunkHeader() *chunkHeader {
	h := &chunkHeader{}
	defer func() {
		if p := recover(); p != nil {

		}
	}()

	fmt.Println("=====readChunkHeader=====")
	buf := get_byte(r.conn)
	fmt.Println("rtmp basic header: ", buf)

	h.chunkfmt = (buf & 0xC0) >> 6
	fmt.Println("header format: ", h.chunkfmt)

	h.csid = uint32(buf & 0x3F)
	switch h.csid {
	case 0:
		h.csid = uint32(get_byte(r.conn)) + 64
	case 1:
		h.csid = uint32(get_byte(r.conn)) + 64
		h.csid += uint32(get_byte(r.conn)) * 256
	default:
	}

	switch h.chunkfmt {
	// Chunks of Type 0 are 11 bytes long. This type MUST be used at the
	// start of a chunk stream, and whenever the stream timestamp goes
	// backward (e.g., because of a backward seek).
	//
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                   timestamp                   |message length |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |     message length (cont)     |message type id| msg stream id |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |           message stream id (cont)            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       Figure 9 Chunk Message Header – Type 0
	case 0:
		h.timestamp = get_three_byte(r.conn)
		h.msglen = get_three_byte(r.conn)
		h.typeid = get_byte(r.conn)
		h.streamid = get_four_byte_LE(r.conn)
	// Chunks of Type 1 are 7 bytes long. The message stream ID is not
	// included; this chunk takes the same stream ID as the preceding chunk.
	// Streams with variable-sized messages (for example, many video
	// formats) SHOULD use this format for the first chunk of each new
	// message after the first.
	//
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                timestamp delta                |message length |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |     message length (cont)     |message type id|
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       Figure 10 Chunk Message Header – Type 1
	case 1:
		h.timestamp = get_three_byte(r.conn)
		h.msglen = get_three_byte(r.conn)
		h.typeid = get_byte(r.conn)
	// Chunks of Type 2 are 3 bytes long. Neither the stream ID nor the
	// message length is included; this chunk has the same stream ID and
	// message length as the preceding chunk. Streams with constant-sized
	// messages (for example, some audio and data formats) SHOULD use this
	// format for the first chunk of each message after the first.
	//
	//  0                   1                   2
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                timestamp delta                |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//       Figure 11 Chunk Message Header – Type 2
	case 2:
		h.timestamp = get_three_byte(r.conn)
	// Chunks of Type 3 have no header. Stream ID, message length and
	// timestamp delta are not present; chunks of this type take values from
	// the preceding chunk. When a single message is split into chunks, all
	// chunks of a message except the first one, SHOULD use this type. Refer
	// to example 2 in section 6.2.2. Stream consisting of messages of
	// exactly the same size, stream ID and spacing in time SHOULD use this
	// type for all chunks after chunk of Type 2. Refer to example 1 in
	// section 6.2.1. If the delta between the first message and the second
	// message is same as the time stamp of first message, then chunk of
	// type 3 would immediately follow the chunk of type 0 as there is no
	// need for a chunk of type 2 to register the delta. If Type 3 chunk
	// follows a Type 0 chunk, then timestamp delta for this Type 3 chunk is
	// the same as the timestamp of Type 0 chunk.
	default:
	}

	if h.timestamp == 0xffffff {
		h.extendedts = get_four_byte(r.conn)
	}

	fmt.Println("chunk stream id: ", h.csid)
	fmt.Println("timestamp: ", h.timestamp)
	fmt.Println("amfSize: ", h.msglen)
	fmt.Println("amfType: ", h.typeid)
	fmt.Println("streamID:", h.streamid)
	fmt.Println("=====readChunkHeader===end=====")
	return h
}

//解析一个rtmp包, 包含chunk header和body部分, 当一个packet数据部分读取完毕之后返回
//rtmppacket, 若未读取完毕则返回nil
func (r *rtmpMessage) readPacket() *rtmpPacket {
	fmt.Println("===========readChunkPacket============")

	var size uint32
	h := r.readchunkHeader()
	if h == nil {
		panic("hah")
	}
	packet, ok := r.msgs[h.csid]

	if !ok {
		packet = &rtmpPacket{data: &bytes.Buffer{}}
		r.msgs[h.csid] = packet
	}

	packet.chunkfmt = h.chunkfmt
	switch packet.chunkfmt {
	case 0:
		packet.csid = h.csid
		packet.timestamp = h.timestamp
		packet.msglen = h.msglen
		packet.typeid = h.typeid
		packet.streamid = h.streamid
		if h.extendedts != 0 {
			packet.extendedts = h.extendedts
			packet.currts = packet.timestamp + h.extendedts
			break
		}
		packet.currts = packet.timestamp
	case 1:
		packet.csid = h.csid
		packet.timestamp = h.timestamp
		packet.msglen = h.msglen
		packet.typeid = h.typeid
		if h.extendedts != 0 {
			packet.extendedts = h.extendedts
			packet.currts += packet.timestamp + h.extendedts
			break
		}
		packet.currts += packet.timestamp
	case 2:
		packet.csid = h.csid
		packet.timestamp = h.timestamp
		if h.extendedts != 0 {
			packet.extendedts = h.extendedts
			packet.currts += packet.timestamp + h.extendedts
			break
		}
		packet.currts += packet.timestamp
	default:
	}

	left := packet.msglen - uint32(packet.data.Len())
	size = 128
	if size > left {
		size = left
	}

	if size > 0 {
		io.CopyN(packet.data, r.conn, int64(size))
	}

	if size == left {
		m := new(rtmpPacket)
		*m = *packet
		packet.data = &bytes.Buffer{}
		fmt.Println("===========readChunkPacket=====end=======")
		return m
	}

	return nil
}

func (r *rtmpMessage)writePacket(p *rtmpPacket) {
	buf := new(bytes.Buffer)

	if p.typeid == VIDEO_TYPE || p.typeid == AUDIO_TYPE{
		if p.csid < 64 {
			cHeader := (p.chunkfmt << 6) & 0xC0
			cHeader = cHeader | uint8(p.csid)
			binary.Write(buf, binary.BigEndian, cHeader)
		} else if p.csid > 63 && p.csid < 320 {
			cHeader := (uint16(p.chunkfmt) << 14) & 0xC000
			cHeader = cHeader | uint16(p.csid-64)
			binary.Write(buf, binary.BigEndian, cHeader)
		} else {
			tempHeader := (p.chunkfmt << 6) & 0x01
			buf.WriteByte(tempHeader)
			tmp := uint16(p.csid - 64)
			binary.Write(buf, binary.LittleEndian, &tmp)
		}
		fmt.Println("packet header type: ", p.chunkfmt)

        size := p.msglen
		//left := p.data.Len()
        fmt.Println("data.len & msglen", p.data.Len(), "&", p.msglen)

		switch p.chunkfmt {
		case 0:
			binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
			binary.Write(buf, binary.BigEndian, change_three_byte(uint32(size)))
			binary.Write(buf, binary.BigEndian, p.typeid)
			binary.Write(buf, binary.BigEndian, change_four_byte_LE(p.streamid))
		case 1:
			binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
			binary.Write(buf, binary.BigEndian, change_three_byte(uint32(size)))
			binary.Write(buf, binary.BigEndian, p.typeid)
		case 2:
			binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
		}

		io.CopyN(buf, p.data, int64(size))
		/*left -= size
		if left == 0 {
			return
		}

		for left > 0 {
			if size > left {
				size = left
			}

			byteHeader := make([]byte, 1)
			byteHeader[0] = (0x3 << 6) | byte(p.csid)
			buf.Write(byteHeader)
            //binary.Write(buf, binary.BigEndian, []byte{0x17, 0, 0, 0, 0})

			io.CopyN(buf, p.data, int64(size))
			left -= (size)
			if left == 0 {
				break
			}
		}*/

		r.conn.Write(buf.Bytes())
        return
	}

	if p.csid < 64 {
		cHeader := (p.chunkfmt << 6) & 0xC0
		cHeader = cHeader | uint8(p.csid)
		binary.Write(buf, binary.BigEndian, cHeader)
	} else if p.csid > 63 && p.csid < 320 {
		cHeader := (uint16(p.chunkfmt) << 14) & 0xC000
		cHeader = cHeader | uint16(p.csid-64)
		binary.Write(buf, binary.BigEndian, cHeader)
	} else {
		tempHeader := (p.chunkfmt << 6) & 0x01
		buf.WriteByte(tempHeader)
		tmp := uint16(p.csid - 64)
		binary.Write(buf, binary.LittleEndian, &tmp)
	}
	fmt.Println("packet header type: ", p.chunkfmt)

	switch p.chunkfmt {
	case 0:
		binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
		binary.Write(buf, binary.BigEndian, change_three_byte(p.msglen))
		binary.Write(buf, binary.BigEndian, p.typeid)
		binary.Write(buf, binary.BigEndian, change_four_byte_LE(p.streamid))
	case 1:
		binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
		binary.Write(buf, binary.BigEndian, change_three_byte(p.msglen))
		binary.Write(buf, binary.BigEndian, p.typeid)
	case 2:
		binary.Write(buf, binary.BigEndian, change_three_byte(p.timestamp))
	}

	size := 128
	left := p.data.Len()

	for left > 0 {
		if size > left {
			size = left
		}

		io.CopyN(buf, p.data, int64(size))
		left -= size
		if left == 0 {
			break
		}

		byteHeader := make([]byte, 1)
		byteHeader[0] = (0x3 << 6) | byte(p.csid)
		buf.Write(byteHeader)
	}

	if p.typeid == VIDEO_TYPE {
		fmt.Println("hah", buf.Bytes())
	}
	r.conn.Write(buf.Bytes())
}

func (r *rtmpMessage) close() {
	r.conn.Close()
}
