package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Host() {
	// 获取当前工作目录
	dir, err := os.Getwd()
	dir += "/tttang"
	if err != nil {
		fmt.Println("无法获取当前目录：", err)
		return
	}

	// 设置处理静态文件的处理器
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 设置目录处理器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 如果请求的是根路径，则重定向到 /index
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}

		// 读取目录下的文件列表
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintln(w, "无法读取目录：", err)
			return
		}

		// 构建目录页面内容
		var content strings.Builder
		content.WriteString("<h1>目录</h1>")
		content.WriteString("<ul>")
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".html" {
				content.WriteString("<li><a href=\"/static/")
				content.WriteString(file.Name())
				content.WriteString("\">")
				content.WriteString(file.Name())
				content.WriteString("</a></li>")
				// 读取 HTML 文件内容
				filePath := filepath.Join(dir, file.Name())
				fileContent, err := ioutil.ReadFile(filePath)
				if err != nil {
					fmt.Fprintf(w, "无法读取文件 %s：%s\n", file.Name(), err)
					continue
				}
				// 在文件内容开头插入 UTF-8 编码声明
				fileContent = append([]byte("<meta charset=\"UTF-8\">\n"), fileContent...)
				// 保存带有 UTF-8 编码声明的 HTML 文件
				err = ioutil.WriteFile(filePath, fileContent, 0644)
				if err != nil {
					fmt.Fprintf(w, "无法保存文件 %s：%s\n", file.Name(), err)
				}
			}
		}
		content.WriteString("</ul>")

		// 设置响应头和内容
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Write([]byte(content.String()))
	})

	// 启动服务器
	fmt.Println("服务器正在运行...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("服务器启动失败：", err)
	}
}
