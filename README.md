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
- 实时获取下载状态
## 🐱‍🏍 计划
- 多线程下载
- 限速下载
## 🎊 安装
```
go get -u gitee.com/rock_rabbit/downloader
```
## 📖 文档
```
https://godoc.org/gitee.com/rock_rabbit/downloader
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
```go
package main
import (
	"fmt"
	"gitee.com/rock_rabbit/downloader"
)
func main() {
	url := "https://dev46.baidupan.com/121116bb/2020/12/10/a7ae9c58120be493375988f8e02475d4.7z?st=sE7sMJalY78rlP59XLuPrQ&e=1607675694&b=U1FdMgB0UXlZXF5kVWQDflZmCCYNOgEzVloBPVB5X2pTL15sVWQEMgVpUjEHCAdTVnYINlQ5AW4IOApYV2JUZlMzXW0AMVFmWTxeMVUrAzBWeQ_c_c&fi=34054584&pid=39-70-47-126&up="
    outPath := `C:\Users\Administrator\Desktop\新建文件夹`
    // 以下执行流程解析
    // 设置不显示进度条，设置下载进度回调，添加下载结束defer，设置文件名称，开始
	if err := downloader.New(url, outPath).SetIsBar(false).SetOnProgress(func(size, speed, downloadedSize int64) {
		fmt.Printf("\r总大小：%d byte  已下载：%d byte  下载速度：%d byte/s", size, downloadedSize, speed)
	}).AddDfer(func(dl *downloader.Downloader) {
		fmt.Printf("  下载结束\n")
	}).SetOutputName(fmt.Sprintf("Pot-Player64.7z")).Start(); err != nil {
		panic(err)
	}
}

```