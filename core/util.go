package core

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

func FormatDuration(d time.Duration) string {
	switch {
	case d >= time.Hour:
		return fmt.Sprintf("%.2f hours", d.Hours())
	case d >= time.Minute:
		return fmt.Sprintf("%.2f minutes", d.Minutes())
	case d >= time.Second:
		return fmt.Sprintf("%.2f seconds", d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%v milliseconds", d.Milliseconds())
	case d >= time.Microsecond:
		return fmt.Sprintf("%v microseconds", d.Microseconds())
	default:
		return fmt.Sprintf("%v nanoseconds", d.Nanoseconds())
	}
}

// IsYesterday 判断给定的时间是否为昨天,方便测试
func IsYesterday(date *time.Time, raw time.Time) bool {
	// 获取当前时间
	now := raw

	// 获取昨天的日期
	yesterday := now.AddDate(0, 0, -1)

	// 将输入的时间和昨天的时间的日期部分进行比较（只比较日期，忽略时分秒）
	return date.Year() == yesterday.Year() &&
		date.Month() == yesterday.Month() &&
		date.Day() == yesterday.Day()
}

// containsIgnoreCase 检查字符串 s 是否包含 substr，忽略大小写
func ContainsIgnoreCase(s string, substr ...string) bool {
	for _, sub := range substr {
		if strings.Contains(strings.ToLower(s), strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func GenerateFilename(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	filename := make([]byte, length)
	for i := range filename {
		filename[i] = charset[rand.Intn(len(charset))]
	}
	return string(filename)
}

func Trim(s string) string {
	// 去除字符串两端的空白字符（包括空格、换行符等）
	trimmedStr := strings.TrimSpace(s)
	// 去除字符串中所有的空白字符
	noWhitespaceStr := strings.Join(strings.Fields(trimmedStr), "")
	return noWhitespaceStr
}

// EnsureDirectoryExists 检查目录是否存在，如果不存在则逐级创建
func EnsureDirectoryExists(path string) error {
	// 检查目录是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 目录不存在，逐级创建
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create directory %s: %v", path, err)
		}
		log.Debug("Directory created:", path)
	}
	//else {
	//log.Debug("Directory already exists:", path)
	//}
	return nil
}

func compareDates(date1, date2 time.Time) bool {
	return date1.Year() == date2.Year() &&
		date1.Month() == date2.Month() &&
		date1.Day() == date2.Day()
}

// 判断切片中是否存在
func Contains(slice []string, item string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}

// 定时任务结构体
type ScheduledTask struct {
	TargetTime  time.Time    // 每天执行的目标时间
	TaskFunc    func()       // 任务函数
	ticker      *time.Ticker // 定时器
	stopChannel chan bool    // 停止信号通道
	Name        string
}

// 创建新的定时任务
func NewScheduledTask(name string, targetHour int, targetMinute int, taskFunc func()) *ScheduledTask {
	// 计算今天的目标时间
	now := time.Now()
	targetTime := time.Date(now.Year(), now.Month(), now.Day(), targetHour, targetMinute, 0, 0, now.Location())

	// 如果目标时间已经过去，设置为明天的这个时间
	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}
	return &ScheduledTask{
		TargetTime:  targetTime,
		TaskFunc:    taskFunc,
		stopChannel: make(chan bool),
		Name:        name,
	}
}

// 计算下一个任务执行时间
func (st *ScheduledTask) getNextRunTime() time.Duration {
	now := time.Now()

	// 如果当前时间已经过了目标时间，设置为明天的目标时间
	if now.After(st.TargetTime) {
		st.TargetTime = st.TargetTime.Add(24 * time.Hour)
	}

	// 计算从现在到下次执行时间的时间间隔
	return time.Until(st.TargetTime)
}

// 启动定时任务
func (st *ScheduledTask) Start() {
	go func() {
		log.Infof("%s Scheduled Task Register, Next Execution at: %v", st.Name, st.TargetTime)
		for {
			// 计算下一次任务执行时间
			waitDuration := st.getNextRunTime()

			// 打印等待时间
			log.Infof("%s will start after: %v", st.Name, waitDuration)

			// 等待到达下一个执行时间
			select {
			case <-time.After(waitDuration):
				log.Infof("%s Task Start Executing", st.Name)
				st.TaskFunc()
				log.Infof("%s Task finished", st.Name)
				// 更新目标时间为第二天
				st.TargetTime = st.TargetTime.Add(24 * time.Hour)
			case <-st.stopChannel:
				log.Infof("%s Scheduled Task Stopped.", st.Name)
				return
			}
		}
	}()
}

// 停止定时任务
func (st *ScheduledTask) Stop() {
	st.stopChannel <- true
}

func ParseHostURL(rawUrl string) string {
	// 解析 URL
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Error(err)
		return rawUrl
	}
	// 构建新的 URL 只保留协议和主机部分
	cleanUrl := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	return cleanUrl
}
