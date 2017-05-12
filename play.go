package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"time"
	"strconv"
	"path"
	"runtime"
	"strings"
)

//play命令
func handlePlay(msg *rtmpMessage, packet *rtmpPacket, streamname *amfObj, reset *amfObj) {
	fmt.Println("handle Play")

	sufid := strconv.Itoa(int(sid))
	s := streamname.str+sufid
	streamid[s] = uint32(sid)

	r := new(rtmpPacket)
	//Set Chunk Size 4096
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

	//Stream Begin 1
	//r.data.Reset()
	r.chunkfmt = 0
	r.csid = 2
	r.streamid = 0
	r.timestamp = 0
	r.typeid = USER_CONTROL_MESSAGE
	b = change_two_byte(uint16(0))
	for _, temp := range change_four_byte(streamid[s]) {
		b = append(b, temp)
	}
	r.data = bytes.NewBuffer(b)
	r.msglen = uint32(len(r.data.Bytes()))
	msg.writePacket(r)
	fmt.Println("Stream Begin 1")
	fmt.Println("")

	//onstatus(NetStream.Play.Reset)
	if reset.i == 1 {
		r.data.Reset()
		r.chunkfmt = 0
		r.csid = 4
		r.streamid = streamid[s]
		r.timestamp = 0
		r.typeid = COMMAND_AMF0
		video := streamname.str
		amfObjs := []amfObj{
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
		fmt.Println("onstatus(NetStream.Play.Reset)")
		fmt.Println("")
	}

	//onstatus(NetStream.Play.Start)
	r.data.Reset()
	r.chunkfmt = 0
	r.csid = 4
	r.streamid = streamid[s]
	r.timestamp = 0
	r.typeid = COMMAND_AMF0
	amfObjs := []amfObj{
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
	fmt.Println("onstatus(NetStream.Play.Start)")
	fmt.Println("")

	//onstatus(NetStream.Data.Start)
	/*r.data.Reset()
	r.chunkfmt = 0
	r.csid = packet.csid
	r.streamid = uint32(streamid)
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
	fmt.Println("")
	fmt.Println("")*/

	//点播一个flv视频
	_, fullFilename, _, _ := runtime.Caller(0)
	filenamewithSuffix := path.Base(fullFilename)
	filedir := strings.TrimSuffix(fullFilename, filenamewithSuffix)
	filedir = filedir + "video/" +streamname.str + ".flv"
	fmt.Println("videodir", filedir)
	videodata, err := ioutil.ReadFile(fmt.Sprintf("%s", filedir))
	if err == nil {
		go func() {
			flvbyte := videodata[13:]
			var length, lastts uint32
			for {
				fmt.Println("tag header: ", flvbyte[0])
				fmt.Println("streamid:", streamid[s])
				switch flvbyte[0] {
				case 8: //音频tag
					length = uint32(flvbyte[1])<<16 | uint32(flvbyte[2])<<8 | uint32(flvbyte[3])
					r.chunkfmt = 0
					r.csid = 4
					r.streamid = streamid[s]
					r.timestamp = uint32(flvbyte[4])<<16 | uint32(flvbyte[5])<<8 | uint32(flvbyte[6])
					r.typeid = AUDIO_TYPE
					r.data.Reset()
					binary.Write(r.data, binary.BigEndian, flvbyte[11:(11+length)])
					r.msglen = uint32(len(r.data.Bytes()))
					msg.writePacket(r)
					//fmt.Println("audio data: ", flvbyte[11:(11+length)])
					flvbyte = flvbyte[(11 + length):]
					flvbyte = flvbyte[4:]
				case 9: //视频tag
					length = uint32(flvbyte[1])<<16 | uint32(flvbyte[2])<<8 | uint32(flvbyte[3])
					r.chunkfmt = 0
					r.csid = 4
					r.streamid = streamid[s]
					r.timestamp = uint32(flvbyte[4])<<16 | uint32(flvbyte[5])<<8 | uint32(flvbyte[6])
					r.typeid = VIDEO_TYPE
					r.data.Reset()
					binary.Write(r.data, binary.BigEndian, flvbyte[11:(11+length)])
					r.msglen = uint32(len(r.data.Bytes()))
					time.Sleep(time.Duration(r.timestamp-lastts) * time.Millisecond) //根据时间戳调整帧间延迟
					lastts = r.timestamp
					msg.writePacket(r)
					//fmt.Println("video data: ", flvbyte[11:(11+length)])
					flvbyte = flvbyte[(11 + length):]
					flvbyte = flvbyte[4:]
				case 18: //脚本tag
					length = uint32(flvbyte[1])<<16 | uint32(flvbyte[2])<<8 | uint32(flvbyte[3])
					metadata := flvbyte[11:(11 + length)]
					bf := new(bytes.Buffer)
					bf.Write(metadata)
					amf_read(bf)		//"onMetaData"
					meta := amf_read(bf)
					var metadatacreator string
					var hasKeyframes, hasVideo, hasAudio, hasMetadata byte
					var width, height, framerate, audiosamplerate, duration, videodatarate, videocodecid, filesize float64
					for k, m := range meta.objs {
						if k == "metadatacreator" {
							metadatacreator = m.str
						}
						if k == "hasKeyframes" {
							hasKeyframes = m.i
						}
						if k == "hasVideo" {
							hasVideo = m.i
						}
						if k == "hasAudio" {
							hasAudio = m.i
						}
						if k == "hasMetadata" {
							hasMetadata = m.i
						}
						if k == "width" {
							width = m.f64
						}
						if k == "height" {
							height = m.f64
						}
						if k == "framerate" {
							framerate = m.f64
						}
						if k == "audiosamplerate" {
							audiosamplerate = m.f64
						}
						if k == "duration" {
							duration = m.f64
						}
						if k == "videodatarate" {
							videodatarate = m.f64
						}
						if k == "videocodecid" {
							videocodecid = m.f64
						}
						if k == "filesize" {
							filesize = m.f64
						}
					}
					//|RtmpSampleAccess
					r.data.Reset()
					r.chunkfmt = 0
					r.csid = 4
					r.timestamp = 0
					r.streamid = streamid[s]
					r.typeid = DATA_AMF0
					amfObjs = []amfObj{
						amfObj{objType: AMF_STRING, str: "|RtmpSampleAccess"},
						amfObj{objType: AMF_BOOLEAN, i: 0},
						amfObj{objType: AMF_BOOLEAN, i: 0},
					}
					for _, o := range amfObjs {
						binary.Write(r.data, binary.BigEndian, amf_write(&o))
					}
					r.msglen = uint32(len(r.data.Bytes()))
					msg.writePacket(r)
					fmt.Println("|RtmpSampleAccess")
					fmt.Println("")

					//onMetaData
					r.data.Reset()
					r.chunkfmt = 0
					r.csid = packet.csid
					r.streamid = streamid[s]
					r.timestamp = 0
					r.typeid = COMMAND_AMF0
					r.data.Reset()
					amfObjs = []amfObj{
						amfObj{objType: AMF_STRING, str: "onMetaData"},
						amfObj{objType: AMF_OBJECT,
							objs: map[string]*amfObj{
								"server":        	&amfObj{objType: AMF_STRING, str: "Golang Rtmp Server"},
								"encoder":       	&amfObj{objType: AMF_STRING, str: "Lavf57.66.104"},
								"metadatacreator":  &amfObj{objType: AMF_NUMBER, str: metadatacreator},
								"width":         	&amfObj{objType: AMF_NUMBER, f64: width},
								"height":        	&amfObj{objType: AMF_NUMBER, f64: height},
								"framerate":     	&amfObj{objType: AMF_NUMBER, f64: framerate},
								"duration":      	&amfObj{objType: AMF_NUMBER, f64: duration},
								"videodatarate": 	&amfObj{objType: AMF_NUMBER, f64: videodatarate},
								"videocodecid":  	&amfObj{objType: AMF_NUMBER, f64: videocodecid},
								"filesize":      	&amfObj{objType: AMF_NUMBER, f64: filesize},
								"audiosamplerate":  &amfObj{objType: AMF_NUMBER, f64: audiosamplerate},
								"hasKeyframes":     &amfObj{objType: AMF_NUMBER, i: hasKeyframes},
								"hasMetadata":      &amfObj{objType: AMF_NUMBER, i: hasMetadata},
								"hasVideo":      	&amfObj{objType: AMF_NUMBER, i: hasVideo},
								"hasAudio":      	&amfObj{objType: AMF_NUMBER, i: hasAudio},
							},
						},
					}
					for _, o := range amfObjs {
						binary.Write(r.data, binary.BigEndian, amf_write(&o))
					}
					r.msglen = uint32(len(r.data.Bytes()))
					msg.writePacket(r)
					fmt.Println("onMetaData")
					fmt.Println("")
					flvbyte = flvbyte[(11 + length):]
					flvbyte = flvbyte[4:]
				}
				if len(flvbyte) == 0 {
					break
				}
			}
		}()
	} else {
		//转发视频
		appMap[s] = make(chan* rtmpPacket, 1000)
		go func() {
			var lastts uint32
			if app, ok := appMap[s]; ok {
				playpublish[s] = true
				for {
					if videoinfo[streamname.str] != nil {
						break
					}
				}
				fmt.Println("videoinfodata:", videoinfo[streamname.str].data.Bytes())
				fmt.Println(videoinfo[streamname.str])
				tmpvinfo := new(rtmpPacket)
				*tmpvinfo = *(videoinfo[streamname.str])
				tmpvinfo.data = new(bytes.Buffer)
				var tb1 []byte
				tb1 = append(tb1, videoinfo[streamname.str].data.Bytes()...)
				tmpvinfo.data.Write(tb1)
				tmpvinfo.streamid = streamid[s]
				msg.writePacket(tmpvinfo)

				if audioinfo[streamname.str] != nil {
					tmpainfo := new(rtmpPacket)
					*tmpainfo = *(audioinfo[streamname.str])
					tmpainfo.data = new(bytes.Buffer)
					var tb2 []byte
					tb2 = append(tb2, audioinfo[streamname.str].data.Bytes()...)
					tmpainfo.data.Write(tb2)
					tmpainfo.streamid = streamid[s]	
					msg.writePacket(tmpainfo)
				}
				for {
					vv, open := <-app
					if !open {
						break
					}
					fmt.Println("out len(app):", len(app))
					//fmt.Println("streamname:", streamname.str)
					r.chunkfmt = 0
					r.csid = 4
					r.streamid = streamid[s]
					r.typeid = vv.typeid
					r.timestamp = vv.timestamp - publishstartts[s]
					if r.typeid == VIDEO_TYPE {
						time.Sleep(time.Duration(r.timestamp-lastts) * time.Millisecond) //根据时间戳调整帧间延迟
						lastts = r.timestamp
					}
					r.data.Reset()
					r.data = bytes.NewBuffer(vv.data.Bytes())
					r.msglen = uint32(len(r.data.Bytes()))
					msg.writePacket(r)
				}
			}
		}()
	}
}
