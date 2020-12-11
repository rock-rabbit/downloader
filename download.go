package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// Rquest 请求参数
type Rquest struct {
	Client *http.Client
	Method string      // 请求方式 默认 GET
	Body   io.Reader   // 请求Body
	Header http.Header // 头文件
}

// Options 下载参数
type Options struct {
	OutputPath string // 保存路径
	OutputName string // 保存文件名 为空：自动生成

	Replace bool   // 是否允许覆盖文件
	Rquest  Rquest // 请求参数

	ThreadNum int // 下载线程数
}

// OnRequest 设置
type OnRequest func(*http.Request)

// OnDefer 完成事件
type OnDefer func(dl *Downloader)

// OnProgress 载进度回调
// size				文件总大小，读取不出文件大小为0
// speed			每秒下载大小，用来计算下载速度/s
// downloadedSize	已下载文件大小
type OnProgress func(size, speed, downloadedSize int64)

// ProgressBar 进度条参数
type ProgressBar struct {
	IsBar       bool                   // 是否显示进度条
	BarTemplate pb.ProgressBarTemplate // 进度条样式
}

// DownloadInfo 下载时的信息
type DownloadInfo struct {
	Size           int64  // 文件总大小
	Speed          int64  // 每秒下载大小，用来计算下载速度/s
	DownloadedSize *int64 // 已下载文件大小，原子操作
}

// Downloader 下载信息
type Downloader struct {
	URL         string        // 下载URL
	Options     *Options      // 下载参数
	ProgressBar ProgressBar   // 进度条参数
	Defer       []OnDefer     // 下载关闭事件
	Info        *DownloadInfo // 下载时的信息
	OnProgress  OnProgress    // 下载进度回调
}

// New 创建一个简单的下载器
func New(url string, outputPath string) *Downloader {
	dl := NewDownloader(url).SetOutputPath(outputPath)
	return dl
}

// NewDownloader 创建下载器
func NewDownloader(url string) *Downloader {
	return &Downloader{
		URL:     url,
		Options: NewOptions(),
		ProgressBar: ProgressBar{
			IsBar:       true,
			BarTemplate: pb.Full,
		},
		Defer: []OnDefer{},
		Info: &DownloadInfo{
			DownloadedSize: new(int64),
		},
		OnProgress: func(size, speed, downloadedSize int64) {},
	}
}

// NewOptions 创建下载参数
func NewOptions() *Options {
	return &Options{
		OutputPath: "./",
		OutputName: "",
		Replace:    true,
		Rquest: Rquest{
			Client: &http.Client{
				Timeout: time.Second * 500,
			},
			Method: "GET",
			Body:   nil,
			Header: http.Header{},
		},
		ThreadNum: 1,
	}
}

// IsExist 目录与文件是否存在处理
// 自动创建目录
// 检查文件是否完整，若出现异常直接返回不完整
func (dl *Downloader) IsExist(size int64, lastModified string) error {
	var err error
	if err = os.MkdirAll(dl.Options.OutputPath, 0666); err != nil {
		return err
	}
	info, err := os.Stat(dl.GetPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !dl.Options.Replace {
		return errors.New("you are not allowed to replace files because dl.Options.Replace is false")
	}
	fileCreationTime := GetFileCreateTime(info.Sys())
	t, _ := time.Parse(time.RFC1123, lastModified)
	if size != 0 && info.Size() == size && fileCreationTime >= t.Unix() {
		return errors.New("file already exists")
	}
	return nil
}

// DownloadOver 下载完成，强制覆盖，修改文件名称
func (dl *Downloader) DownloadOver() error {
	os.Remove(dl.GetPath())
	if err := os.Rename(dl.GetTempPath(), dl.GetPath()); err != nil {
		return err
	}
	return nil
}

// Response 创建Response
func (dl *Downloader) Response(onRequest OnRequest) (*http.Response, error) {
	request, err := http.NewRequest(dl.Options.Rquest.Method, dl.URL, dl.Options.Rquest.Body)
	if err != nil {
		return nil, err
	}
	request.Header = dl.Options.Rquest.Header
	onRequest(request)
	resp, err := dl.Options.Rquest.Client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// IsRanges 判断是否支持断点续传
func (dl *Downloader) IsRanges() bool {
	resp, err := dl.Response(func(r *http.Request) {
		r.Header.Set("Range", "bytes=0-9")
	})
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.Header.Get("Accept-Ranges") != ""
}

// Start 启动下载
func (dl *Downloader) Start() error {
	if dl.Options.ThreadNum <= 1 {
		return dl.ThreadOne()
	}
	return dl.Thread()
}

// Thread 多线程下载器
func (dl *Downloader) Thread() error {
	if !dl.IsRanges() {
		// 不支持多线程下载时，自动单线程下载
		return dl.ThreadOne()
	}
	return nil
}

// ThreadOne 单线程下载器 支持断点续传
func (dl *Downloader) ThreadOne() error {
	defer Defer(dl)
	resp, err := dl.Response(func(_ *http.Request) {})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("status code not 200")
	}
	var size int64
	sizeStr := resp.Header.Get("Content-Length")
	if sizeStr != "" {
		if size, err = strconv.ParseInt(sizeStr, 10, 0); err != nil {
			return err
		}
	}
	dl.Info.Size = size
	if dl.Options.OutputName == "" {
		dl.Options.OutputName = GetFilename(dl.URL, resp.Header.Get("Content-Disposition"), resp.Header.Get("Content-Type"))
	} else {
		dl.Options.OutputName = GetFiltrationFilename(dl.Options.OutputName)
	}
	if err = dl.IsExist(size, resp.Header.Get("Last-Modified")); err != nil {
		return err
	}
	var tempFileSize int64
	tempFileInfo, err := os.Stat(dl.GetTempPath())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		tempFileSize = tempFileInfo.Size()
	}
	if size != 0 && tempFileSize == size {
		return dl.DownloadOver()
	}
	if tempFileSize != 0 && dl.IsRanges() {
		dl.Info.AddDownloadedSize(tempFileSize)
		go dl.StatSpeed() // 统计下载速度
		resp, err := dl.Response(func(req *http.Request) {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", tempFileSize))
		})
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 206 {
			return errors.New("status code not 206")
		}
		f, err := os.OpenFile(dl.GetTempPath(), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		f.Seek(tempFileSize, 0)
		reader := dl.ProxyReader(resp.Body)
		if dl.GetIsBar() {
			reader = BarThreadOne(dl, tempFileSize, size, reader)
		}
		if _, err := io.Copy(f, reader); err != nil {
			return err
		}
		f.Close()
		return dl.DownloadOver()
	}
	go dl.StatSpeed() // 统计下载速度
	f, err := os.OpenFile(dl.GetTempPath(), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := dl.ProxyReader(resp.Body)
	if dl.GetIsBar() {
		reader = BarThreadOne(dl, 0, size, reader)
	}
	if _, err := io.Copy(f, reader); err != nil {
		return err
	}
	f.Close()
	return dl.DownloadOver()
}

// StatSpeed 统计下载速度
func (dl *Downloader) StatSpeed() {
	tempSize := dl.Info.GetDownloadedSize()
	for {
		time.Sleep(time.Second)
		dl.Info.Speed = dl.Info.GetDownloadedSize() - tempSize
		tempSize = dl.Info.Speed
	}
}

// Defer 下载关闭事件
func Defer(dl *Downloader) {
	for _, i := range dl.Defer {
		i(dl)
	}
}
