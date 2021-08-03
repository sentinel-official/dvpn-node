package http

import (
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

	var (
		mux = cmux.New(l)
	)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	go func() {
		_ = http.Serve(
			mux.Match(cmux.Any()),
			handler,
		)
	}()

	go func() {
		_ = http.Serve(
			tls.NewListener(
				mux.Match(cmux.TLS()),
				&tls.Config{
					Certificates: []tls.Certificate{
						cert,
					},
				},
			),
			handler,
		)
	}()

	return mux.Serve()
}
