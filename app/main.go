package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
)

var (
	// Config options
	addr        = flag.String("addr", ":8080", "TCP address to listen to")
	kafka       = flag.String("kafka", "127.0.0.1:9092", "Kafka endpoints")
	enableKafka = flag.Bool("enable-kafka", false, "Enable Kafka or nor")
	amqpUri     = flag.String("amqp", "amqp://guest:guest@127.0.0.1:5672/", "AMQP URI")
	enableAmqp  = flag.Bool("enable-amqp", false, "Enable AMQP or not")
	sqsUri      = flag.String("sqs-uri", "", "SQS URI")
	sqsId       = flag.String("sqs-id", "", "SQS Access id")
	sqsSecret   = flag.String("sqs-secret", "", "SQS Secret key")
	enableSqs   = flag.Bool("enable-sqs", false, "Enable SQS or not")

	producer   sarama.SyncProducer
	queue      amqp.Queue
	ch         *amqp.Channel
	sqsSession *session.Session
	sqsSvc     *sqs.SQS

	// Declaring prometheus metrics
	apiDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "api_durations_seconds",
			Help:       "API duration seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	)

	apiRequestsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_count",
			Help: "API requests count",
		},
		[]string{"backend"},
	)

	apiErrorsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_errors_count",
			Help: "API errors count",
		},
		[]string{"type", "backend"},
	)
)

func init() {
	// Registering prometheus metrics
	prometheus.MustRegister(apiDurations)
	prometheus.MustRegister(apiRequestsCount)
	prometheus.MustRegister(apiErrorsCount)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func main() {
	var err error
	var conn *amqp.Connection
	flag.Parse()

	if *enableKafka {
		log.Printf("Enabling kafka: %s\n", *kafka)
		// Setup producer
		config := sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForLocal
		config.Producer.Compression = sarama.CompressionSnappy
		config.Producer.Return.Successes = true
		brokers := strings.Split(*kafka, ",")
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err != nil {
			panic(err)
		}
		defer producer.Close()
	}

	if *enableAmqp {
		log.Printf("Enabling amqp: %s\n", *amqpUri)
		conn, err = amqp.Dial(*amqpUri)
		if err != nil {
			log.Fatalf("Can't connect to rabbitmq")
		}
		defer conn.Close()

		ch, err = conn.Channel()
		if err != nil {
			log.Fatalf("Can't open channel")
		}
		defer ch.Close()

		queue, err = ch.QueueDeclare(
			"load", // name
			true,   // durable - flush to disk
			false,  // delete when unused
			false,  // exclusive - only accessible by the connection that declares
			false,  // no-wait - the queue will assume to be declared on the server
			nil,    // arguments -
		)
		if err != nil {
			log.Fatalf("Can't create queue")
		}
	}

	if *enableSqs {
		log.Printf("Enabling sqs: %s\n", *sqsUri)
		sqsSession = session.New(&aws.Config{
			Region:      aws.String("ru-central1"),
			Credentials: credentials.NewStaticCredentials(*sqsId, *sqsSecret, ""),
			MaxRetries:  aws.Int(5),
			Endpoint:    aws.String("message-queue.api.cloud.yandex.net"),
		})

		sqsSvc = sqs.New(sqsSession)

	}

	http.HandleFunc("/status", statusHandlerFunc)
	http.HandleFunc("/post/kafka", postKafkaHandlerFunc)
	http.HandleFunc("/post/amqp", postAMQPHandlerFunc)
	http.HandleFunc("/post/sqs", postSQSHandlerFunc)
	http.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

// STATUS method
func statusHandlerFunc(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

func isJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

// POST Method
func postKafkaHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !*enableKafka {
		http.Error(w, "Kafka endpoint is disabled", http.StatusNotFound)
		return
	}

	// Incrementing requests count metric
	apiRequestsCount.WithLabelValues("kafka").Inc()

	// Observing request time
	timer := prometheus.NewTimer(apiDurations)
	defer timer.ObserveDuration()

	// Reading post body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiErrorsCount.WithLabelValues("body", "kafka").Inc()
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}

	// Checking if json is correct
	if !isJSON(body) {
		apiErrorsCount.WithLabelValues("json", "kafka").Inc()
		log.Printf("Invalid json provided")
		http.Error(w, "Can't parse json", http.StatusBadRequest)
		return
	}

	// Posting data to kafka
	msg := &sarama.ProducerMessage{Topic: "loader", Value: sarama.ByteEncoder(body)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		apiErrorsCount.WithLabelValues("kafka", "kafka").Inc()
		log.Printf("Kafka Error: %s\n", err)
		http.Error(w, "Can't post to kafka", http.StatusInternalServerError)
		return
	}

	// Writing response
	response := struct {
		Status    string `json:"status"`
		Partition int32  `json:"partition"`
		Offset    int64  `json:offset`
	}{"ok", partition, offset}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func postSQSHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !*enableSqs {
		http.Error(w, "SQS endpoint is disabled", http.StatusNotFound)
		return
	}

	// Incrementing requests count metric
	apiRequestsCount.WithLabelValues("sqs").Inc()

	// Observing request time
	timer := prometheus.NewTimer(apiDurations)
	defer timer.ObserveDuration()

	// Reading post body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiErrorsCount.WithLabelValues("body", "sqs").Inc()
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}

	// Checking if json is correct
	if !isJSON(body) {
		apiErrorsCount.WithLabelValues("json", "sqs").Inc()
		log.Printf("Invalid json provided")
		http.Error(w, "Can't parse json", http.StatusBadRequest)
		return
	}

	// Posting data to sqs

	params := &sqs.SendMessageInput{
		MessageBody: aws.String(string(body)), // Required
		QueueUrl:    aws.String(*sqsUri),      // Required
	}
	_, err = sqsSvc.SendMessage(params)

	if err != nil {
		apiErrorsCount.WithLabelValues("sqs", "sqs").Inc()
		log.Printf("SQS Error: %s\n", err)
		http.Error(w, "Can't post to sqs", http.StatusInternalServerError)
		return
	}

	// Writing response
	response := struct {
		Status string `json:"status"`
	}{"ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST Method
func postAMQPHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if !*enableAmqp {
		http.Error(w, "Amqp endpoint is disabled", http.StatusNotFound)
		return
	}

	// Incrementing requests count metric
	apiRequestsCount.WithLabelValues("amqp").Inc()

	// Observing request time
	timer := prometheus.NewTimer(apiDurations)
	defer timer.ObserveDuration()

	// Reading post body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiErrorsCount.WithLabelValues("body", "amqp").Inc()
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}

	// Checking if json is correct
	if !isJSON(body) {
		apiErrorsCount.WithLabelValues("json", "amqp").Inc()
		log.Printf("Invalid json provided")
		http.Error(w, "Can't parse json", http.StatusBadRequest)
		return
	}

	err = ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory - could return an error if there are no consumers or queue
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})

	if err != nil {
		apiErrorsCount.WithLabelValues("amqp", "amqp").Inc()
		log.Printf("AMQP Error: %s\n", err)
		http.Error(w, "Can't post to amqp", http.StatusInternalServerError)
		return
	}

	// Writing response
	response := struct {
		Status string `json:"status"`
	}{"ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
