package downloader

import (
	"testing"
	"time"
)

func TestDownload(t *testing.T) {
	// 测试下载
	if err := New("https://gw.alipayobjects.com/mdn/prod_resou/afts/file/A*oIndSLbrp0kAAAAAAAAAAABjARQnAQ", `C:\Users\Administrator\Desktop\新建文件夹`).Start(); err != nil {
		t.Log(err)
	}
}

func TestDownloadedSize(t *testing.T) {
	// 测试外部获取进度
	dl := New("https://gw.alipayobjects.com/mdn/prod_resou/afts/file/A*oIndSLbrp0kAAAAAAAAAAABjARQnAQ", `C:\Users\Administrator\Desktop\新建文件夹`)
	go func() {
		for {
			time.Sleep(time.Microsecond)
			t.Log(dl.Info.GetDownloadedSize())
		}
	}()
	if err := dl.SetIsBar(false).Start(); err != nil {
		t.Log(err)
	}
}
