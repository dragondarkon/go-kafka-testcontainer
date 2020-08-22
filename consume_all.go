package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
)

func TestConsume(Topic string, endOffset int) []Msg {
	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true

	brokers := strings.Split(kafkaConn, ",")
	client, err := sarama.NewClient(brokers, config) // I am not giving any configuration
	if err != nil {
		panic(err)
	}
	lastoffset, err := client.GetOffset(Topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}
	if endOffset == -1 {
		endOffset = int(lastoffset)
	}
	lastoffset = (lastoffset)
	fmt.Println(endOffset)
	fmt.Println(lastoffset)
	if endOffset > int(lastoffset) {
		return nil
	}

	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	topics := []string{Topic}

	consumer, errors := consume(topics, master)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0
	index := endOffset - 1
	// Get signnal for finish
	doneCh := make(chan struct{})
	msgs := []Msg{}
	go func() {
		for {
			if index >= 0 {
				select {
				case msg := <-consumer:
					msgCount++
					index--
					res := map[string]interface{}{}
					if err := json.Unmarshal(msg.Value, &res); err != nil {
						fmt.Println(err)
					}
					m := Msg{
						Partition: msg.Partition,
						Offset:    msg.Offset,
						Key:       string(msg.Key),
						Value:     res,
					}
					msgs = append(msgs, m)
					fmt.Println()
					fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
					fmt.Println()
				case consumerError := <-errors:
					msgCount++
					fmt.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
					doneCh <- struct{}{}
				case <-signals:
					fmt.Println("Interrupt is detected")
					doneCh <- struct{}{}
				}
			} else {
				doneCh <- struct{}{}
				fmt.Println("End")

			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")
	return msgs
}

func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)
		// this only consumes partition no 1, you would probably want to consume all partitions
		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			fmt.Printf("Topic %v Partitions: %v", topic, partitions)
			panic(err)
		}
		fmt.Println(" Start consuming topic ", topic)
		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					errors <- consumerError
					fmt.Println("consumerError: ", consumerError.Err)

				case msg := <-consumer.Messages():
					consumers <- msg
					// fmt.Println("Got message on topic ", topic, string(msg.Value))
				}
			}
		}(topic, consumer)
	}

	return consumers, errors
}
