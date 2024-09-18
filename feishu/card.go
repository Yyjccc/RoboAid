package feishu

import (
	"RoboAid/core"
	"encoding/json"
)

// 提示卡片
func NewTipCard(content string) string {
	var variables = make(map[string]string)
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

func NewRSSCard(source *core.RssSource, record *core.RssRecord) string {
	var variables = make(map[string]string)
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
	TemplateID          string            `json:"template_id"`
	TemplateVersionName string            `json:"template_version_name"`
	TemplateVariable    map[string]string `json:"template_variable"`
}

type Apply struct {
	Id     string
	Date   string
	UserId string
	Source *core.RssSource
	add    bool
	Note   string
}

// 申请卡片
func NeaApplyCard(apply Apply) string {

	source := apply.Source
	var variables = make(map[string]string)
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
