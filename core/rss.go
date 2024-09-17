package core

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"time"
)

var log = DefaultLogger

// RssSource rss订阅源
type RssSource struct {
	Name        string
	Link        string
	Description string
	Creator     string
	CreateTime  time.Time
	UpdateTime  time.Time
}

func T() {
	// 创建解析器
	fp := gofeed.NewParser()
	// 替换为你想订阅的 RSS 链接
	rssURL := "https://www.leavesongs.com/feed/"

	// 从 RSS 源获取 feed
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		fmt.Printf("读取 RSS 失败: %v\n", err)
		return
	}

	// 打印 feed 信息
	fmt.Printf("Feed 标题: %s\n", feed.Title)
	fmt.Printf("Feed 描述: %s\n", feed.Description)

	// 打印每篇文章的标题和链接
	for _, item := range feed.Items {
		fmt.Printf("\n文章标题: %s\n", item.Title)
		fmt.Printf("链接: %s\n", item.Link)
	}
}

func (r *RssSource) Get() {

}
