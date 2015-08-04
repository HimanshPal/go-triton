package triton

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	//"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// Some types to make sure our lists of func args don't get confused
type ShardID string

type SequenceNumber string

// A ShardStreamReader provides records from a Kinesis stream.
// It's specific to a single shard. A Stream is blocking, and will avoid
// overloading a shard by limiting how often it attempts to consume records.
type ShardStreamReader struct {
	StreamName         string
	ShardID            ShardID
	ShardIteratorType  string
	NextIteratorValue  *string
	LastSequenceNumber *SequenceNumber

	service     KinesisService
	records     []*kinesis.Record
	lastRequest *time.Time
}

// Recommended minimum polling interval to keep from overloading a Kinesis
// shard.
const MIN_POLL_INTERVAL = 1.0 * time.Second

func (s *ShardStreamReader) initIterator() (err error) {
	gsi := kinesis.GetShardIteratorInput{
		StreamName:        aws.String(s.StreamName),
		ShardID:           aws.String(string(s.ShardID)),
		ShardIteratorType: aws.String(s.ShardIteratorType),
	}

	if s.LastSequenceNumber != nil {
		gsi.StartingSequenceNumber = aws.String(string(*s.LastSequenceNumber))
	}

	gso, err := s.service.GetShardIterator(&gsi)
	if err != nil {
		return err
	}

	s.NextIteratorValue = gso.ShardIterator
	return nil
}

func (s *ShardStreamReader) wait(minInterval time.Duration) {
	if s.lastRequest != nil {
		sleepTime := minInterval - time.Since(*s.lastRequest)
		if sleepTime >= time.Duration(0) {
			time.Sleep(sleepTime)
		}
	}

	n := time.Now()
	s.lastRequest = &n
}

func (s *ShardStreamReader) fetchMoreRecords() (err error) {
	s.wait(MIN_POLL_INTERVAL)

	if s.NextIteratorValue == nil {
		err := s.initIterator()
		if err != nil {
			return err
		}
	}

	gri := &kinesis.GetRecordsInput{
		Limit:         aws.Int64(1000),
		ShardIterator: s.NextIteratorValue,
	}

	gro, err := s.service.GetRecords(gri)
	if err != nil {
		return err
	}

	s.records = gro.Records
	s.NextIteratorValue = gro.NextShardIterator

	return nil
}

func (s *ShardStreamReader) Get() (r *kinesis.Record, err error) {
	for {
		if len(s.records) > 0 {
			r := s.records[0]
			s.records = s.records[1:]
			return r, nil
		} else {
			err := s.fetchMoreRecords()
			if err != nil {
				return nil, err
			}
		}
	}
}

// Create a new stream given a specific sequence number
//
// This uses the Kinesis AFTER_SEQUENCE_NUMBER interator type, so this assumes
// the provided sequenceNumber has already been processed, and the caller wants
// records produced since.
func NewShardStreamReaderFromSequence(svc KinesisService, streamName string, sid ShardID, sn SequenceNumber) (s *ShardStreamReader) {
	s = &ShardStreamReader{
		StreamName:         streamName,
		ShardID:            sid,
		ShardIteratorType:  "AFTER_SEQUENCE_NUMBER",
		LastSequenceNumber: &sn,
		service:            svc,
	}

	return s
}

// Create a new stream starting at the latest position
//
// This uses the Kinesis LATEST iterator type and assumes the caller only wants new data.
func NewShardStreamReader(svc KinesisService, streamName string, sid ShardID) (s *ShardStreamReader) {
	s = &ShardStreamReader{
		StreamName:        streamName,
		ShardID:           sid,
		ShardIteratorType: "LATEST",
		service:           svc,
	}

	return s
}

// Utility function to pick a shard id given an integer shard number.
// Use this if you want the 2nd shard, but don't know what the id would be.
func PickShardID(svc KinesisService, streamName string, shardNum int) (sid ShardID, err error) {
	resp, err := svc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return "", fmt.Errorf("Failed to find stream %v", awsErr.Message())
			}
		}

		return
	}

	if len(resp.StreamDescription.Shards) < shardNum {
		err = fmt.Errorf("Stream doesn't have a shard %d", shardNum)
		return
	}

	sid = ShardID(*resp.StreamDescription.Shards[shardNum].ShardID)
	return
}

func ListShards(svc KinesisService, streamName string) (shards []string, err error) {
	resp, err := svc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(streamName)})
	if err != nil {
		return
	}

	for _, s := range resp.StreamDescription.Shards {
		shards = append(shards, *s.ShardID)
	}

	return
}
