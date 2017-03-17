# gortmp
A RTMP server realized publish and play function
</br>

## run a simple rtmp server
    package main

    import (
        "github.com/rtmp"
    )
    func main() {
        rtmp.Serve()
    }
已经实现了play和publish功能，但是点播本地flv视频并没有成功，将flv文件转为rtmp流还在学习中