package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	//"log"
	"net"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	proxyTCPconnection = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "proxy_tcp_connection",
			Help: "Current number of established Connection",
		},
	)
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	ClusterEnvironment := strings.ToLower(os.Getenv("ClusterEnvironment"))

	if ClusterEnvironment == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

	// Register Prometheus Gauge
	prometheus.MustRegister(proxyTCPconnection) // prom
}

func main() {
	var (
		listenAddr     = flag.String("l", "", "local address to listen on")
		remoteAddr     = flag.String("r", "", "remote address to dial")
		logConnections = flag.Bool("logconn", false, "log connections")
		//metricsaddr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

	)

	flag.Parse()

	if *listenAddr == "" {
		log.Fatalf("must supply local address to listen on, -l option")
	}

	if *remoteAddr == "" {
		log.Fatalf("must supply remote address to dial, -r option")
	}

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("listening: %v", err)
	}

	log.Info("Staring Prom endpoint /metrics")

	// prom
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(":8080", nil)

	log.Info("Staring Proxy...")
	log.Printf("serving : %v %v ", *listenAddr, *remoteAddr)

	proxy(ln, *remoteAddr, *logConnections)

}

func proxy(ln net.Listener, remoteAddr string, logConnections bool) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		if logConnections {
			log.Printf("connected: %s", conn.RemoteAddr())
			proxyTCPconnection.Inc() // prom
		}

		go handle(conn, remoteAddr)
	}
}

func handle(conn net.Conn, remoteAddr string) {
	defer conn.Close()

	rconn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("dialing remote: %v", err)
		proxyTCPconnection.Dec()
		return
	}
	defer proxyTCPconnection.Dec()
	defer rconn.Close()

	copy(conn, rconn)
}

func copy(a, b io.ReadWriter) {
	done := make(chan struct{})
	go func() {
		io.Copy(a, b)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(b, a)
		done <- struct{}{}
	}()
	<-done
	<-done
}

/*
dns lookup

ips, err := net.LookupIP("google.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}

*/
