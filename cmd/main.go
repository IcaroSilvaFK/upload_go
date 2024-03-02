package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Client *s3.S3
	s3Bucket string

	wg sync.WaitGroup
)

func init() {
	s, err := session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				"id",
				"password",
				"token",
			),
		},
	)

	if err != nil {
		panic(err)
	}

	s3Client = s3.New(s)
	s3Bucket = "bucket"

}

func main() {

	dir, err := os.Open("./tmp")

	uploadControl := make(chan struct{}, 100) // struct is a zero bytes
	errorFileUpload := make(chan string, 10)

	if err != nil {
		panic(err)
	}

	defer dir.Close()

	go func() {

		for {
			select {
			case filename := <-errorFileUpload:
				{

					uploadControl <- struct{}{}
					wg.Add(1)
					go uploadFile(filename, uploadControl, errorFileUpload)
				}
			}
		}

	}()

	for {
		files, err := dir.Readdir(1)

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}
		wg.Add(1)
		uploadControl <- struct{}{}

		go uploadFile(files[0].Name(), uploadControl, errorFileUpload)
	}
	wg.Wait()

}

func uploadFile(filename string, uploadControl <-chan struct{}, errorFileUpload chan<- string) {

	defer wg.Done()

	completeFileName := fmt.Sprintf("%s/%s", "./tmp", filename)

	file, err := os.Open(completeFileName)

	if err != nil {
		fmt.Printf("Error opening file: %s, %s", err, filename)
		<-uploadControl // unblocking
		errorFileUpload <- filename
		return
	}

	defer file.Close()

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		fmt.Printf("Error uploading file: %s, %s", err, filename)
		errorFileUpload <- filename
		<-uploadControl // unblocking
		return
	}

	fmt.Printf("File uploaded: %s\n", filename)
	<-uploadControl // unblocking
}
