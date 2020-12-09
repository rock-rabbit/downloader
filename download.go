package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// Options 下载参数
type Options struct {
	OutputPath string // 保存路径
	OutputName string // 保存文件名 为空：自动生成

	Replace bool // 是否允许覆盖文件

	Client       *http.Client
	RquestMethod string      // 请求方式 默认 GET
	RquestBody   io.Reader   // 请求Body
	RquestHeader http.Header // 头文件

	ThreadNum int // 下载线程数
}

// OnOver 文件下载完成事件
type OnOver func(dl *Downloader)

// OnRequest 设置
type OnRequest func(*http.Request)

// OnDefer 完成事件
type OnDefer func(dl *Downloader)

// Downloader 下载信息
type Downloader struct {
	URL     string                 // 下载URL
	Options *Options               // 下载参数
	IsBar   bool                   // 是否显示进度条
	Bar     pb.ProgressBarTemplate // 进度条
	OnOver  OnOver                 // 下载完成事件，没有去掉.download
	Defer   []OnDefer
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
		IsBar:   true,
		Bar:     pb.Full,
		OnOver:  func(dl *Downloader) {},
		Defer:   []OnDefer{},
	}
}

// NewOptions 创建下载参数
func NewOptions() *Options {
	return &Options{
		OutputPath: "./",
		OutputName: "",
		Replace:    true,
		Client: &http.Client{
			Timeout: time.Second * 500,
		},
		RquestBody:   nil,
		RquestHeader: http.Header{},

		ThreadNum: 1,
	}
}

// AddDfer 添加Defer
func (dl *Downloader) AddDfer(d OnDefer) *Downloader {
	dl.Defer = append(dl.Defer, d)
	return dl
}

// SetOutputPath 设置输出目录
func (dl *Downloader) SetOutputPath(outputPath string) *Downloader {
	if outputPath == "" {
		outputPath = "./"
	}
	dl.Options.OutputPath = outputPath
	return dl
}

// SetOutputName 设置输出文件名称
func (dl *Downloader) SetOutputName(outputName string) *Downloader {
	dl.Options.OutputName = outputName
	return dl
}

// SetOnOver 设置OnOver
func (dl *Downloader) SetOnOver(o OnOver) *Downloader {
	dl.OnOver = o
	return dl
}

// SetIsBar 设置是否打印进度条
func (dl *Downloader) SetIsBar(i bool) *Downloader {
	dl.IsBar = i
	return dl
}

// SetThreadNum 设置下载线程数
func (dl *Downloader) SetThreadNum(n int) *Downloader {
	if n <= 1 {
		dl.Options.ThreadNum = 1
	} else {
		dl.Options.ThreadNum = n
	}
	return dl
}

// GetPath 获取文件完整路径
func (dl *Downloader) GetPath() string {
	return path.Join(dl.Options.OutputPath, dl.Options.OutputName)
}

// GetTempName 获取临时文件名
func (dl *Downloader) GetTempName() string {
	return fmt.Sprintf("%s.download", dl.Options.OutputName)
}

// GetTempPath 获取临时文件完整路径
func (dl *Downloader) GetTempPath() string {
	return path.Join(dl.Options.OutputPath, dl.GetTempName())
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
	dl.OnOver(dl)
	os.Remove(dl.GetPath())
	if err := os.Rename(dl.GetTempPath(), dl.GetPath()); err != nil {
		return err
	}
	return nil
}

// Response 创建Response
func (dl *Downloader) Response(onRequest OnRequest) (*http.Response, error) {
	request, err := http.NewRequest(dl.Options.RquestMethod, dl.URL, dl.Options.RquestBody)
	if err != nil {
		return nil, err
	}
	request.Header = dl.Options.RquestHeader
	onRequest(request)
	resp, err := dl.Options.Client.Do(request)
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
		var reader io.Reader = resp.Body
		if dl.IsBar {
			reader = BarThreadOne(dl, tempFileSize, size, resp.Body)
		}
		if _, err := io.Copy(f, reader); err != nil {
			return err
		}
		f.Close()
		return dl.DownloadOver()
	}
	f, err := os.OpenFile(dl.GetTempPath(), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	var reader io.Reader = resp.Body
	if dl.IsBar {
		reader = BarThreadOne(dl, 0, size, resp.Body)
	}
	if _, err := io.Copy(f, reader); err != nil {
		return err
	}
	f.Close()
	return dl.DownloadOver()
}

// Defer 下载关闭事件
func Defer(dl *Downloader) {
	for _, i := range dl.Defer {
		i(dl)
	}
}
