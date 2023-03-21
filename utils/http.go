package utils

import (
	"crypto/rand"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
)

func ListenAndServeTLS(address, certFile, keyFile string, handler http.Handler) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	var (
		mux    = cmux.New(l)
		tlsMux = mux.Match(cmux.TLS())
		anyMux = mux.Match(cmux.Any())
	)

	go func() {
		if err := http.Serve(
			tls.NewListener(
				tlsMux,
				&tls.Config{
					Certificates: []tls.Certificate{
						cert,
					},
					Rand: rand.Reader,
				},
			),
			handler,
		); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := http.Serve(
			anyMux,
			handler,
		); err != nil {
			panic(err)
		}
	}()

	return mux.Serve()
}
