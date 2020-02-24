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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Config options
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	kafka    = flag.String("kafka", "127.0.0.1:9092", "Kafka endpoints")
	producer sarama.SyncProducer

	// Declaring prometheus metrics
	apiDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "api_durations_seconds",
			Help:       "API duration seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	)

	apiRequestsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "api_requests_count",
			Help: "API requests count",
		})

	apiErrorsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_errors_count",
			Help: "API errors count",
		},
		[]string{"type"},
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
	flag.Parse()


    log.Printf("Got kafka addr: %s\n", *kafka)
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

	http.HandleFunc("/status", statusHandlerFunc)
	http.HandleFunc("/post", postHandlerFunc)
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
func postHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// Incrementing requests count metric
	apiRequestsCount.Inc()

	// Observing request time
	timer := prometheus.NewTimer(apiDurations)
	defer timer.ObserveDuration()

	// Reading post body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiErrorsCount.WithLabelValues("body").Inc()
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Can't read body", http.StatusBadRequest)
		return
	}

	// Checking if json is correct
	if !isJSON(body) {
		apiErrorsCount.WithLabelValues("json").Inc()
		log.Printf("Invalid json provided")
		http.Error(w, "Can't parse json", http.StatusBadRequest)
		return
	}

	// Posting data to kafka
	msg := &sarama.ProducerMessage{Topic: "loader", Value: sarama.ByteEncoder(body)}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		apiErrorsCount.WithLabelValues("kafka").Inc()
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
