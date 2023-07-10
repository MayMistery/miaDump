package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func downloadPage(url string, savePath string, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error downloading page %s: %s\n", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body for %s: %s\n", url, err)
		return
	}

	fileName := filepath.Base(url)
	filePath := filepath.Join(savePath, fileName)
	fmt.Printf(savePath)

	err = os.WriteFile(savePath, body, 0644)
	if err != nil {
		fmt.Printf("Error saving file %s: %s\n", filePath, err)
		return
	}

	fmt.Printf("Page %s downloaded and saved at %s\n", url, filePath)
}

func hello() {
	start := 1                  // 起始页面
	end := 10                   // 结束页面
	savePath := "/tttang"       // 保存路径
	maxThreads := 5             // 最大线程数
	timeout := 10 * time.Second // 请求超时时间

	// 确保保存路径存在
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		os.Mkdir(savePath, os.ModePerm)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxThreads)

	for i := start; i <= end; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(pageNum int) {
			defer func() { <-semaphore }()
			url := fmt.Sprintf("http://tttang.com/archive/%d/", pageNum)
			downloadPage(url, savePath, timeout, &wg)
		}(i)
	}

	wg.Wait()
	close(semaphore)
}
