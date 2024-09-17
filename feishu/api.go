package feishu

import (
	"DoRssBot/config"
	"DoRssBot/core"
	"context"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"math/rand"
	"net/http"
	"time"
)

var (
	ctx    = context.Background()
	log    = core.DefaultLogger
	cfg    = config.BotConfig
	client = lark.NewClient(cfg.AppId, cfg.SecretKey,
		lark.WithLogLevel(larkcore.LogLevelDebug),
		lark.WithReqTimeout(10*time.Second),
		lark.WithEnableTokenCache(true),
		lark.WithHelpdeskCredential("id", "token"),
		lark.WithHttpClient(http.DefaultClient))
)

// 每日金句
type QuoteDaily struct {
	Id      int    `json:"id,omitempty"`
	Content string `json:"hitokoto,omitempty"`
}

// 发送消息
func SendText(content, openid string) error {
	text := larkim.NewTextMsgBuilder().Text(content).Build()
	// 发送消息
	createMessageReq := larkim.NewCreateMessageReqBuilder().
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType("text").
			Content(text).
			ReceiveId(openid).
			Build()).
		ReceiveIdType("open_id").
		Build()

	createMessageResp, err := client.Im.Message.Create(ctx, createMessageReq)
	if !createMessageResp.Success() {
		log.Errorf("client.Im.Message.Create failed, code: %d, msg: %s, log_id: %s\n",
			createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
		log.Error("飞书bot", "消息发送失败")

		return createMessageResp.CodeError
	}
	return err
}

// 回复消息
func ReplyText(context context.Context, content, msgId string) error {
	text := larkim.NewTextMsgBuilder().Text(content).Build()
	replyReq := larkim.NewReplyMessageReqBuilder().Body(
		larkim.NewReplyMessageReqBodyBuilder().
			MsgType("text").
			Content(text).
			Build()).
		MessageId(msgId).Build()
	replyResp, err := client.Im.Message.Reply(context, replyReq)
	if !replyResp.Success() {
		log.Errorf("client.Im.Message.Create failed, code: %d, msg: %s, log_id: %s\n",
			replyResp.Code, replyResp.Msg, replyResp.RequestId())
		log.Error("飞书bot", "消息发送失败")
		return replyResp.CodeError
	} else {
		log.Debugf("bot 发送消息: %s", content)
	}
	return err
}

// 发送卡片消息
func SendCard(content, openId string) error {
	// 发送消息
	createMessageReq := larkim.NewCreateMessageReqBuilder().
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType("interactive").
			Content(content).
			ReceiveId(openId).
			Build()).
		ReceiveIdType("open_id").
		Build()

	createMessageResp, err := client.Im.Message.Create(ctx, createMessageReq)
	if !createMessageResp.Success() {
		log.Errorf("client.Im.Message.Create failed, code: %d, msg: %s, log_id: %s\n",
			createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
		log.Error("飞书bot", "消息发送失败")

		return createMessageResp.CodeError
	} else {
		log.Debugf("bot 成功推送消息卡片,openId:%s", openId)
	}
	return err
}

// 通过配置的群聊获取用户列表
func getUserList() []string {
	req := larkim.NewGetChatMembersReqBuilder().ChatId(cfg.GroupID).Build()
	var res []string
	resp, err := client.Im.ChatMembers.Get(ctx, req)
	if err != nil {
		return make([]string, 0)
	}
	items := resp.Data.Items
	for _, item := range items {
		res = append(res, *item.MemberId)
	}
	return res

}

// 获取每日一句
func getQuoteDaily() string {
	get, err := core.Http.Request(context.Background()).
		SetResult(QuoteDaily{}).Get("https://v1.hitokoto.cn")
	if err != nil {
		log.Errorf("api error,%s", err)
		return randomQuoteDaily()
	}
	daily := get.Result().(*QuoteDaily)
	if daily.Content == "" {
		return randomQuoteDaily()
	}
	return daily.Content + randomEmoji()
}

// 随机产生一句
func randomQuoteDaily() string {
	QuoteList := []string{
		"大本钟下送快递——上面摆，下面寄。", "浮世景色百千年依旧，人之在世却如白露与泡影。", "来人间一趟 你要看看太阳。",
		"人类把最精密的保密系统，都用在了自我毁灭上。", "相逢一醉是前缘，风雨散、飘然何处。", "希望你别像风，在我这里面掀起万翻般波澜，却又跟云去了远方。",
		"何须浅碧深红色，自是花中第一流。", "若你困与无风之地，我将奏响高天之歌。", "断剑重铸之日，骑士归来之时。",
	}
	index := rand.Intn(len(QuoteList))
	return QuoteList[index] + randomEmoji()
}

func randomEmoji() string {
	EmojiList := []string{
		"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🥲", "😊", "😇", "🤣",
		"🙂", "🙃", "😉", "🥰", "😍", "🤩", "😘", "😗", "😚", "😙", "😋",
		"😛", "😜", "🤪", "😝", "🤑", "🤗", "🤭", "🤔", "😴", "🥵", "🤓",
		"🤠", "🧐", "😦", "🥺", "😨", "😰", "😭", "😡", "🤬", "😥", "😫",
	}
	index := rand.Intn(len(EmojiList))
	return EmojiList[index]
}
