package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

var (
	appMap map[string]chan *rtmpPacket = make(map[string]chan *rtmpPacket)
	f float64 = 1
)

func handle(conn net.Conn) {
	fmt.Println("New Connection")

	message := &rtmpMessage{conn: conn, msgs: map[uint32]*rtmpPacket{}}

	defer message.close()
	defer func() {
		if p := recover(); p != "hah" {
			if app, ok := appMap[message.app]; ok {
				app <- nil
				delete(appMap, message.app)
			}
		}
	}()

	if handShake(conn) < 0 {
		fmt.Println("rtmp handshake failed")
		return
	}

	for {
		packet := message.readPacket()
		if packet == nil {
			continue
		}
		fmt.Println("data : ", packet.data.Bytes(), "typeid : ", packet.typeid)

		switch packet.typeid {
		case DATA_AMF0, COMMAND_AMF0:
			fmt.Println("handle command amf0")
			obj := amf_read(packet.data)
			switch obj.str {
			case "connect":
				obj2 := amf_read(packet.data)
				obj3 := amf_read(packet.data)
				handleConnect(message, obj2.f64, obj3)
			case "createStream":
				obj2 := amf_read(packet.data)
				handleCreateStream(message, obj2)
			case "publish":
				handlePublish(message, obj)
			case "play":
				handlePlay(message, packet)
			case "deleteStream":
				handleDeleteStream(message)
			default:
				fmt.Println("error command from client")
			}
		case DATA_AMF3, COMMAND_AMF3:
			fmt.Println("handle command amf3")
			data := new(bytes.Buffer)
			fmt.Println("data : ", packet.data.Bytes()[1:], "typeid : ", packet.typeid)
			data.Write(packet.data.Bytes()[1:])
			obj := amf_read(data)
			switch obj.str {
			case "connect":
				obj2 := amf_read(data)
				obj3 := amf_read(data)
				handleConnect(message, obj2.f64, obj3)
			case "createStream":
				obj2 := amf_read(data)
				handleCreateStream(message, obj2)
			case "publish":
				handlePublish(message,obj)
			case "play":
				handlePlay(message, packet)
			case "deleteStream":
				handleDeleteStream(message)
			default:
				fmt.Println("error command from client")
			}
		case VIDEO_TYPE:
			handleVideo(message, packet)
		case AUDIO_TYPE:
			handleAudio(message, packet)
		}
	}
}

//connect命令
func handleConnect(msg *rtmpMessage, f float64, obj *amfObj) {
	fmt.Println("handle Connection")

	msg.app = obj.objs["app"].str
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

	//Set Peer Bandwidth
	r.data.Reset()
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = SET_PEER_BANDWIDTH
	b = change_four_byte(uint32(2500000))
	b = append(b, byte(2))
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

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

	//onBWDone()
	r.data.Reset()
	r.chunkfmt = 0
	r.csid = 3
	r.streamid = 0
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
}

//createStream命令
func handleCreateStream(msg *rtmpMessage, obj *amfObj) {
	fmt.Println("handle CreateStream")

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
		amfObj{objType: AMF_NUMBER, f64: f},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	f = f+1
}

