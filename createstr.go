package rtmp

import (
	"fmt"
	"bytes"
	"encoding/binary"
)

//createStream命令
func handleCreateStream(msg *rtmpMessage, obj *amfObj) {
	fmt.Println("handle CreateStream")

	sid = sid + 1
	r := new(rtmpPacket)
	//_result(CreateStream.Success)
	r.chunkfmt = 0
	r.csid = 3
	r.streamid = 0
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data = new(bytes.Buffer)
	amfObjs := []amfObj{
		amfObj{objType: AMF_STRING, str: "_result"},
		amfObj{objType: AMF_NUMBER, f64: obj.f64},
		amfObj{objType: AMF_NULL},
		amfObj{objType: AMF_NUMBER, f64: float64(sid)},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("")
	fmt.Println("_result(CreateStream.Success)")
}

//deleteStream命令
func handleDeleteStream(msg *rtmpMessage) {
	panic("deletestream")
}