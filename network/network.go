package network

import (
	"crypto/tls"
	"errors"
	"github.com/zxchris/swaggerly/config"
	"github.com/zxchris/swaggerly/logger"
	"net"
)

func GetListener() (net.Listener, error) {

	cfg, _ := config.Get() // Don't worry about error. If there was something wrong with the config, we'd know by now.

	useTLS := 0
	if len(cfg.TLSCertificate) > 0 {
		useTLS++
	}
	if len(cfg.TLSKey) > 0 {
		useTLS++
	}

	// If no cert & key, then we're to run in plain-text mode
	if useTLS == 0 {
		logger.Infof(nil, "listening on %s for unsecured connections", cfg.BindAddr)
		return net.Listen("tcp", cfg.BindAddr)
	}

	if useTLS == 1 {
		return nil, errors.New("You must provide both a certificate and a key to enable TLS")
	}

	// Okay, we're building a TLS listener
	crt, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSKey)
	if err != nil {
		return nil, err
	}

	// Be really secure!
	tlscfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		Certificates: []tls.Certificate{crt},
	}

	logger.Infof(nil, "listening on %s for SECURED connections", cfg.BindAddr)
	return tls.Listen("tcp", cfg.BindAddr, tlscfg)

	//w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	// TODO Do we need to disable HTTP/2 ? : TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
}
