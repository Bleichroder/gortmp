package rtmp

import (
	"bytes"
	"fmt"
	"net"
)

var (
	appMap			map[string]chan *rtmpPacket = make(map[string]chan *rtmpPacket)
	publishstartts	map[string]uint32			= make(map[string]uint32)
	videoinfo		map[string]*rtmpPacket 		= make(map[string]*rtmpPacket)
	audioinfo		map[string]*rtmpPacket 		= make(map[string]*rtmpPacket)
	setVideoinfo	map[string]bool				= make(map[string]bool)
	setAudioinfo	map[string]bool				= make(map[string]bool)
	playpublish		map[string]bool				= make(map[string]bool)
	streamid		map[string]uint32			= make(map[string]uint32)
	sid				uint32						= 0
)

func handle(conn net.Conn) {
	fmt.Println("New Connection")

	message := &rtmpMessage{conn: conn, msgs: map[uint32]*rtmpPacket{}}

	defer func() {
		if p := recover(); p == "deletestream" {
			return
		}
	}()
	defer message.close()
	defer func() {
		if app, ok := appMap[message.sn]; ok {
			close(app)
			delete(appMap, message.sn)
		}
	}()

	if handShake(conn) < 0 {
		fmt.Println("rtmp handshake failed")
		return
	}

	ishead := true
	for {
		packet := message.readPacket(ishead)
		if packet == nil {
			ishead = false
			continue
		}
		ishead = true
		fmt.Println("")

		switch packet.typeid {
		case DATA_AMF0, COMMAND_AMF0:
			fmt.Println("")
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
				amf_read(packet.data)
				amf_read(packet.data)
				obj = amf_read(packet.data)
				handlePublish(message, obj)
			case "play":
				amf_read(packet.data)
				amf_read(packet.data)
				streamname := amf_read(packet.data)
				amf_read(packet.data)
				amf_read(packet.data)
				reset := amf_read(packet.data)
				handlePlay(message, packet, streamname, reset)
			case "deleteStream":
				handleDeleteStream(message)
			case "@setDataFrame":
			default:
				fmt.Println("error command from client")
			}
		case DATA_AMF3, COMMAND_AMF3:
			fmt.Println("")
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
				amf_read(data)
				amf_read(data)
				obj = amf_read(data)
				handlePublish(message, obj)
			case "play":
				amf_read(data)
				amf_read(data)
				streamname := amf_read(data)
				amf_read(data)
				amf_read(data)
				reset := amf_read(data)
				handlePlay(message, packet, streamname, reset)
			case "deleteStream":
				handleDeleteStream(message)
			case "@setDataFrame":
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
