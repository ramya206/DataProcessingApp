package main

import (
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
		fmt.Println("shard.......",shard);
		go getRecords(client, stream, shard.ShardId)
		wg.Add(1)
	}

	wg.Wait()
}

func getRecords(client *kinesis.Kinesis, stream *string, shardID *string) {


	fmt.Println("Get records........",client);

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

		fmt.Println("Get records.polling every sec.......");
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

			RawStringData := string(record.Data)

			fmt.Println("Get records...record.data......",string(record.Data));

			var rawSensorData RawSensorData
			json.Unmarshal([]byte(RawStringData), &rawSensorData)

			fmt.Println("the Raw data is ",rawSensorData)

			var dataFromStream SensorData
			dataFromStream = rawSensorData.State.Reported

			var n json.Number
			n = rawSensorData.Timestamp
			numToInt,err := n.Int64()
			if err!= nil {
				panic(err)
			}

			tm := time.Unix(numToInt,0)
			fmt.Println(tm)
			dataFromStream.Time = tm
			fmt.Print("Hello prettyyyy the data from stream.....",dataFromStream)


			StreamToSocket := StreamToSocket{
				Type:"RealTime",
				Data: dataFromStream,
			}

			b, err := json.Marshal(StreamToSocket)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("The final marshalled Json is......",string(b))

			pool.broadcast<-StreamToSocket
			pool.DeviceRegister<-StreamToSocket
		}

		shardIterator = records.NextShardIterator
	}
}
