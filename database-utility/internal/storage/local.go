package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func SaveLocal(dir, name string, r io.Reader) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return path, nil
}

func UploadToS3Placeholder(bucket, key, path string) error {
	// placeholder - real implementation will use AWS SDK
	fmt.Printf("Pretend uploading %s to s3://%s/%s\n", path, bucket, key)
	return nil
}
