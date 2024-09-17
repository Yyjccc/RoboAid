package feishu

import (
	"context"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"net/http"
	"strconv"
)

func ServerStart() {
	handler := dispatcher.NewEventDispatcher("", "").
		//接收到消息的处理
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {

			return nil
		})

	hookHandler := httpserverext.NewEventHandlerFunc(handler, larkevent.WithLogLevel(larkcore.LogLevelDebug))
	http.HandleFunc("/", hookHandler)
	log.Info("start webhook on ", cfg.ServerPort)
	// 启动 http 服务
	err := http.ListenAndServe(":"+strconv.Itoa(cfg.ServerPort), nil)
	if err != nil {
		panic(err)
	}
}
