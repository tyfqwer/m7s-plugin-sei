package m7s_plugin_sei

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

/*
自定义配置结构体
配置文件中可以添加相关配置来设置结构体的值
demo:

	http:
	publish:
	subscribe:
	foo: bar
*/
type SeiConfig struct {
	config.HTTP
	config.Publish
	config.Subscribe
	config.Engine
	Foo string `default:"bar"`
}

var seiConfig SeiConfig

// 安装插件
var SeiPlugin = InstallPlugin(&seiConfig)

// 插件事件回调，来自事件总线
func (conf *SeiConfig) OnEvent(event any) {
	switch event.(type) {
	case FirstConfig:
		// 插件启动事件
		break
	}
}

func (conf *SeiConfig) API_insertSEI(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	streamPath := q.Get("streamPath")
	s := Streams.Get(streamPath)
	if s == nil {
		util.ReturnError(util.APIErrorNoStream, NO_SUCH_STREAM, w, r)
		return
	}
	t := q.Get("type")
	tb, err := strconv.ParseInt(t, 10, 8)
	if err != nil {
		if t == "" {
			tb = 5
		} else {
			util.ReturnError(util.APIErrorQueryParse, "type must a number", w, r)
			return
		}
	}
	sei, err := io.ReadAll(r.Body)
	if err == nil {
		// 构造SEI数据
		l := len(sei)
		var buffer util.Buffer
		buffer.WriteByte(byte(tb))
		for l >= 255 {
			buffer.WriteByte(255)
			l -= 255
		}
		buffer.WriteByte(byte(l))
		buffer.Write(sei)
		buffer.WriteByte(0x80)

		//vt := s.Tracks.Video[0]
		//var au util.BLL
		//au.Push(vt.SpesificTrack.GetNALU_SEI())
		//au.Push(vt.BytesPool.GetShell(buffer))
		//vt.Info("sei", zap.Int("len", len(buffer)))
		//vt.Value.AUList.UnshiftValue(&au)

		// 添加SEI数据到所有视频轨道
		for _, vt := range s.Tracks.Video {
			var au util.BLL
			item := vt.BytesPool.Get(1)

			switch vt.CodecID {
			case codec.CodecID_H264:
				item.Value[0] = byte(codec.NALU_SEI)
			case codec.CodecID_H265:
				item.Value[0] = 0b00000000 | byte(codec.NAL_UNIT_SEI<<1)
			default:
				fmt.Println(vt.CodecID)
			}
			au.Push(item)
			au.Push(vt.BytesPool.GetShell(buffer))
			//vt.Info("sei", zap.Int("len", len(buffer)))
			vt.Value.AUList.UnshiftValue(&au)
		}

		util.ReturnOK(w, r)
		//fmt.Println(sei, tb)
		//if s.Tracks.AddSEI(byte(tb), sei) {
		//	util.ReturnOK(w, r)
		//} else {
		//	util.ReturnError(util.APIErrorNoSEI, "no sei track", w, r)
		//}
	} else {
		util.ReturnError(util.APIErrorNoBody, err.Error(), w, r)
	}
}

// http://localhost:8080/demo/api/test/pub
// func (conf *DemoConfig) API_test_pub(rw http.ResponseWriter, r *http.Request) {
// 	var pub DemoPublisher
// 	err := DemoPlugin.Publish("demo/test", &pub)
// 	if err != nil {
// 		rw.Write([]byte(err.Error()))
// 		return
// 	} else {
// 		vt := track.NewH264(pub.Stream)
// 		// 根据实际情况写入视频帧，需要注意pts和dts需要写入正确的值 即毫秒数*90
// 		vt.WriteAnnexB(0, 0, []byte{0, 0, 0, 1})
// 	}
// 	rw.Write([]byte("test_pub"))
// }

// // http://localhost:8080/demo/api/test/sub
//
//	func (conf *DemoConfig) API_test_sub(rw http.ResponseWriter, r *http.Request) {
//		var sub DemoSubscriber
//		err := DemoPlugin.Subscribe("demo/test", &sub)
//		if err != nil {
//			rw.Write([]byte(err.Error()))
//			return
//		} else {
//			sub.PlayRaw()
//		}
//		rw.Write([]byte("test_sub"))
//	}
//
// 自定义发布者
type SeiPublisher struct {
	Publisher
}

// 发布者事件回调
func (pub *SeiPublisher) OnEvent(event any) {
	switch event.(type) {
	case IPublisher:
		// 发布成功
	default:
		pub.Publisher.OnEvent(event)
	}
}

// 自定义订阅者
type SeiSubscriber struct {
	Subscriber
}

// 订阅者事件回调
func (sub *SeiSubscriber) OnEvent(event any) {
	switch event.(type) {
	case ISubscriber:
		// 订阅成功
	case AudioFrame:
		// 音频帧处理
	case VideoFrame:
		// 视频帧处理
	default:
		sub.Subscriber.OnEvent(event)
	}
}
