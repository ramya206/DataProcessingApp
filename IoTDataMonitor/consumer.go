package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"os"
	"sync"
	"time"
)



//func main() {
//	stream := flag.String("stream", "RealTimeDataStream", "Stream Name")
//	flag.Parse()

//	sess := session.Must(
//		session.NewSessionWithOptions(
//			session.Options{
//				SharedConfigState: session.SharedConfigEnable,
//			},
//		),
//	)

//	pollShards(kinesis.New(sess), stream)
//}

func pollShards(client *kinesis.Kinesis, stream *string) {
	var wg sync.WaitGroup

	streamDescription, err := client.DescribeStream(
		&kinesis.DescribeStreamInput{
			StreamName: stream,
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, shard := range streamDescription.StreamDescription.Shards {
		go getRecords(client, stream, shard.ShardId)
		wg.Add(1)
	}

	wg.Wait()
}

func getRecords(client *kinesis.Kinesis, stream *string, shardID *string) {
	shardIteratorRes, err := client.GetShardIterator(
		&kinesis.GetShardIteratorInput{
			StreamName:        stream,
			ShardId:           shardID,
			ShardIteratorType: aws.String("LATEST"),
		},
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	shardIterator := shardIteratorRes.ShardIterator
	ticker := time.NewTicker(time.Second)


	for range ticker.C {
		records, err := client.GetRecords(
			&kinesis.GetRecordsInput{
				ShardIterator: shardIterator,
			},
		)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, record := range records.Records {
			var prettyJSON bytes.Buffer

			//fmt.Print(string(record.Data))

			if err := json.Indent(&prettyJSON, record.Data, "", "    "); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}


			var dataFromStream SensorData
			err := json.Unmarshal(record.Data,&dataFromStream)
			if err!=nil {
				panic(err)
				fmt.Print("getRecords: Error in Json unmarshal ",err)
			}
			StreamToSocket := StreamToSocket{
				Type:"RealTime",
				Data: dataFromStream,
			}

			pool.broadcast<-StreamToSocket
			fmt.Print(string(prettyJSON.Bytes()))
		}

		shardIterator = records.NextShardIterator
	}
}
