package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/cmd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var listen = flag.String("listen", ":9012", "metrics listen address")
	flag.Parse()

	// https://github.com/go-ble/ble/tree/master/examples

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)
	dev := d.(*linux.Device)

	// https://electronics.stackexchange.com/questions/82098/ble-scan-interval-and-window
	if err := dev.HCI.Send(&cmd.LESetScanParameters{
		LEScanType:           0x01,   // 0x00: passive, 0x01: active
		LEScanInterval:       0x0040, // 0x0004 - 0x4000; N * 0.625msec
		LEScanWindow:         0x0030, // 0x0004 - 0x4000; N * 0.625msec
		OwnAddressType:       0x01,   // 0x00: public, 0x01: random
		ScanningFilterPolicy: 0x00,   // 0x00: accept all, 0x01: ignore non-white-listed.
	}, nil); err != nil {
		panic(err)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*listen, nil))
	}()

	for {
		// 60 secs in 1 cycle (11 + 49)
		ctx, cancel := context.WithTimeout(context.Background(), 11*time.Second)
		ble.Scan(ble.WithSigHandler(ctx, cancel), true, advHandler, nil) // heavy function
		time.Sleep(49*time.Second)
	}
}

var (
	ns = "switchbot"
	batteryGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "battery",
			Help:        "battery level (0-100)",
		},
		[]string{"hw"})
	temperatureGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "temperature",
			Help:        "temperature in Celsius",
		},
		[]string{"hw"})
	humidityGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "humidity",
			Help:        "humidity (0-100)",
		},
		[]string{"hw"})
	positionGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "position",
			Help:        "position (0-100)",
		},
		[]string{"hw"})
	brightnessGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   ns,
			Name:        "brightness",
			Help:        "brightness (0-10)",
		},
		[]string{"hw"})
)

func advHandler(a ble.Advertisement) {
	found := false
	services := a.Services()
	for _, uuid := range services {
		if uuid.String() == "cba20d00224d11e69fb80002a5d5c51b" {
			found = true
		}
	}
	if !found {
		return
	}

	addr := a.Addr().String()
	for _, data := range a.ServiceData() {
		switch data.Data[0] {
		case 0x54:
			// SwitchBot MeterTH
			// spec: https://github.com/OpenWonderLabs/python-host/wiki/Meter-BLE-open-API
			if len(data.Data) < 6 {
				continue
			}

			battery := float64(data.Data[2] & 0x7f)
			temp := float64(data.Data[3] & 0xff) / 10
			temp += float64(data.Data[4] & 0x7f)
			humidity := float64(data.Data[5] & 0x7f)

			batteryGauge.With(prometheus.Labels{"hw": addr}).Set(battery)
			temperatureGauge.With(prometheus.Labels{"hw": addr}).Set(temp)
			humidityGauge.With(prometheus.Labels{"hw": addr}).Set(humidity)

		case 0x63:
			// SwitchBot Curtain
			// spec: https://github.com/OpenWonderLabs/python-host/wiki/Curtain-BLE-open-API
			if len(data.Data) < 5 {
				continue
			}

			battery := float64(data.Data[2] & 0x7f)
			position := float64(data.Data[3] & 0x7f)
			brightness := float64(data.Data[4] >> 4)

			batteryGauge.With(prometheus.Labels{"hw": addr}).Set(battery)
			positionGauge.With(prometheus.Labels{"hw": addr}).Set(position)
			brightnessGauge.With(prometheus.Labels{"hw": addr}).Set(brightness)
		}
	}
}
