package core

import (
	"RoboAid/config"
	"fmt"
	"github.com/mmcdole/gofeed"
	"time"
)

var (
	log = DefaultLogger
	cfg = config.BotConfig
)

const date_format = "2006-01-02"

// RssSource rss订阅源
type RssSource struct {
	ID           int64
	Name         string `json:"name"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	Creator      string `json:"user_id"`
	Public       int
	CollectCount int
	CollectDate  string
	UpdateTime   string
}

type RssRecord struct {
	ID          int64
	SourceID    int64
	Description string
	Title       string
	Link        string
	PublishDate string
	Author      string
}

func (r *RssSource) Get(t time.Time) []*RssRecord {
	// 创建解析器
	fp := gofeed.NewParser()
	//var res = make([]*RssRecord, 0)
	// 从 RSS 源获取 feed
	feed, err := fp.ParseURL(r.Link)
	if err != nil {
		log.Errorf("读取 RSS ,%s失败: %v", r.Link, err)
		return nil
	}
	// 打印每篇文章的标题和链接
	var res []*RssRecord
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}
		//昨天发布的，进行记录和推送
		if IsYesterday(item.PublishedParsed, t) {
			author := ""
			if item.Author != nil {
				author = item.Author.Name
			}
			record := &RssRecord{
				SourceID:    r.ID,
				Description: item.Description,
				Title:       item.Title,
				Link:        item.Link,
				PublishDate: item.PublishedParsed.Format(date_format),
				Author:      author,
			}
			res = append(res, record)
		}
	}
	return res
}

func (r *RssSource) Show(record *RssRecord) string {

	author := ""
	if record.Author != "" {
		author = fmt.Sprintf("- 作者:%s", record.Author)
	}

	return fmt.Sprintf(`
- 名称:   %s
- 描述:   %s
- 链接:   %s
%s
	`, r.Name, r.Description, ParseHostURL(r.Link), author)
}
