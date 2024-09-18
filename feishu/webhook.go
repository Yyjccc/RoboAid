package feishu

import (
	"RoboAid/core"
	"context"
	"encoding/json"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"net/http"
	"strconv"
	"time"
)

const date_format = "2006-01-02"

func ServerStart() {
	handler := dispatcher.NewEventDispatcher(cfg.VerifyToken, "").
		//接收到消息的处理
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			id := *event.Event.Sender.SenderId.OpenId
			return SendCard(NewTipCard("暂不支持聊天"), id)
		}).
		OnP2BotMenuV6(func(ctx context.Context, event *larkapplication.P2BotMenuV6) error {
			// 菜单类型
			menuType := *event.Event.EventKey
			openID := *event.Event.Operator.OperatorId.OpenId
			switch menuType {
			case "custom.bot.status":
				return SendCard(NewTipCard("bot在线中"), openID)
			case "custom.subscribe.status":
				//查询用户订阅状态
				//查询用户订阅状态
				info := fsDb.GetSubscribeInfo(openID)
				if info == nil {
					subscribeInfo := &SubscribeInfo{
						OpenId:     openID,
						Subscribe:  1,
						UpdateTime: time.Now(),
					}
					err := fsDb.InsertSubscribeInfo(subscribeInfo)
					if err != nil {
						return err
					}
					return SendCard(NewTipCard("你已**开启**漏洞信息订阅"), openID)
				} else {
					if info.Subscribe == 0 {
						return SendCard(NewTipCard("你已**关闭**信息订阅"), openID)
					} else {
						return SendCard(NewTipCard("你已**开启**信息订阅"), openID)
					}
				}
			case "custom.event.open":
				//开启订阅
				err := fsDb.UpdateSubscribeInfo(openID, 1)
				if err != nil {
					log.Error(err)
				}
				return SendCard(NewTipCard("**成功开启**订阅"), openID)
			case "custom.event.close":
				//开启订阅
				err := fsDb.UpdateSubscribeInfo(openID, 0)
				if err != nil {
					log.Error(err)
				}
				return SendCard(NewTipCard("**成功关闭**订阅"), openID)
			case "":

			}
			return nil
		}).
		//卡片回传
		OnCustomizedEvent("card.action.trigger", func(ctx context.Context, event *larkevent.EventReq) error {
			var callback CallBack
			err := json.Unmarshal(event.Body, &callback)
			if err != nil {
				log.Error(err)
				return err
			}
			openID := *callback.Event.Operator.OpenId
			form := callback.Event.Action.FormValue
			callType := callback.Event.Action.Value
			switch callType {
			case "add":
				source := &core.RssSource{
					Name:         form.Name,
					Link:         form.Link,
					Description:  form.Description,
					Creator:      openID,
					Public:       form.Public,
					CollectCount: 0,
					CollectDate:  time.Now().Format(date_format),
					UpdateTime:   time.Now().Format(date_format),
				}
				//添加订阅源
				if form.Public == 1 {
					//推送申请卡片

					//return SendCard()
				} else {
					//私有的直接存入
					err := core.RssDb.InsertRssSource(source)
					if err != nil {
						log.Error(err)
						return SendCard(NewTipCard("错误:"+err.Error()), openID)
					}
					return SendCard(NewTipCard("订阅成功"), openID)
				}
			case "apply":
				//公共RSS申请通过
			}
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

type RssService struct {
	//主定时器
	MainTask *core.ScheduledTask
}

func NewRssService() *RssService {
	taskFunc := func() {
		source, err := core.RssDb.GetAllRssSource()
		if err != nil {
			log.Error(err)
			return
		}
		for _, rssSource := range source {
			go Do(rssSource, time.Now())
		}
	}

	service := &RssService{
		MainTask: core.NewScheduledTask("RssService", 8, 0, taskFunc),
	}
	service.MainTask.Start()
	return service

}

func Do(rss *core.RssSource, t time.Time) {
	defer func() {
		if r := recover(); r != nil {
			// 捕获并处理panic
			log.Errorf("Recovered from panic: %v", r)
		}
	}()
	// 执行任务
	records := rss.Get(t)
	if len(records) != 0 {
		//写入数据库
		for _, record := range records {
			//替换html
			s, _ := Render(record.Description, core.ParseHostURL(record.Link))
			record.Description = s
			id, err := core.RssDb.InsertRssRecord(record)
			if err != nil {
				log.Error(err)
			}
			record.ID = id
			// 公开rss 向所有人推送
			if rss.Public == 1 {
				sender.SendPublicRecord(rss, record)
			} else {
				sender.SendPrivateRecord(rss, record)
			}
		}
	}

}

type CallBack struct {
	*larkevent.EventV2Base // 事件基础数据
	Event                  struct {
		Operator larkim.UserId `json:"operator"`
		Token    string        `json:"token"`
		Action   *ActionForm   `json:"action"`
	} `json:"event"`
}
type ActionForm struct {
	Value     string `json:"value"`
	Tag       string `json:"tag"`
	Timezone  string `json:"timezone"`
	FormValue struct {
		core.RssSource
		Note string `json:"note"`
	} `json:"form_value"`
	Name string `json:"name"`
}
