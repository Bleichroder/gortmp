# gortmp
一个简单的RTMP流媒体服务器包，整体包功能包括点播server端本地视频和摄像头实时推流点播两部分
</br>

## run a simple rtmp server
    package main

    import (
        "github.com/gortmp"
    )
    func main() {
        rtmp.Serve()
    }

## 点播
默认视频路径为"gortmp/video/"，使用vlc播放器或adobe Flash Player网页测试版输入点播url:"rtmp://IP/gortmp/视频名称"，即可开始播放视频，目前仅支持视频h264编码的flv文件

## 直播
使用第三方程序ffmpeg或etc文件夹中的LIVE视频采集程序采集本地USB摄像头或网络摄像头，选择视频编码方式为h264，音频编码方式AVC，将摄像头信号推送到"rtmp://IP/live/livestream"，客户端通过vlc或Adobe Flash Player网页测试版打开url即可看到直播画面

## 参考
#### SPEC
* Adobe官方RTMP协议[Adobe's Real Time Messaging Protocol](http://wwwimages.adobe.com/content/dam/Adobe/en/devnet/rtmp/pdf/rtmp_specification_1.0.pdf)
</br>
此协议属于应用层，被设计用来在合适的传输协议（如TCP）上复用和打包多媒体传输流（如音频、视频和互动内容），里面介绍了基础的定义，握手、连接和传输的方式
* AMF0格式协议[Action Message Format -- AMF 0](http://wwwimages.adobe.com/content/dam/Adobe/en/devnet/amf/pdf/amf0-file-format-specification.pdf)
</br>
AMF是Adobe独家开发出来的通讯协议，它采用二进制压缩，序列化、反序列化、传输数据，从而为Flash 播放器与Flash Remoting网关通信提供了一种轻量级的、高效能的通信方式
* FLV文件协议[Adobe Flash Video File Format Specificati
on](http://download.macromedia.com/f4v/video_file_format_spec_v10_1.pdf)
</br>
介绍了flv文件的封装格式，flv是一种流媒体常用的网络视频格式，被Adobe Flash Player原生支持

#### 博客
* [Adobe 官方公布的 RTMP 规范+未公布的部分](http://blog.csdn.net/simongyley/article/details/24977705)
</br>
[从流程上对rtmp协议经行总结](http://blog.csdn.net/simongyley/article/details/29851337)
</br>
介绍了RTMP的复杂握手模式，并通过抓包的方法对官方协议中没有介绍的消息命令进行解释
* [使用RTMP协议进行视频直播](http://lovejesse.wang/live-rtmp/)
</br>
用实例分析了对一个flv文件解析并传输的过程