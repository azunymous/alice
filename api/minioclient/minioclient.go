package minioclient

import (
	"errors"
	"github.com/minio/minio-go/v6"
	"io"
	"log"
	"time"
)

type MinioClient struct {
	client        *minio.Client
	defaultBucket string
}

func (m MinioClient) DefaultBucket() string {
	return m.defaultBucket
}

func ConnectToMinioWithTimeout(addr, defaultBucket, accessKey, secretAccessKey string) (client MinioClient, err error) {
	done := make(chan bool, 1)
	returnClient := make(chan MinioClient, 1)
	returnError := make(chan error, 1)
	go func() {
		minioClient, err := ConnectToMinio(addr, defaultBucket, accessKey, secretAccessKey)
		returnClient <- minioClient
		returnError <- err
		done <- true
	}()

	select {
	case <-done:
		return <-returnClient, <-returnError
	case <-time.After(10 * time.Second):
		return MinioClient{}, errors.New("minio client connectivity timeout after 10 seconds")
	}
}

func ConnectToMinio(addr, defaultBucket, accessKey, secretAccessKey string) (client MinioClient, err error) {
	minioClient, err := minio.New(addr, accessKey, secretAccessKey, false)
	if err != nil {
		return MinioClient{}, err
	}

	mc := MinioClient{client: minioClient, defaultBucket: defaultBucket}
	mc.createBucket(defaultBucket)
	err = mc.client.SetBucketPolicy(defaultBucket, "download")
	if err != nil {
		log.Printf("Error making default bucket %s available as download host: %v", defaultBucket, err)
		return MinioClient{}, err
	}
	return mc, nil
}

func (m MinioClient) Store(obj io.Reader, bucket, name string, size int64) (string, error) {
	_, err := m.client.PutObject(bucket, name, obj, size, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}

	panic("implement me")
}

func (m MinioClient) createBucket(name string) {
	exists, _ := m.client.BucketExists(name)
	if exists {
		return
	}
	err := m.client.MakeBucket(name, "")
	if err != nil {
		log.Printf("Error making bucket %s\n", err.Error())
	} else {
		log.Printf("Successfully created %s\n", name)
	}
}
