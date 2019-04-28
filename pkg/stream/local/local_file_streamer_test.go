package local

//go:generate mockgen -destination=mock_file_system.go -package=local github.com/spf13/afero File
//go:generate mockgen -destination=mock_file.go -package=local github.com/spf13/afero Fs
//go:generate mockgen -destination=mock_file_info.go -package=local os FileInfo

import (
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestLocalFileStreamer_WithHttpsUrl_ShouldReturnFalse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	streamer := newStreamer(NewMockFs(ctrl))
	assert.False(t, streamer.CanStream("https://www.google.com"))
}

func TestLocalFileStreamer_WithFileUrl_ShouldReturnTrue(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	streamer := newStreamer(NewMockFs(ctrl))
	assert.True(t, streamer.CanStream("file:///my/file"))
}

func TestLocalFileStreamer_WithFilePath_ShouldReturnTrue(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	streamer := newStreamer(NewMockFs(ctrl))
	assert.True(t, streamer.CanStream("/this/is/a/path"))
}

func TestLocalFileStreamer_WithHttpsUrl_ShouldReturnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	streamer := newStreamer(NewMockFs(ctrl))
	_, _, _, err := streamer.CreateStream("https://www.google.com")
	assert.Error(t, err)
}

func TestLocalFileStreamer_ShouldReturnOpenError_OnMd5Open(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "/test/stream"

	mockFileSystem := NewMockFs(ctrl)
	mockFileSystem.EXPECT().Open(filePath).Return(nil, io.ErrUnexpectedEOF)

	streamer := newStreamer(mockFileSystem)
	_, _, _, err := streamer.CreateStream(filePath)

	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestLocalFileStreamer_ShouldReturnOpenError_OnStreamOpen(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	contents := []byte("12345")
	filePath := "/test/stream"
	shouldError := false

	mockFile := NewMockFile(ctrl)
	mockFileSystem := NewMockFs(ctrl)
	mockFileSystem.EXPECT().Open(filePath).DoAndReturn(func(_ interface{}) (afero.File, error) {
		if shouldError {
			return nil, os.ErrInvalid
		}
		shouldError = true
		return mockFile, nil
	}).Times(2)
	mockFile.EXPECT().Read(gomock.Any()).Do(func(ibuf interface{}) {
		buf := ibuf.([]byte)
		copy(buf, contents)
	}).Return(len(contents), io.EOF).Times(1)
	mockFile.EXPECT().Close().Times(1)

	streamer := newStreamer(mockFileSystem)
	_, _, _, err := streamer.CreateStream(filePath)

	assert.Equal(t, os.ErrInvalid, err)
}

func TestLocalFileStreamer_ShouldReturnStatError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	contents := []byte("12345")
	filePath := "/test/stream"

	mockFile := NewMockFile(ctrl)
	mockFileSystem := NewMockFs(ctrl)
	mockFileInfo := NewMockFileInfo(ctrl)
	mockFileSystem.EXPECT().Open(filePath).Return(mockFile, nil).Times(2)
	mockFile.EXPECT().Stat().Return(mockFileInfo, os.ErrPermission).Times(1)
	mockFile.EXPECT().Read(gomock.Any()).Do(func(ibuf interface{}) {
		buf := ibuf.([]byte)
		copy(buf, contents)
	}).Return(len(contents), io.EOF).Times(1)
	mockFile.EXPECT().Close().Times(1)

	streamer := newStreamer(mockFileSystem)
	_, _, _, err := streamer.CreateStream(filePath)

	assert.Equal(t, os.ErrPermission, err)
}

func TestLocalFileStreamer_WithLocalFile_ShouldReturnStreamAndMd5(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	contents := []byte("12345")
	filePath := "/test/stream"

	mockFile := NewMockFile(ctrl)
	mockFileSystem := NewMockFs(ctrl)
	mockFileInfo := NewMockFileInfo(ctrl)
	mockFileSystem.EXPECT().Open(filePath).Return(mockFile, nil).Times(2)
	mockFileInfo.EXPECT().Size().Return(int64(len(contents))).Times(1)
	mockFile.EXPECT().Stat().Return(mockFileInfo, nil).Times(1)
	mockFile.EXPECT().Read(gomock.Any()).Do(func(ibuf interface{}) {
		buf := ibuf.([]byte)
		copy(buf, contents)
	}).Return(len(contents), io.EOF).Times(1)
	mockFile.EXPECT().Close().Times(1)

	streamer := newStreamer(mockFileSystem)
	stream, contentLength, md5, err := streamer.CreateStream(filePath)

	assert.Equal(t, mockFile, stream)
	assert.Equal(t, "827ccb0eea8a706c4c34a16891f84e7b", md5)
	assert.Equal(t, int64(len(contents)), contentLength)
	assert.Nil(t, err)
}
