package main

import (
	"context"
	"github.com/cylScripter/chest/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "106.53.50.226:9000"
	accessKeyID := "9leF4JaT6BnNKxKCNHrB"
	secretAccessKey := "djMZ05R736EEiUo7uJ7FgzrEVSSUJCu4LAmzGYoa"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Errorf("Error creating minio client: %v", err)
		return
	}

	log.Infof("%#v", minioClient) // minioClient is now set up
	// Make a new bucket called testbucket.
	bucketName := "test"

	ctx := context.Background()

	location, err := minioClient.GetBucketLocation(ctx, bucketName)
	if err != nil {
		log.Errorf("Error getting bucket location: %v", err)
		return
	}
	log.Infof("bucket location: %s", location)

	// Upload the test file
	// Change the value of filePath if the file is in another location
	objectName := "index/testdata"
	filePath := "/Users/cyl/project/github.com/cylScripter/chest/minio/test.txt"
	contentType := "application/octet-stream"

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Infof("err %v", err)
	}

	log.Infof("Successfully uploaded %s of size %d\n", objectName, info.Size)

}
