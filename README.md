## downloader
---
## ✨ 简介
一个基于go语言的http下载器
## 🎉 功能
- 文件夹自动创建
- 命令行进度条
- 文件名过滤特殊字符
- 断点续传
- 单线程下载
## 🐱‍🏍 计划
- 多线程下载
## 🎊 安装
```
go get -u gitee.com/rock_rabbit/downloader
```
## 🎠 使用

```go
package main
import "gitee.com/rock_rabbit/downloader"
func main(){
    url := "https://desk-fd.zol-img.com.cn/t_s960x600c5/g3/M05/0B/0E/ChMlWF7xvaWIcxEzAB9uRBF9dyoAAVJlgLjyx8AH25c171.jpg"
    err := downloader.New(url,"./").Start()
    if err != nil{
        panic(err)
    }
}
```