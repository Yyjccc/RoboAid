package feishu

import (
	"RoboAid/core"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// 提示卡片
func NewTipCard(content string) string {
	var variables = make(map[string]interface{})
	variables["content"] = content
	variables["quote_daily"] = getQuoteDaily()
	card := &Card{
		Type: "template",
		Data: CardTemplate{
			TemplateID:          cfg.GetTmpl("tips").ID,
			TemplateVersionName: cfg.GetTmpl("tips").Version,
			TemplateVariable:    variables,
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		log.Error(err)
	}
	log.Infof("创建tip卡片:%s", content)
	return string(data)
}

// 错误提示
func NewErrCard(err error) string {
	return NewTipCard("发生错误:" + err.Error())
}

func NewRSSCard(source *core.RssSource, record *core.RssRecord) string {
	var variables = make(map[string]interface{})
	variables["title"] = record.Title
	variables["content"] = record.Description
	variables["source"] = source.Show(record)
	variables["creator"] = source.Creator
	variables["quote_daily"] = getQuoteDaily()
	variables["link"] = record.Link
	card := &Card{
		Type: "template",
		Data: CardTemplate{
			TemplateID:          cfg.GetTmpl("rss").ID,
			TemplateVersionName: cfg.GetTmpl("rss").Version,
			TemplateVariable:    variables,
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		log.Error(err)
	}
	return string(data)
}

type Card struct {
	Type string       `json:"type"`
	Data CardTemplate `json:"data"`
}

type CardTemplate struct {
	TemplateID          string                 `json:"template_id"`
	TemplateVersionName string                 `json:"template_version_name"`
	TemplateVariable    map[string]interface{} `json:"template_variable"`
}

type Apply struct {
	Id        string
	Date      string
	UserId    string
	Source    *core.RssSource
	SrcCardId string
	add       bool
	Note      string
}

func NewApply(s *core.RssSource, openID, srcCardID, note string, add bool) *Apply {
	// 生成 UUID
	id := uuid.New()
	// 取前10位字符作为ID
	shortID := id.String()[:10]
	log.Infof("create apply,id:%s", shortID)

	return &Apply{
		Id:        shortID,
		Date:      time.Now().Format(date_format),
		UserId:    openID,
		Source:    s,
		SrcCardId: srcCardID,
		add:       add,
		Note:      note,
	}
}

func NewRssAddCard() string {
	var variables = make(map[string]interface{})
	variables["quote_daily"] = getQuoteDaily()
	card := &Card{
		Type: "template",
		Data: CardTemplate{
			TemplateID:          cfg.GetTmpl("addRss").ID,
			TemplateVersionName: cfg.GetTmpl("addRss").Version,
			TemplateVariable:    variables,
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		log.Error(err)
	}
	return string(data)
}

// RSS列表卡片
func NewRssListCard(public, private []*core.RssSource) string {
	var variables = make(map[string]interface{})
	variables["prviate_list"] = private
	variables["public_list"] = public
	variables["quote_daily"] = getQuoteDaily()
	card := &Card{
		Type: "template",
		Data: CardTemplate{
			TemplateID:          cfg.GetTmpl("rssList").ID,
			TemplateVersionName: cfg.GetTmpl("rssList").Version,
			TemplateVariable:    variables,
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		log.Error(err)
	}
	log.Info("创建RSS列表卡片")
	return string(data)

}

// 申请卡片
func NewApplyCard(apply *Apply) string {

	source := apply.Source
	var variables = make(map[string]interface{})
	variables["apply_id"] = apply.Id
	variables["user_id"] = apply.UserId
	variables["name"] = source.Name
	variables["link"] = source.Link
	variables["description"] = source.Description
	variables["date"] = apply.Date
	if apply.add {
		variables["op"] = "添加"
	} else {
		variables["op"] = "删除"
	}
	variables["note"] = apply.Note
	variables["quote_daily"] = getQuoteDaily()
	card := &Card{
		Type: "template",
		Data: CardTemplate{
			TemplateID:          cfg.GetTmpl("apply").ID,
			TemplateVersionName: cfg.GetTmpl("apply").Version,
			TemplateVariable:    variables,
		},
	}
	data, err := json.Marshal(card)
	if err != nil {
		log.Error(err)
	}
	log.Info("创建申请卡片")
	return string(data)
}
