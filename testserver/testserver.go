package main

import (
	"crypto/tls"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var server = http.Server{
	Addr: ":8080",
	TLSConfig: &tls.Config{
		Certificates: []tls.Certificate{}, // populated in cert.go
	},
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := delay()
		time.Sleep(d)
		log.Println(d, r)
		w.WriteHeader(*statusFlag)
	})
}

func delay() time.Duration {
	t, tmax := *reqTime, *maxReqTime
	if t < tmax {
		t += time.Duration(rand.Int63n(int64(tmax - t)))
	}
	return t
}

var httpsFlag *bool
var statusFlag *int
var reqTime *time.Duration
var maxReqTime *time.Duration

func init() {
	httpsFlag = rootCmd.PersistentFlags().BoolP("secure", "s", false, "listen for HTTPS (instead of HTTP)")
	statusFlag = rootCmd.PersistentFlags().IntP("status", "S", 200, "set response status code (default 200)")
	reqTime = rootCmd.PersistentFlags().DurationP("reqtime", "t", 0, "set min response delay time (default 0)")
	maxReqTime = rootCmd.PersistentFlags().DurationP("maxreqtime", "T", 0, "set max response delay time (default 0)")
}

var rootCmd = cobra.Command{
	Use:   "testserver [bindaddr]",
	Short: "a simple HTTP server for testing",
	Long: "a simple HTTP server for testing.\n\n" +
		"When -s is used for HTTPS mode, use certificate 'testserver/key.pem' to trust the server.",
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			server.Addr = args[0]
		}

		if *httpsFlag {
			return server.ListenAndServeTLS("", "")
		} else {
			return server.ListenAndServe()
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
