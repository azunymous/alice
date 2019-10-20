package minioclient

import (
	"errors"
	"github.com/alice-ws/alice/data"
	"github.com/minio/minio-go/v6"
	"io"
	"log"
	"mime"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type MinioClient struct {
	client        *minio.Client
	defaultBucket string
}

func (m MinioClient) DefaultBucket() string {
	return m.defaultBucket
}

func ConnectToMinioWithTimeout(addr, defaultBucket, accessKey, secretAccessKey string, timeout time.Duration) (client data.MediaRepo, err error) {
	done := make(chan bool, 1)
	returnClient := make(chan data.MediaRepo, 1)
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
	case <-time.After(timeout):
		return MinioClient{}, errors.New("minio client connectivity timeout after " + timeout.String() + " seconds")
	}
}

func ConnectToMinio(addr, defaultBucket, accessKey, secretAccessKey string) (client data.MediaRepo, err error) {
	minioClient, err := minio.New(addr, accessKey, secretAccessKey, false)
	if err != nil {
		return MinioClient{}, err
	}

	mc := MinioClient{client: minioClient, defaultBucket: defaultBucket}
	mc.createBucket(defaultBucket)
	policy := ` {"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetBucketLocation","s3:ListBucket"],"Resource":["arn:aws:s3:::images"]},{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::images/*"]}]}`
	err = mc.client.SetBucketPolicy(defaultBucket, policy)
	setPolicy, _ := mc.client.GetBucketPolicy(defaultBucket)
	log.Printf("Current %s policy: %v", defaultBucket, setPolicy)
	if err != nil {
		log.Printf("Error making default bucket %s available as download host: %v", defaultBucket, err)
		return MinioClient{}, err
	}
	return mc, nil
}

func (m MinioClient) Store(obj io.Reader, bucket, name string, size int64) (string, error) {

	_, err := m.client.PutObject(bucket, name, obj, size, minio.PutObjectOptions{ContentType: mime.TypeByExtension(filepath.Ext(name))})
	if err != nil {
		return "", err
	}

	log.Printf("Storing image %s in bucket %s", name, bucket)
	return bucket + "/" + name, nil
}

func (m MinioClient) GenerateUniqueName(fileName string) string {
	ext := path.Ext(fileName)
	return strconv.FormatInt(time.Now().UnixNano(), 10) + ext
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
