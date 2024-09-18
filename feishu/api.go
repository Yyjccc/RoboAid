package feishu

import (
	"RoboAid/config"
	"RoboAid/core"
	"context"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"math/rand"
	"net/http"
	"regexp"
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
	sender *Sender
)

func init() {
	sender = &Sender{
		DataChannel: make(chan *RecordWrapper, 5),
		StopChannel: make(chan bool),
	}
	sender.StartDataListener()
}

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

// 上传图片
func Upload(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// 检查请求是否成功
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image: %s, status code: %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()
	createImageReq := larkim.NewCreateImageReqBuilder().
		Body(larkim.NewCreateImageReqBodyBuilder().
			ImageType("message").
			Image(resp.Body).
			Build()).
		Build()
	createImageResp, err := client.Im.Image.Create(context.Background(), createImageReq)
	if err != nil {
		return "", err
	}
	if !createImageResp.Success() {
		return "", fmt.Errorf("client.Im.Image.Create failed, code: %d, msg: %s, log_id: %s\n",
			createImageResp.Code, createImageResp.Msg, createImageResp.RequestId())
	} else {
		key := *createImageResp.Data.ImageKey
		log.Infof("上传图片：%s,%s", url, key)
		return key, nil
	}
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

	createMessageResp, _ := client.Im.Message.Create(ctx, createMessageReq)
	if !createMessageResp.Success() {
		log.Errorf("client.Im.Message.Create failed, code: %d, msg: %s, log_id: %s\n",
			createMessageResp.Code, createMessageResp.Msg, createMessageResp.RequestId())
		log.Error("飞书bot", "消息发送失败")
		return createMessageResp.CodeError
	} else {
		log.Debugf("bot 成功推送消息卡片,openId:%s", openId)
	}
	return nil
}

// 通过配置的群聊获取用户列表
func GetUserList() []string {
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

type RecordWrapper struct {
	Record *core.RssRecord
	Public bool
	OpenId string
	Source *core.RssSource
}

type Sender struct {
	DataChannel chan *RecordWrapper

	StopChannel chan bool
}

func (s *Sender) StartDataListener() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 捕获并处理panic
				log.Errorf("Recovered from panic: %v", r)
			}
		}()
		for {
			// 创建一个用于接收数据的 channel
			select {
			case <-s.StopChannel: // 接收到停止信号，退出循环
				return
			default:
				for {
					// 接收 channels 进行数据接收
					data, ok := <-s.DataChannel
					if !ok {
						break
					}
					if data.Public {
						list := GetUserList()
						for _, v := range list {
							s.SendTo(v, data.Source, data.Record)
						}
					} else {
						s.SendTo(data.OpenId, data.Source, data.Record)

					}
					time.Sleep(1 * time.Minute)
				}
				// 添加一个小的休眠时间，以减少 CPU 使用率
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (s *Sender) Stop() {
	s.StopChannel <- true
}
func (s *Sender) SendPublicRecord(source *core.RssSource, record *core.RssRecord) {
	wrapper := &RecordWrapper{
		Source: source,
		Record: record,
		Public: true,
		OpenId: "",
	}
	s.DataChannel <- wrapper
}

func (s *Sender) SendPrivateRecord(source *core.RssSource, record *core.RssRecord) error {
	//查询 openid
	rss, err := fsDb.GetAllPrivateRss(record.SourceID)
	if err != nil {
		return err
	}
	if len(rss) == 0 {
		err := fmt.Errorf("not found private rss resouce user,%s", record.SourceID)
		log.Error(err)
		return err
	}
	//遍历发送
	for _, v := range rss {
		wrapper := &RecordWrapper{
			Source: source,
			Record: record,
			Public: false,
			OpenId: v.OpenID,
		}
		s.DataChannel <- wrapper
	}
	return nil
}

func (s *Sender) SendTo(openID string, source *core.RssSource, record *core.RssRecord) error {
	//查询是否开启订阅
	info := fsDb.GetSubscribeInfo(openID)
	if info == nil {
		subscribeInfo := &SubscribeInfo{
			OpenId:     openID,
			Subscribe:  1,
			UpdateTime: time.Now(),
		}
		err := fsDb.InsertSubscribeInfo(subscribeInfo)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		//关闭推送的返回
		if info.Subscribe == 0 {
			return nil
		}
	}
	//如果开启订阅才推送
	//发送消息
	err := SendCard(NewRSSCard(source, record), openID)
	if err != nil {
		log.Error(err)
		return err
	}
	//推送记录,存入数据库
	err = fsDb.InsertPushRecord(openID, record.ID)
	if err != nil {
		log.Error(err)
		return err

	}
	return nil
}

// html转化为md 并上传图片
func Render(html string, baseURL string) (string, error) {
	// 创建转换器
	converter := md.NewConverter("", true, nil)
	// 将HTML转换为Markdown
	markdown, err := converter.ConvertString(html)
	if err != nil {
		log.Error(err)
		return html, err
	}

	// 定义正则表达式来匹配外层的 Markdown 链接格式 [![image](image-url)](link-url)
	re := regexp.MustCompile(`\[(\!\[.*?\]\(.*?\))\]\(.*?\)`)

	// 使用正则替换，把外层的 []() 替换掉，仅保留 ![]() 的内容
	result := re.ReplaceAllString(markdown, "$1")

	//替换图片
	// 定义正则表达式来匹配 Markdown 中的图片链接
	// `!\[.*?\]\((.*?)\)` 匹配 ![alt text](url)，其中 (.*?) 是要提取的 URL 部分
	imageRegex := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)

	// 替换图片链接
	replacedMarkdown := imageRegex.ReplaceAllStringFunc(result, func(match string) string {
		// 提取当前图片链接
		originalURL := imageRegex.FindStringSubmatch(match)[1]
		// 替换为新链接
		newURL := baseURL + originalURL
		imgKey, err := Upload(newURL)
		if err != nil {
			log.Error(err)
			return fmt.Sprintf("![image](%s)", originalURL)
		} else {
			return fmt.Sprintf("![image](%s)", imgKey)
		}
	})
	return replacedMarkdown, nil
}
