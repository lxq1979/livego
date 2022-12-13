package httpflv

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/rtmp"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	handler av.Handler
}

//LXQ:publisher 與 player 均使用此結構體
//其中Key:由 appname/home組成，ex:   live/movie
//ID，代理publisher 或 player 的一個終端
type stream struct {
	Key string `json:"key"`
	Id  string `json:"id"`
}

//LXQ: streams表示流的集合結果，Publishers表示，當前直播服務器中所有 發布者
type streams struct {
	Publishers []stream `json:"publishers"`
	Players    []stream `json:"players"`
}

func NewServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

//提供HTTPFlv服務
func (server *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()

	//處理每一個player的播放請求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server.handleConn(w, r)
	})

	//可用于監聽每個livego服務器的負載，及publisher與 player數量
	mux.HandleFunc("/streams", func(w http.ResponseWriter, r *http.Request) {
		server.getStream(w, r)
	})
	if err := http.Serve(l, mux); err != nil {
		return err
	}
	return nil
}

// 获取所有 publisher 和 每個publisher的player信息
func (server *Server) getStreams(w http.ResponseWriter, r *http.Request) *streams {
	rtmpStream := server.handler.(*rtmp.RtmpStream)
	if rtmpStream == nil {
		return nil
	}
	msgs := new(streams)

	rtmpStream.GetStreams().Range(func(key, val interface{}) bool {
		if s, ok := val.(*rtmp.Stream); ok {
			if s.GetReader() != nil {
				msg := stream{key.(string), s.GetReader().Info().UID}
				msgs.Publishers = append(msgs.Publishers, msg)
			}
		}
		return true
	})

	rtmpStream.GetStreams().Range(func(key, val interface{}) bool {
		ws := val.(*rtmp.Stream).GetWs()

		ws.Range(func(k, v interface{}) bool {
			if pw, ok := v.(*rtmp.PackWriterCloser); ok {
				if pw.GetWriter() != nil {
					msg := stream{key.(string), pw.GetWriter().Info().UID}
					msgs.Players = append(msgs.Players, msg)
				}
			}
			return true
		})
		return true
	})

	return msgs
}

func (server *Server) getStream(w http.ResponseWriter, r *http.Request) {
	msgs := server.getStreams(w, r)
	if msgs == nil {
		return
	}
	resp, _ := json.Marshal(msgs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

//HTTPFlv每一個client 的請求，將來由handleConn來處理
func (server *Server) handleConn(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("http flv handleConn panic: ", r)
		}
	}()

	url := r.URL.String()
	u := r.URL.Path
	//LXQ: 不是.FLV結尾的報錯，并返回
	if pos := strings.LastIndex(u, "."); pos < 0 || u[pos:] != ".flv" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	//處理后的URL 127.0.0.1:7001/live/monitor
	path := strings.TrimSuffix(strings.TrimLeft(u, "/"), ".flv")
	paths := strings.SplitN(path, "/", 2)
	log.Debug("url:", u, " path:", path, " paths:", paths)

	if len(paths) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// 判断视屏流是否发布,如果没有发布,直接返回404
	msgs := server.getStreams(w, r)
	if msgs == nil || len(msgs.Publishers) == 0 {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	} else {
		include := false
		for _, item := range msgs.Publishers {
			if item.Key == path {
				include = true
				break
			}
		}
		if include == false {
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	writer := NewFLVWriter(paths[0], paths[1], url, w)

	server.handler.HandleWriter(writer)
	writer.Wait()
}
