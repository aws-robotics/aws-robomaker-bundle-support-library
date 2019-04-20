package bundle

import (
	"io"
	"time"
)

type proxyReadSeeker struct {
	r                     io.ReadSeeker
	contentLength         int64
	readStartTime         time.Time
	lastUpdated           time.Time
	callback              ProgressCallback
	callbackRateInSeconds int
}

func (r *proxyReadSeeker) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)

	currentPos, _ := r.Seek(0, io.SeekCurrent)
	percentDone := float32((float64(currentPos) / float64(r.contentLength)) * 100)

	if time.Since(r.lastUpdated).Seconds() > float64(r.callbackRateInSeconds) || currentPos == r.contentLength {
		r.callback(percentDone, time.Since(r.readStartTime))
		r.lastUpdated = time.Now()
	}

	return
}

func (r *proxyReadSeeker) Seek(offset int64, whence int) (newOffset int64, err error) {
	if whence == io.SeekEnd {
		r.callback(100.0, time.Since(r.readStartTime))
	}

	return r.r.Seek(offset, whence)
}
