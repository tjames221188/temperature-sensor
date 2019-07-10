package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d2r2/go-dht"
	dhtLogger "github.com/d2r2/go-logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/urfave/cli"
)

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Exiting....")
		os.Exit(0)
	}()

	var gateway_host, gateway_port string

	dhtLogger.ChangePackageLogLevel("dht", dhtLogger.InfoLevel)
	app := cli.NewApp()
	app.Name = "Temperature Sensor"
	app.Usage = "Foo"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "metrics-gateway, g",
			Value: "metrics.englishlanguageitutoring.com",
			Usage: "The prometheus pushgateway host",
			Destination: &gateway_host,
		},
		cli.StringFlag{
			Name: "metrics-gateway-port, p",
			Value: "9091",
			Usage: "The prometheus pushgateway port",
			Destination: &gateway_port,
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Printf("host: %s\nport: %s\n", gateway_host, gateway_port)
		sensorType := dht.DHT22
		pin := 14

		// Metrics
		temperatureCollector := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "office_temperature_celsius",
			Help: "The temperature in the Bristol office measured in celsius",
		})
		humidityCollector := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "office_humidity_percent",
		})

		pusher := push.New(gateway_host + ":" + gateway_port, "office_monitor").
			Collector(temperatureCollector).
			Collector(humidityCollector)

		for {
			temperature, humidity, retried, err := dht.ReadDHTxxWithRetry(sensorType, pin, false, 10)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("Temp: %.2f, Humidity: %.2f (retried %d times)", temperature, humidity, retried)
			temperatureCollector.Set(float64(temperature))
			humidityCollector.Set(float64(humidity))
			err = pusher.Push()
			if err != nil {
				log.Println(err)
			}

			time.Sleep(5 * time.Minute)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