//play命令
func handlePlay(msg *rtmpMessage, packet *rtmpPacket) {
	fmt.Println("handle Play")

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
		amfObj{objType: AMF_NUMBER, f64: 2},
		amfObj{objType: AMF_NULL},
		amfObj{objType: AMF_NUMBER, f64: 1},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//Stream Begin 1
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = USER_CONTROL_MESSAGE
	b := change_two_byte(uint16(0))
	for _, temp := range change_four_byte(uint32(1)) {
		b = append(b, temp)
	}
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//onstatus(NetStream.Play.Reset)
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = packet.streamid
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data.Reset()
	video := "test.flv"
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "onStatus"},
		amfObj{objType: AMF_NUMBER, f64: 0},
		amfObj{objType: AMF_NULL},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"level":       &amfObj{objType: AMF_STRING, str: "status"},
				"code":        &amfObj{objType: AMF_STRING, str: "NetStream.Play.Reset"},
				"description": &amfObj{objType: AMF_STRING, str: "Playing and reseting" + video},
				"details":     &amfObj{objType: AMF_STRING, str: video},
				//"clientid": &amfObj{objType: AMF_STRING, str: "oAAJAAAA"},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//onstatus(NetStream.Play.Start)
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = packet.streamid
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data.Reset()
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "onStatus"},
		amfObj{objType: AMF_NUMBER, f64: 0},
		amfObj{objType: AMF_NULL},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"level":       &amfObj{objType: AMF_STRING, str: "status"},
				"code":        &amfObj{objType: AMF_STRING, str: "NetStream.Play.Start"},
				"description": &amfObj{objType: AMF_STRING, str: "Start live."},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//|RtmpSampleAccess
	r.chunkfmt = 1
	r.data.Reset()
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "|RtmpSampleAccess"},
		amfObj{objType: AMF_BOOLEAN, i: 1},
		amfObj{objType: AMF_BOOLEAN, i: 1},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//onstatus(NetStream.Play.Start)
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = packet.streamid
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data.Reset()
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "onStatus"},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"code": &amfObj{objType: AMF_STRING, str: "NetStream.Data.Start"},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//onMetaData
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = packet.streamid
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data.Reset()
	amfObjs = []amfObj{
		amfObj{objType: AMF_STRING, str: "onMetaData"},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"server": &amfObj{objType: AMF_STRING, str: "Golang Rtmp Server"},
				"width":         &amfObj{objType: AMF_NUMBER, f64: 800},
				"height":        &amfObj{objType: AMF_NUMBER, f64: 600},
				"framerate":	 &amfObj{objType: AMF_NUMBER, f64: 25},
				"encoder":		 &amfObj{objType: AMF_STRING, str: "Lavf57.66.104"},
				"duration":      &amfObj{objType: AMF_NUMBER, f64: 0},
				"videodatarate": &amfObj{objType: AMF_NUMBER, f64: 0},
				"videocodecid":  &amfObj{objType: AMF_NUMBER, f64: 7},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)

	//转发视频
	if app,ok := appMap[msg.app]; ok {
		r.data.Reset()
        nr := 0
        for {
            vv := <- app
            if vv == nil {
                break
            }
			r.typeid = vv.typeid
			fmt.Println("send typeid:", r.typeid)
            if nr == 0 {
				r.chunkfmt = 0
				r.csid = packet.csid
				r.streamid = packet.streamid
				r.data  = bytes.NewBuffer(vv.data.Bytes())
				r.timestamp = vv.timestamp
				r.msglen = uint32(len(r.data.Bytes()))
                msg.writePacket(r)
                nr++
                continue
            }
			r.chunkfmt = 1
			r.csid = packet.csid
			r.data  = bytes.NewBuffer(vv.data.Bytes())
			r.timestamp = vv.timestamp
			r.msglen = uint32(len(r.data.Bytes()))
            msg.writePacket(r)
        }
    }

	//点播一个flv视频
	/*dir := "/home/blue/go/rtmp-blue/server/test.264"
	data, _ := ioutil.ReadFile(fmt.Sprintf("%s", dir))
	data = append([]byte{0x17, 0, 0, 0, 0}, data...)
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = packet.streamid
	r.timestamp = 0
	r.typeid = VIDEO_TYPE
	r.data.Reset()
	binary.Write(r.data, binary.BigEndian, data)
	r.msglen = uint32(len(r.data.Bytes()))
	fmt.Println(r.chunkfmt, "&", r.csid, "&", r.streamid, "&", r.timestamp, "&", r.typeid, "&")
	msg.writePacket(r)*/
}

//publish命令
func handlePublish(msg *rtmpMessage, obj *amfObj) {
	fmt.Println("handle handlePublish")

	if _, ok := appMap[msg.app]; ok {
		fmt.Printf("app %s already exists", msg.app)
		msg.close()
		return
	}
	
	appMap[msg.app] = make(chan *rtmpPacket, 16)

	r := new(rtmpPacket)

	//onstatus(NetStream.Publish.Start)
	r.chunkfmt = 0
	r.csid = 3
	r.streamid = 0
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	r.data = new(bytes.Buffer)
	amfObjs := []amfObj{
		amfObj{objType: AMF_STRING, str: "onStatus"},
		amfObj{objType: AMF_NUMBER, f64: obj.f64},
		amfObj{objType: AMF_NULL},
		amfObj{objType: AMF_OBJECT,
			objs: map[string]*amfObj{
				"level":       &amfObj{objType: AMF_STRING, str: "status"},
				"code":        &amfObj{objType: AMF_STRING, str: "NetStream.Publish.Start"},
				"description": &amfObj{objType: AMF_STRING, str: "Start publising."},
			},
		},
	}
	for _, o := range amfObjs {
		binary.Write(r.data, binary.BigEndian, amf_write(&o))
	}
	r.msglen = uint32(len(r.data.Bytes()))
	
	msg.writePacket(r)
}

//deleteStream命令, 还需要学习
func handleDeleteStream(msg *rtmpMessage) {}

//Video内容
func handleVideo(msg *rtmpMessage, packet *rtmpPacket) {
	fmt.Println("handleVideo")

    if app,ok := appMap[msg.app]; ok {
        app <- packet
    }
}

//Audio内容
func handleAudio(msg *rtmpMessage, packet *rtmpPacket) {
	fmt.Println("handleAudio")

    if app,ok := appMap[msg.app]; ok {
        app <- packet
    }
}

//StreamID=(ChannelID-4)/5+1
func Serve() {
	//runtime.GOMAXPROCS(runtime.NumCPU())

	ln, err := net.Listen("tcp", ":1935")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}
