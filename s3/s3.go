package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/digitalmonsters/go-common/boilerplate"
	"time"
)

type Uploader struct {
	config   *boilerplate.S3Config
	session  *session.Session
	s3Client *s3.S3
}

func NewUploader(cfg *boilerplate.S3Config) *Uploader {
	u := &Uploader{
		config: cfg,
	}
	return u
}

func (u *Uploader) GetObjectSignedUrl(path string, urlExpiration time.Duration) (string, error) {
	client, err := u.getClient()
	if err != nil {
		return "", err
	}

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(path),
	})

	signedUrl, err := req.Presign(urlExpiration)
	if err != nil {
		return "", err
	}

	return signedUrl, nil
}

func (u *Uploader) PutObjectSignedUrl(path string, urlExpiration time.Duration) (string, error) {
	client, err := u.getClient()
	if err != nil {
		return "", err
	}

	req, _ := client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(path),
	})

	signedUrl, err := req.Presign(urlExpiration)
	if err != nil {
		return "", err
	}

	return signedUrl, nil
}

func (u *Uploader) UploadObject(path string, data []byte, contentType string) error {
	client, err := u.getClient()
	if err != nil {
		return err
	}
	_, err = client.PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(data),
		Key:         aws.String(path),
		Bucket:      aws.String(u.config.Bucket),
		ContentType: aws.String(contentType),
	})
	return err
}

func (u *Uploader) getClient() (*s3.S3, error) {
	if u.session == nil {
		if sess, err := session.NewSession(&aws.Config{Region: aws.String(u.config.Region)}); err != nil {
			return nil, err
		} else {
			u.session = sess
		}
	}

	if u.s3Client == nil {
		u.s3Client = s3.New(u.session)
	}
	return u.s3Client, nil
}
