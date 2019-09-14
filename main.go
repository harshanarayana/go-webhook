package main

import "net/http"

func main() {

	http.HandleFunc("/mutating/add-secret-label", MutatingLabelHandler)
	http.HandleFunc("/alive", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	server := &http.Server{
		Addr: ":6543",
		TLSConfig: TLSConfig(Config{
			CertFile: "/data/ssl/cert.pem",
			KeyFile:  "/data/ssl/key.pem",
		}),
	}
	//noinspection GoUnhandledErrorResult
	server.ListenAndServeTLS("", "")
}
