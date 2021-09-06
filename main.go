package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	promConfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/timescale/promscale/pkg/log"
	"github.com/timescale/promscale/pkg/migration-tool/utils"
)

const (
	extension = ".zip"
	nameLabel = "__name__"
)

type config struct {
	mint, maxt     int64
	concurrentPull int
	metricRegex    string
	readerURL      string
	output         string
	timeout        time.Duration
}

func main() {
	handleErrIfAny(log.Init(log.Config{Level: "debug", Format: "logfmt"}))
	conf := new(config)
	args := os.Args[1:]
	parseFlags(conf, args)

	log.Info("msg", "initializing", "configuration", fmt.Sprintf("%v", conf))
	handleErrIfAny(validateConfig(conf))

	conf.mint *= 1000
	conf.maxt *= 1000

	conf.output = conf.output + extension

	client, err := prepareClient(conf)
	handleErrIfAny(err)

	matcher, err := labels.NewMatcher(labels.MatchRegexp, nameLabel, conf.metricRegex)
	handleErrIfAny(err)

	readRequest, err := utils.CreatePrombQuery(conf.mint, conf.maxt, []*labels.Matcher{matcher})
	handleErrIfAny(err)

	response, numBytesCompressed, numBytesUncompressed, err := client.Read(context.Background(), readRequest, "")
	handleErrIfAny(err)

	log.Info("msg", "response received", "compressed bytes", numBytesCompressed, "uncompressed bytes", numBytesUncompressed)

	tsRefs := response.Timeseries
	log.Info("msg", "time-series received", "count", len(tsRefs))

	bSlice, err := response.Marshal()
	handleErrIfAny(err)

	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	_, err = w.Write(bSlice)
	handleErrIfAny(err)

	handleErrIfAny(w.Flush())

	f, err := os.Create(conf.output)
	handleErrIfAny(err)

	numWritten, err := f.Write(compressed.Bytes())
	handleErrIfAny(err)

	log.Info("msg", fmt.Sprintf("written %d bytes", numWritten))
}


func prepareClient(cfg *config) (*utils.Client, error) {
	clientConf := utils.ClientConfig{
		URL:       cfg.readerURL,
		Timeout:   cfg.timeout,
		OnTimeout: utils.Retry,
		OnErr:     utils.Abort,
		MaxRetry:  5,
		Delay:     5 * time.Second,
	}
	client, err := utils.NewClient("prometheus_metric_binary", clientConf, promConfig.HTTPClientConfig{})
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	return client, nil
}

func parseFlags(cfg *config, args []string) {
	flag.StringVar(&cfg.readerURL, "reader-url", "", "URL of read storage.")
	flag.StringVar(&cfg.output, "output", "default_dump_"+randomString(10), "Name of the output file for storing dump.")
	flag.StringVar(&cfg.metricRegex, "metric-regex", ".*", "Regex string for metric name to be fetched. By default, it fetches all metrics.")
	flag.DurationVar(&cfg.timeout, "timeout", time.Minute, "Timeout for read request.")
	flag.IntVar(&cfg.concurrentPull, "concurrent-pull", 1, "Number of concurrent pulls for fetching a metric.")
	flag.Int64Var(&cfg.mint, "start", 0, "Start time in unix seconds for fetching data.")
	flag.Int64Var(&cfg.maxt, "end", timestamp.FromTime(time.Now()), "End time in unix seconds for fetching data. By default, it sets the end to now.")
	_ = flag.CommandLine.Parse(args)
}

func validateConfig(cfg *config) error {
	if cfg.mint == 0 {
		return fmt.Errorf("'mint' cannot be 0")
	}
	return nil
}

func handleErrIfAny(err error) {
	if err == nil {
		return
	}
	log.Fatal("error", err.Error())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(numChars int) string {
	b := make([]byte, numChars)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
