# switchbot_exporter

A Prometheus exporter for SwitchBot Meter.
It collects metrics via Bluetooth.

## requirements

* Linux
* go compiler

## build

```bash
$ go build
```

## run

```bash
$ sudo ./switchbot_exporter &
```

## get metrics

```bash
$ curl http://localhost:9012/metrics
```

## stop

```bash
$ pkill switchbot_exporter
```
