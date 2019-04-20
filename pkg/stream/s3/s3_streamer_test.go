package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ofHeadObjectInput struct {
	Bucket string
	Key    string
}

func OfHeadObjectInput(bucket string, key string) gomock.Matcher {
	return &ofHeadObjectInput{Bucket: bucket, Key: key}
}

func (o *ofHeadObjectInput) Matches(x interface{}) bool {
	that, ok := x.(*s3.HeadObjectInput)
	if !ok {
		return false
	}
	return *that.Bucket == o.Bucket && *that.Key == o.Key
}

func (o *ofHeadObjectInput) String() string {
	return "Bucket: " + o.Bucket + " Key: " + o.Key
}

func TestS3Streamer_WithS3Url_CanStreamTrue(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "s3://test/stream"
	streamer := NewStreamer(nil)
	assert.True(t, streamer.CanStream(filePath))
}

func TestS3Streamer_WithS3Url_CanStreamFalse(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	filePath := "https://www.file.com"
	streamer := NewStreamer(nil)
	assert.False(t, streamer.CanStream(filePath))
}

func TestPathToStream_WithS3Url_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	url := "s3://test/stream"
	err := fmt.Errorf("test error")

	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(gomock.Any()).Return(nil, err)

	streamer := NewStreamer(mockS3Client)
	stream, _, _, err := streamer.CreateStream(url)

	assert.Nil(t, stream)
	assert.NotNil(t, err)
}

func TestPathToStream_WithValidS3URL_ShouldReturnValidStream(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// s3etag is wrapped in quotes
	s3Etag := "\"abcdefg\""

	// we expect our final etag to have quotes stripped
	expectedEtag := "\"abcdefg\""

	var expectedContentLength int64 = 12345

	expectedHeadObjectOutput := &s3.HeadObjectOutput{
		ETag:          &s3Etag,
		ContentLength: &expectedContentLength,
	}

	mockS3Client := NewMockS3API(ctrl)
	mockS3Client.EXPECT().HeadObject(OfHeadObjectInput("test", "stream")).Return(expectedHeadObjectOutput, nil).Times(2)

	url := "s3://test/stream"

	streamer := NewStreamer(mockS3Client)
	stream, _, _, err := streamer.CreateStream(url)

	stream, etag, contentLength, err := streamer.CreateStream(url)

	assert.NotNil(t, stream)
	assert.Equal(t, expectedEtag, etag)
	assert.Equal(t, expectedContentLength, contentLength)
	assert.Nil(t, err)
}

func TestPathToStream_WithS3UrlWithoutClientAndWithoutRegion_ShouldReturnUnsupportedError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	url := "s3://test/stream"
	err := fmt.Errorf("test error")

	streamer := NewStreamer(nil)
	stream, _, _, err := streamer.CreateStream(url)

	stream, md5, _, err := streamer.CreateStream(url)

	assert.Nil(t, stream)
	assert.Equal(t, "", md5)
	assert.NotNil(t, err)
}
