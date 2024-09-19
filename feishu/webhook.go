package feishu

import (
	"RoboAid/core"
	"context"
	"encoding/json"
	"fmt"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkapplication "github.com/larksuite/oapi-sdk-go/v3/service/application/v6"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"net/http"
	"strconv"
	"time"
)

const date_format = "2006-01-02"

var apply_list = make(map[string]*Apply)

func ServerStart() {
	handler := dispatcher.NewEventDispatcher(cfg.VerifyToken, "").
		//接收到消息的处理
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			msgId := *event.Event.Message.MessageId
			return ReplyCard(NewTipCard("暂不支持聊天"), msgId)
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
			case "event.rss.list":
				//rss订阅列表
				source, err := core.RssDb.GetAllRssSource()
				if err != nil {
					return SendCard(NewErrCard(err), openID)
				}
				var public_list = make([]*core.RssSource, 0)
				for _, v := range source {
					if v.Public == 1 {
						public_list = append(public_list, v)
					}
				}
				private_list, err := fsDb.GetAllPrivateRssByUserID(openID)
				if err != nil {
					return SendCard(NewErrCard(err), openID)
				}
				log.Infof("public RSS count: %d;private RSS count:%d", len(public_list), len(private_list))
				return SendCard(NewRssListCard(public_list, private_list), openID)
			case "event.rss.add":
				return SendCard(NewRssAddCard(), openID)
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
			callType := callback.Event.Action.Value.OpCode
			switch callType {
			case "add":
				// 将字符串转换为整数
				name := form.Name
				ex := apply_list[name]
				if ex != nil || queryByName(name) {
					return SendCard(NewErrCard(fmt.Errorf("%s 已经存在", name)), openID)
				}
				num, _ := strconv.Atoi(form.Type)
				source := &core.RssSource{
					Name:         form.Name,
					Link:         form.Link,
					Description:  form.Description,
					Creator:      openID,
					Public:       num,
					CollectCount: 0,
					CollectDate:  time.Now().Format(date_format),
					UpdateTime:   time.Now().Format(date_format),
				}
				//添加订阅源
				if source.Public == 1 {
					//推送申请卡片
					apply := NewApply(source, openID, callback.Event.Context.OpenMessageID, form.Note, true)
					log.Infof("generate apply: %v", apply)
					apply_list[name] = apply
					return SendCard(NewApplyCard(apply), cfg.Owner)
				} else {
					//私有的直接存入
					rssID, err := core.RssDb.InsertRssSource(source)
					if err != nil {
						log.Error(err)
						return SendCard(NewTipCard("错误:"+err.Error()), openID)
					}
					privateRss := &PrivateRss{
						ID:         0,
						SourceID:   rssID,
						OpenID:     openID,
						CreateDate: time.Now().Format(date_format),
					}
					_, err = fsDb.InsertPrivateRSS(privateRss)
					if err != nil {
						log.Error(err)
						return SendCard(NewTipCard("错误:"+err.Error()), openID)
					}
					return SendCard(NewTipCard("订阅成功"), openID)
				}
			case "del":
				rssId, _ := strconv.ParseInt(callback.Event.Action.Value.ApplyId, 10, 64)
				source := core.RssDb.GetRssSource(rssId)
				if source == nil {
					return SendCard(NewErrCard(fmt.Errorf("数据库错误：%v", err)), openID)
				}
				if source.Public == 1 {
					// 发送申请
					apply := NewApply(source, openID, callback.Event.Context.OpenMessageID, "", false)
					log.Infof("generate apply: %v", apply)
					apply_list[source.Name] = apply
					return SendCard(NewApplyCard(apply), cfg.Owner)
				} else {
					// 直接删除
					log.Debugf("准备删除：%s", source.Name)
					err := core.RssDb.DeleteRssSource(source.Name)
					if err != nil {
						log.Error(err)
						return SendCard(NewErrCard(err), openID)
					}
					err = fsDb.DelPrivateRSS(rssId)
					if err != nil {
						log.Error(err)
						return SendCard(NewErrCard(err), openID)
					}
					return SendCard(NewTipCard("取消订阅成功"), openID)
				}
			case "pass":
				//公共RSS申请通过
				log.Debugf("申请等待队列：%v", apply_list)
				apply := GetApply(callback.Event.Action.Value.ApplyId)
				if apply == nil {
					return SendCard(NewErrCard(fmt.Errorf("申请流程错误,申请对象为空!")), openID)
				}
				//剔除等待队列
				delete(apply_list, apply.Source.Name)
				if apply.add {
					//添加操作
					id, err := core.RssDb.InsertRssSource(apply.Source)
					if err != nil {
						log.Error(err)
						return SendCard(NewErrCard(fmt.Errorf("订阅错误,%s", err)), openID)
					}
					apply.Source.ID = id
					//向申请者发送审核通过
					//撤回原来的卡片
					SendCard(NewTipCard("审核已通过,订阅成功："+apply.Source.Name), apply.UserId)
					//向审核者发送提示
					return SendCard(NewTipCard("审核已生效,"+apply.Source.Name), openID)
				} else {
					//删除操作
					err := core.RssDb.DeleteRssSource(apply.Source.Name)
					if err != nil {
						return SendCard(NewErrCard(fmt.Errorf("取消订阅错误,%s", err)), openID)
					}

					SendCard(NewTipCard("审核已通过,取消订阅成功："+apply.Source.Name), apply.UserId)
					//向审核者发送提示
					return SendCard(NewTipCard("审核已生效,"+apply.Source.Name), openID)
				}
			case "reject":
				apply := GetApply(callback.Event.Action.Value.ApplyId)
				if apply == nil {
					return SendCard(NewErrCard(fmt.Errorf("申请流程错误,申请对象为空!")), openID)
				}
				//剔除等待队列
				delete(apply_list, apply.Source.Name)
				return SendCard(NewTipCard("审核未通过！name:"+apply.Source.Name), apply.UserId)
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
		Operator larkim.UserId    `json:"operator"`
		Token    string           `json:"token"`
		Action   *ActionForm      `json:"action"`
		Context  callback.Context `json:"context"`
	} `json:"event"`
}
type ActionForm struct {
	Value     CallbackEvent `json:"value"`
	Tag       string        `json:"tag"`
	Timezone  string        `json:"timezone"`
	FormValue struct {
		core.RssSource
		Note string `json:"note"`
		Type string `json:"public"`
	} `json:"form_value"`
	Name string `json:"name"`
}

type CallbackEvent struct {
	OpCode  string `json:"opcode"`
	ApplyId string `json:"apply_id"`
}

func queryByName(name string) bool {
	return core.RssDb.HasRss(name)
}

func GetApply(id string) *Apply {
	for _, apply := range apply_list {
		if apply.Id == id {
			return apply
		}
	}
	return nil
}
