package downloader

import (
	"fmt"
	"testing"
)

func TestDownload(t *testing.T) {
	// 测试下载
	url := "https://dev46.baidupan.com/121116bb/2020/12/10/a7ae9c58120be493375988f8e02475d4.7z?st=sE7sMJalY78rlP59XLuPrQ&e=1607675694&b=U1FdMgB0UXlZXF5kVWQDflZmCCYNOgEzVloBPVB5X2pTL15sVWQEMgVpUjEHCAdTVnYINlQ5AW4IOApYV2JUZlMzXW0AMVFmWTxeMVUrAzBWeQ_c_c&fi=34054584&pid=39-70-47-126&up="
	outPath := `C:\Users\Administrator\Desktop\新建文件夹`
	if err := New(url, outPath).SetIsBar(false).SetOnProgress(func(size, speed, downloadedSize int64) {
		fmt.Printf("\r总大小：%d byte  已下载：%d byte  下载速度：%d byte/s", size, downloadedSize, speed)
	}).AddDfer(func(dl *Downloader) {
		fmt.Printf("  下载结束\n")
	}).SetOutputName(fmt.Sprintf("Pot-Player64.7z")).Start(); err != nil {
		t.Log(err)
	}
}
