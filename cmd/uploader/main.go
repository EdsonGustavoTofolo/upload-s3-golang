package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"sync"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg       sync.WaitGroup
)

const (
	awsId     = "-"
	awsSecret = "-"
	awsToken  = "-"
)

func init() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			awsId,
			awsSecret,
			awsToken,
		),
	})
	if err != nil {
		panic(err)
	}

	s3Client = s3.New(sess)
	s3Bucket = "goexpert-bucket"
}

func main() {
	dir, err := os.Open("./temp")

	if err != nil {
		panic(err)
	}

	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			panic(err)
		}
	}(dir)

	uploadControl := make(chan struct{}, 100)
	errorFileUpload := make(chan string, 10)

	go func() {
		for {
			select {
			case fileName := <-errorFileUpload:
				uploadControl <- struct{}{}
				wg.Add(1)
				go uploadFile(fileName, uploadControl, errorFileUpload)
			}
		}
	}()

	for {
		files, err := dir.ReadDir(1)

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading %v\n", err)
			continue
		}

		wg.Add(1)
		uploadControl <- struct{}{}

		go uploadFile(files[0].Name(), uploadControl, errorFileUpload)
	}
	wg.Wait()
}

func uploadFile(fileName string, uploadControl <-chan struct{}, errorFileUpload chan<- string) {
	defer wg.Done()

	completeFileName := fmt.Sprintf("./temp/%s", fileName)

	fmt.Printf("Uploading file %s\n", completeFileName)

	file, err := os.Open(completeFileName)

	if err != nil {
		fmt.Printf("Error opening file: %s - %v\n", completeFileName, err)
		<-uploadControl
		errorFileUpload <- completeFileName
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})

	if err != nil {
		fmt.Printf("Error uploading file: %s - %v\n", completeFileName, err)
		<-uploadControl
		errorFileUpload <- completeFileName
		return
	}

	fmt.Printf("File %s uploaded successfully\n", completeFileName)
	<-uploadControl
}
