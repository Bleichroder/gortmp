package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//publish命令
func handlePublish(msg *rtmpMessage, obj *amfObj) {
	fmt.Println("handle handlePublish")

	msg.sn = obj.str
	streamid[msg.sn] = uint32(sid)
	if _, ok := appMap[msg.sn]; ok {
		fmt.Printf("stream %s already exists", msg.app)
		msg.close()
		return
	}

	appMap[msg.sn] = make(chan *rtmpPacket)

	r := new(rtmpPacket)

	//Set Chunk Size 128
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = SET_CHUNK_SIZE
	b := change_four_byte(uint32(128))
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("Set Chunk Size 128")
	fmt.Println("")

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

//Video内容
func handleVideo(msg *rtmpMessage, packet *rtmpPacket) {
	fmt.Println("handleVideo")

	//fmt.Println("packet.chunkfmt:", packet.chunkfmt)
	msg.ts = calcts(packet.chunkfmt, msg.ts, packet.timestamp)
	var tmppacket *rtmpPacket
	//开始点播正在推流的视频
	for playnow, ok1 := range playpublish {
		//fmt.Println("playnow:", playnow, " ;playnow[:len(msg.sn)]", playnow[:len(msg.sn)])
		if msg.sn == playnow[:len(msg.sn)] {
			if ok1 {
				if app, ok2 := appMap[playnow]; ok2 {
					tmp := packet.data.Bytes()
					fmt.Println("msg.ts:", msg.ts)
					if tmp[0] == 0x17 {
						tmppacket = packet
						tmppacket.timestamp = msg.ts
						if publishstartts[playnow] == 0 {
							publishstartts[playnow] = msg.ts
							//fmt.Println("publishstartts[playnow]:", publishstartts[playnow])
						}
						//fmt.Println("msg.timestamp:", packet.timestamp)
						if publishstartts[playnow] != 0 {
							app <- tmppacket
							fmt.Println("in len(app):", len(app))
						}
					} else if publishstartts[playnow] != 0 {
						fmt.Println("msgname:", msg.sn)
						tmppacket = packet
						tmppacket.timestamp = msg.ts
						//fmt.Println("msg.timestamp:", packet.timestamp)
						app <- tmppacket
						fmt.Println("in len(app):", len(app))
					}
				}
			}
		}
	}

	//未点播视频时获取正在上传视频的PPS、SPS
	if !setVideoinfo[msg.sn] {
		setVideoinfo[msg.sn] = true
		if _, ok := appMap[msg.sn]; ok {
			videoinfo[msg.sn] = packet
		}
	}
}

//Audio内容
func handleAudio(msg *rtmpMessage, packet *rtmpPacket) {
	fmt.Println("handleAudio")

	msg.ts = calcts(packet.chunkfmt, msg.ts, packet.timestamp)
	var tmppacket *rtmpPacket
	//开始点播正在推流的音频
	for playnow, ok := range playpublish {
		if msg.sn == playnow[:len(msg.sn)] {
			if ok {
				if app, ok := appMap[playnow]; ok {
					if publishstartts[playnow] != 0 {
						tmppacket = packet
						tmppacket.timestamp = msg.ts
						//fmt.Println("msg.timestamp:", packet.timestamp)
						app <- tmppacket
					}
				}
			}
		}
	}

	//未点播视频时获取正在上传音频的config
	if !setAudioinfo[msg.sn] {
		setAudioinfo[msg.sn] = true
		if _, ok := appMap[msg.sn]; ok {
			audioinfo[msg.sn] = packet
		}
	}
}

func calcts(fmt uint8, msgts uint32, pacts uint32) uint32 {
	var val uint32
	if fmt == 0 {
		val = pacts
	} else {
		val = msgts + pacts 
	}
	return val
}