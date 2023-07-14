package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	lg := logrus.New()
	// 要访问的第一页URL
	firstPageURL := "https://www.23us.cc/html/661/661753/8651369.html"
	// 创建一个上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// 设置浏览器选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // 是否禁用页面
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		// chromedp.Flag("remote-debugging-port", "9222"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	// 创建一个浏览器实例
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	// 初始化变量
	var title, content, nextPageURL string
	nextPageURL = firstPageURL
	var body string
	// 打开第一页
	err := chromedp.Run(ctx, chromedp.Navigate(firstPageURL))
	if err != nil {
		lg.Error(err)
		return
	}
	f, err := os.OpenFile("novel.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		lg.Error(err)
		return
	}
	defer func() {
		_ = f.Sync()
		_ = f.Close()
	}()
	// 爬取小说内容
	for {
		// 获取标题、内容和下一页链接
		err = chromedp.Run(ctx,
			chromedp.Text(".title", &title, chromedp.ByQuery),
			chromedp.Text("#content", &content, chromedp.ByQuery),
			chromedp.OuterHTML("body", &body),
		)
		if err != nil {
			lg.Error(err)
			return
		}
		//仅在下一章时记录标题
		if !strings.Contains(nextPageURL, "_2") && !strings.Contains(nextPageURL, "_3") {
			appendToFile(f, "\n　　"+strings.TrimSpace(title)+"\n")
			lg.Info(title)
		}
		// 定义正则表达式
		pattern := regexp.MustCompile(`<a href="([^"]+)"[^>]*>(下一页|下一章)<\/a>`)
		// 在HTML字符串中查找匹配项
		match := pattern.FindStringSubmatch(body)
		if len(match) > 0 {
			nextPageURL = "https://www.23us.cc" + match[1]
		} else {
			lg.Error("没有找到下一页或下一章的链接")
			return
		}
		// 记录内容
		content = strings.ReplaceAll(content, "章节错误,点此举报(免注册),举报后维护人员会在两分钟内校正章节内容,请耐心等待,并刷新页面。", "")
		content = strings.TrimSpace(content)
		appendToFile(f, "　　"+content+"\n")
		// 检查是否到达最后一页
		if strings.Contains(nextPageURL, "1000000.html") {
			break
		}
		lg.Info(nextPageURL)
		// 跳转到下一页
		for i := 0; i < 100; i++ {
			ctx2, cc := context.WithTimeout(ctx, time.Second*10)
			err = chromedp.Run(ctx2, chromedp.Navigate(nextPageURL))
			if err != nil {
				if i > 90 {
					cc()
					return
				}
				lg.Error(err)
				time.Sleep(time.Second * 10)
				continue
			}
			break
		}
	}
	lg.Info("爬取完成")
}

// 将文本追加到文件中
func appendToFile(f *os.File, text string) {
	_, _ = f.WriteString(text)
}
