package backup

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-pkgz/repeater"
	"github.com/go-pkgz/repeater/strategy"
	"github.com/mkorenkov/covid-19/pkg/documents"
	"github.com/mkorenkov/covid-19/pkg/requestcontext"
	"github.com/pkg/errors"
)

const (
	repeaterFactor = 1.5
	repeatTimes    = 5
)

// S3Config describes most common configuration options for S3-like storage.
type S3Config interface {
	GetBucket() string
	GetEndpoint() string
	GetRegion() string
	GetAccessKey() string
	GetSecret() string
}

func key(doc documents.CollectionEntry) string {
	name := strings.TrimSpace(doc.GetName())
	if name == "" {
		return ""
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ". ", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")
	keyFormat := fmt.Sprintf("covid19/%v/2006-01-02/15/2006-01-02T15:04:05Z07:00.json", name)
	return doc.GetWhen().Format(keyFormat)
}

// Upload uploads Country / State information to S3 using the client.
func Upload(ctx context.Context, s3Client *s3.S3, bucket string, doc documents.CollectionEntry) error {
	docKey := key(doc)
	if docKey == "" {
		return errors.New("Failed to create doc key")
	}

	var b bytes.Buffer // Buffer needs no initialization.
	if err := doc.Save(&b); err != nil {
		return errors.Wrap(err, "Failed to save document")
	}

	f := func() error {
		_, err := s3Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(b.Bytes()),
			Bucket: aws.String(bucket),
			Key:    aws.String(docKey),
		})
		return err
	}

	r := repeater.New(&strategy.Backoff{
		Repeats: repeatTimes,
		Factor:  repeaterFactor,
		Jitter:  true,
	})
	if err := r.Do(ctx, f); err != nil {
		return errors.Wrap(err, "Repeater tried hard, but no could not upload to S3")
	}
	return nil
}

// ToS3 streams documents to S3 for backup reasons.
func ToS3(ctx context.Context, config S3Config, docs <-chan documents.CollectionEntry) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.GetAccessKey(), config.GetSecret(), ""),

		Region:           aws.String(config.GetRegion()),
		S3ForcePathStyle: aws.Bool(true),
	}
	if config.GetEndpoint() != "" {
		s3Config.Endpoint = aws.String(config.GetEndpoint())
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	errorChan := requestcontext.Errors(ctx)
	if errorChan == nil {
		panic(errors.New("Could not retrieve error chan from context"))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case doc := <-docs:
			if err := Upload(ctx, s3Client, config.GetBucket(), doc); err != nil {
				errorChan <- errors.Wrap(err, "Failed writing to S3")
			}
		}
	}
}
