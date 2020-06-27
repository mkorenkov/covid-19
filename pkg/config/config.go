package config

import (
	"path"
	"time"
)

type Config struct {
	S3Secret       string            `split_words:"true" required:"true"`
	S3AccessKey    string            `split_words:"true" required:"true"`
	S3Region       string            `split_words:"true" default:"us-east-1"`
	S3Endpoint     string            `split_words:"true"`
	S3Bucket       string            `split_words:"true" required:"true"`
	StorageDir     string            `split_words:"true" required:"true"`
	ScrapeInterval time.Duration     `split_words:"true" required:"true"`
	ListenAddr     string            `split_words:"true" required:"true"`
	Credentials    map[string]string `split_words:"true" required:"true"` // comma separated user:password pairs
	SentryDSN      string            `split_words:"true" required:"true"`
}

// ImportsDir where to store the imports.
func (c Config) ImportsDir() string {
	return path.Join(c.StorageDir, "imports")
}

// GetBucket returns S3 bucket.
func (c Config) GetBucket() string {
	return c.S3Bucket
}

// GetEndpoint (optinal) returns S3 endpoint.
func (c Config) GetEndpoint() string {
	return c.S3Endpoint
}

// GetRegion returns S3 region.
func (c Config) GetRegion() string {
	return c.S3Region
}

// GetAccessKey returns AWS access keys.
func (c Config) GetAccessKey() string {
	return c.S3AccessKey
}

// GetSecret returns AWS secret.
func (c Config) GetSecret() string {
	return c.S3Secret
}
