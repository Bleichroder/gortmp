package rtmp

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"path"
	"runtime"
	"strings"
)

//connect命令
func handleConnect(msg *rtmpMessage, f float64, obj *amfObj) {
	fmt.Println("")
	fmt.Println("handle Connection")

	msg.app = obj.objs["app"].str
	_, fullFilename, _, _ := runtime.Caller(0)
	filenamewithSuffix := path.Base(fullFilename)
	fileSuffix := path.Ext(filenamewithSuffix)
	filename := strings.TrimSuffix(filenamewithSuffix, fileSuffix)
	fmt.Println("msg.app: ", msg.app)
	fmt.Println("filename: ", filename)
	if msg.app != filename && msg.app != "live" {
		fmt.Println("App Name is Wrong")

		//_result(Connect.Rejected)
		rt := new(rtmpPacket)
		rt.chunkfmt = 0
		rt.csid = 3
		rt.streamid = 0
		rt.timestamp = 0
		rt.typeid = COMMAND_AMF0
		rt.data = new(bytes.Buffer)
		amfObjs := []amfObj{
			amfObj{objType: AMF_STRING, str: "error"},
			amfObj{objType: AMF_NUMBER, f64: f},
			amfObj{objType: AMF_OBJECT,
				objs: map[string]*amfObj{
					"level":       &amfObj{objType: AMF_STRING, str: "error"},
					"code":        &amfObj{objType: AMF_STRING, str: "NetConnection.Connect.Rejected"},
					"description": &amfObj{objType: AMF_STRING, str: "[ Server.Reject ] : (_defaultRoot_, _defaultVHost_) : Application (" + msg.app + ") is not defined."},
				},
			},
		}
		for _, o := range amfObjs {
			binary.Write(rt.data, binary.BigEndian, amf_write(&o))
		}
		rt.msglen = uint32(len(rt.data.Bytes()))
		msg.writePacket(rt)
		fmt.Println("_result(Connect.Rejected)")
		fmt.Println("")

		//close()
		rt.data.Reset()
		rt.chunkfmt = 0
		rt.csid = 3
		rt.streamid = 0
		rt.timestamp = 0
		rt.typeid = COMMAND_AMF0
		amfObjs = []amfObj{
			amfObj{objType: AMF_STRING, str: "close"},
			amfObj{objType: AMF_NUMBER, f64: 0},
			amfObj{objType: AMF_NULL},
		}
		for _, o := range amfObjs {
			binary.Write(rt.data, binary.BigEndian, amf_write(&o))
		}
		rt.msglen = uint32(len(rt.data.Bytes()))
		msg.writePacket(rt)
		fmt.Println("close()")
		fmt.Println("")

		return
	}
	fmt.Println("")

	r := new(rtmpPacket)
	//Windows Acknowledgement Size
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = WINDOW_ACKNOWLEDGEMENT_SIZE
	b := change_four_byte(uint32(2500000))
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("Windows Acknowledgement Size")
	fmt.Println("")

	//Set Peer Bandwidth
	r.data.Reset()
	r.chunkfmt = 1
	r.csid = 2
	r.timestamp = 0
	r.typeid = SET_PEER_BANDWIDTH
	b = change_four_byte(uint32(2500000))
	b = append(b, byte(2))
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("Set Peer Bandwidth")
	fmt.Println("")

	//Stream Begin 0
	r.data.Reset()
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = USER_CONTROL_MESSAGE
	b = change_two_byte(uint16(0))
	for _, temp := range change_four_byte(uint32(0)) {
		b = append(b, temp)
	}
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("Stream Begin 0")
	fmt.Println("")

	//_result(Connect.Success)
	r.data.Reset()
	r.chunkfmt = 0
	r.csid = 3
	r.streamid = 0
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	amfObjs := []amfObj{
		amfObj{objType: AMF_STRING, str: "_result"},
		amfObj{objType: AMF_NUMBER, f64: f},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"fmtVer":       &amfObj{objType: AMF_STRING, str: "FMS/3,0,1,123"},
				"capabilities": &amfObj{objType: AMF_NUMBER, f64: 31},
				"mode":         &amfObj{objType: AMF_NUMBER, f64: 1},
			},
		},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"level":          &amfObj{objType: AMF_STRING, str: "status"},
				"code":           &amfObj{objType: AMF_STRING, str: "NetConnection.Connect.Success"},
				"description":    &amfObj{objType: AMF_STRING, str: "Connection Success."},
				"objectEncoding": &amfObj{objType: AMF_NUMBER, f64: 3},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("_result(Connect.Success)")
	fmt.Println("")

	//onBWDone()
	r.data.Reset()
	r.chunkfmt = 1
	r.csid = 3
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "onBWDone"},
		amfObj{objType: AMF_NUMBER, f64: 0},
		amfObj{objType: AMF_NULL},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("onBWDone()")
	fmt.Println("")
}