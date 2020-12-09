package downloader

import "io"

// IoProxyReader 代理io读
type IoProxyReader struct {
	io.Reader
	dl *Downloader
}

// Read 读
func (r *IoProxyReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.dl.Info.AddDownloadedSize(int64(n))
	return n, err
}

// Close the wrapped reader when it implements io.Closer
func (r *IoProxyReader) Close() (err error) {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return
}

// ProxyReader 代理io读
func (dl *Downloader) ProxyReader(r io.Reader) io.Reader {
	return &IoProxyReader{r, dl}
}
