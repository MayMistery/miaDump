package dump

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type blogSite struct {
	urlTemplate string
	start       int
	end         int
	path        string
	title       string
}

var tttang = blogSite{
	urlTemplate: "http://tttang.com/archive/{start-end}",
	start:       1,
	//end:         300,
	end:   1900,
	path:  "./tttang/",
	title: "h2.mb-3",
}

var xz = blogSite{
	urlTemplate: "https://xz.aliyun.com/t/{start-end}/",
	start:       1,
	end:         100,
	//end:   100,
	path:  "./xz/",
	title: "span.content-title",
}

func Dump() {

}

func generateURL(urlTemplate string, i int) string {
	// 替换URL中的变量部分
	placeholder := "{start-end}"
	url := strings.Replace(urlTemplate, placeholder, strconv.Itoa(i), 1)
	return url
}

func sanitizeFilename(filename string) string {
	// 定义不允许的字符正则表达式
	invalidChars := regexp.MustCompile(`[\\/:"*?<>|]`)

	// 使用空字符串替换不允许的字符
	sanitized := invalidChars.ReplaceAllString(filename, "-")

	return sanitized
}

func dumpHtml(site blogSite) {
	var wg sync.WaitGroup
	wg.Add(site.end - site.start + 1)

	// 控制并发的通道
	semaphore := make(chan struct{}, 3) // Change the value to adjust the scanning speed

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	var errorURLs []string
	var errorMutex sync.Mutex

	for i := site.start; i <= site.end; i++ {
		go func(i int) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire a semaphore

			url := generateURL(site.urlTemplate, i)

			retryCount := 3
			for retry := 0; retry < retryCount; retry++ {
				response, err := client.Get(url)
				if err != nil {
					log.Printf("无法下载网页 %s: %v\n", url, err)
					if retry < retryCount-1 {
						// 添加延迟，然后继续重试
						time.Sleep(1 * time.Second)
						continue
					} else {
						// 保存错误的 URL
						errorMutex.Lock()
						errorURLs = append(errorURLs, url)
						errorMutex.Unlock()
						<-semaphore // Release the semaphore before returning
						return
					}
				}
				defer response.Body.Close()

				// 读取 response.Body 到一个字节数组中
				bodyBytes, err := io.ReadAll(response.Body)
				if err != nil {
					log.Printf("无法读取 response.Body: %s, %v \n", url, err)
					if retry < retryCount-1 {
						// 添加延迟，然后继续重试
						time.Sleep(1 * time.Second)
						continue
					} else {
						// 保存错误的 URL
						errorMutex.Lock()
						errorURLs = append(errorURLs, url)
						errorMutex.Unlock()
						<-semaphore // Release the semaphore before returning
						return
					}
				}

				// 将字节数组转换为字符串
				htmlContent := string(bodyBytes)

				// 使用 goquery 解析字符串内容
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
				if err != nil {
					log.Printf("无法解析HTML内容: %v\n", err)
					if retry < retryCount-1 {
						// 添加延迟，然后继续重试
						time.Sleep(1 * time.Second)
						continue
					} else {
						// 保存错误的 URL
						errorMutex.Lock()
						errorURLs = append(errorURLs, url)
						errorMutex.Unlock()
						<-semaphore // Release the semaphore before returning
						return
					}
				}

				element := doc.Find(site.title).First()

				// 检查是否找到了匹配的元素
				if element.Length() == 0 {
					//fmt.Println("没有找到匹配的元素")
					<-semaphore // Release the semaphore before returning
					return
				}

				// 提取内容
				title := element.Text()
				//fmt.Printf("提取的内容：%s\n", title)

				sanitizedTitle := sanitizeFilename(title)
				filename := site.path + strconv.Itoa(i) + "-" + sanitizedTitle + ".html"

				file, err := os.Create(filename)
				if err != nil {
					log.Printf("无法创建文件 %s: %v\n", filename, err)
					<-semaphore // Release the semaphore before returning
					return
				}
				defer file.Close()

				// 将网页内容写入文件
				_, err = io.Copy(file, strings.NewReader(htmlContent))
				if err != nil {
					log.Printf("无法保存网页内容到文件 %s: %v\n", filename, err)
					<-semaphore // Release the semaphore before returning
					return
				}

				//fmt.Printf("已保存网页 %s\n", url)

				// 添加延迟
				time.Sleep(1 * time.Second) // Change the duration to adjust the delay

				<-semaphore // Release the semaphore after completing the task

				// 如果成功下载和保存了网页内容，则跳出重试循环
				break
			}
		}(i)
	}

	wg.Wait()

	// 在所有请求完成后处理错误的 URL
	if len(errorURLs) > 0 {
		log.Println("以下 URL 下载失败:")
		for _, url := range errorURLs {
			log.Println(url)
		}
	}
}

func dumpPics() {

}

func replaceUri() {

}
