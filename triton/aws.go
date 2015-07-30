package triton

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// These 'Services' are just shims to allow for easier testing.  They also give
// us an idea of the minimum functionality we're using from our APIs.

type KinesisService interface {
	DescribeStream(*kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error)
	GetShardIterator(*kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)
	GetRecords(*kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)
}

type S3Service interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

type S3UploaderService interface {
	Upload(input *s3manager.UploadInput) (*s3manager.UploadOutput, error)
}

type DynamoDBService interface {
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
}

type nullS3Service struct{}

func (s *nullS3Service) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	goo := &s3.GetObjectOutput{}
	return goo, nil
}
