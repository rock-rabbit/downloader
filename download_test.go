package downloader

import "testing"

func TestDownload(t *testing.T) {
	// 测试下载
	if err := New("https://gw.alipayobjects.com/mdn/prod_resou/afts/file/A*oIndSLbrp0kAAAAAAAAAAABjARQnAQ", `C:\Users\Administrator\Desktop\新建文件夹`).Start(); err != nil {
		t.Log(err)
	}
}
