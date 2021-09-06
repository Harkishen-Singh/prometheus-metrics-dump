# Prometheus metrics dump

* Fetch Prometheus metrics and store in binary format
* Easy sharing
* Supports metric regex matching
* Independent of metric size

## Flags

```shell
[hsingh@fedora prometheus-metrics-dump]$ ./main -h
Usage of ./main:
  -concurrent-pull int
        Number of concurrent pulls for fetching a metric. (default 1)
  -end int
        End time in unix seconds for fetching data. By default, it sets the end to now. (default 1630922674280)
  -metric-regex string
        Regex string for metric name to be fetched. By default, it fetches all metrics. (default ".*")
  -output string
        Name of the output file for storing dump. (default "default_dump_oJnNPGsiuz", random suffix)
  -reader-url string
        URL of read storage.
  -start int
        Start time in unix seconds for fetching data.
  -timeout duration
        Timeout for read request. (default 1m0s)
```
