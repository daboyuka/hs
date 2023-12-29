package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

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
		log.Println(r)
		w.WriteHeader(200)
	})
}

var httpsFlag *bool

func init() {
	httpsFlag = rootCmd.PersistentFlags().BoolP("secure", "s", false, "listen for HTTPS (instead of HTTP)")
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
