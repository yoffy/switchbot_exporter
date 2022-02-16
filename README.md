# switchbot_exporter

A Prometheus exporter for SwitchBot Meter.
It collects metrics via Bluetooth.

## requirements

* Linux
* go compiler

## install

```bash
$ make && sudo make install
```

It is installed as switchbot_exporter.service.

## uninstall

```bash
$ sudo make uninstall
```

## get metrics

```bash
$ curl http://localhost:9012/metrics
```
