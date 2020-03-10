package main

import (
	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	KafkaBrokers  = flag.String("kafka-brokers", "127.0.0.1:9092", "Kafka endpoints")
	KafkaTopic    = flag.String("kafka-topic", "test", "Kafka topic to read from")
	ClickHouseDSN = flag.String("ch-dsn", "tcp://localhost:9000/test_database", "Clickhouse URI")

	// Messages Channel
	msgChan = make(chan *sarama.ConsumerMessage)

	jsonErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tpc_streaming_json_errors",
		Help: "Number of json errors when unmarshaling",
	})

	jsonUnknownFields = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tpc_streaming_json_unknown_fields",
		Help: "Number of unknown fields in json",
	})

	eventsRead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tpc_streaming_kafka_events_read",
		Help: "Number of events have been read from kafka",
	})

	kafkaTopicLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tpc_streaming_kafka_topic_lag",
			Help: "Count of non-found fileds in dict",
		},
		[]string{"topic", "partition"},
	)

	ChFlushTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tpc_streaming_analytics_ch_clickhouse_flush_time",
		Help: "Clickhouse last flush time",
	})

	ChEventsFlushed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tpc_streaming_analytics_ch_clickhouse_events_flushed",
		Help: "Number of events have been flushed to clickhouse",
	})
)

func init() {
	// Initializing prometheus metrics
	prometheus.MustRegister(jsonErrors)
	prometheus.MustRegister(jsonUnknownFields)
	prometheus.MustRegister(eventsRead)
	prometheus.MustRegister(kafkaTopicLag)
	prometheus.MustRegister(ChFlushTime)
	prometheus.MustRegister(ChEventsFlushed)
}

type Consumer struct {
	ready chan bool
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		msgChan <- message
		session.MarkMessage(message, "")
	}

	return nil
}

func main() {
	// Kafka message represents one message from kafka
	var kafkaMsg *sarama.ConsumerMessage
	var err error
	// variable for storing unmarshaled events
	events := []event{}

	// Getting configuration
	flag.Parse()

	// Creating destination to flush into
	err = initClickhouse()
	if err != nil {
		log.Fatalf("ERROR: Can't init clickhouse: %s", err.Error())
	}

	// connecting to kafka
	log.Printf("INFO: Connecting to kafka")
	version, err := sarama.ParseKafkaVersion("2.3.1")
	if err != nil {
		log.Panicf("ERROR: Can't parse Kafka version: %v", err)
	}
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(*KafkaBrokers, ","), "go-kafka-streaming", config)
	if err != nil {
		log.Panicf("ERROR: Can't create consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, strings.Split(*KafkaTopic, ","), &consumer); err != nil {
				log.Panicf("ERROR: consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	log.Printf("INFO: Successfully connected to kafka brokers")

	// Starting metrics listener
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":9146", nil))
	}()

	// Channel which is used to notify when flush is needed
	flushTicker := time.NewTicker(10 * time.Second).C

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Events processing section
	for {
		select {
		case <-ctx.Done():
			log.Printf("INFO: Context cancelled")
			cancel()
			wg.Wait()
			return
		// Receive event from kafka
		case kafkaMsg = <-msgChan:
			// log.Printf("INFO: %s/%d/%d\t%s\t%s\n", kafkaMsg.Topic, kafkaMsg.Partition, kafkaMsg.Offset, kafkaMsg.Key, kafkaMsg.Value)
			cur := event{}
			err := cur.unmarshalJSON(kafkaMsg.Value)
			if err != nil {
				jsonErrors.Inc()
				log.Printf("WARN: Can't Unmashal JSON: %s\n", err.Error())
				continue
			}

			events = append(events, cur)
			eventsRead.Inc()
		case <-flushTicker:
			// flushing data to clickhouse
			err := flushEvents(events)
			if err != nil {
				log.Fatalf("ERROR: Failed to flush to clickhouse: %s\n", err.Error())
			}

			// Resetting events
			events = nil
			// Catch ctrl+c to exit
		case <-signals:
			cancel()
			wg.Wait()
			return
		}
	}
}
