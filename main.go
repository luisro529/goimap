package main

import (
	"crypto/tls"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/stbenjam/go-imap-notmuch/pkg/config"
	"github.com/stbenjam/go-imap-notmuch/pkg/notmuch"

	"github.com/emersion/go-imap/server"
)

func main() {
	logrus.Infof("go-imap-notmuch starting")
	logrus.SetLevel(logrus.InfoLevel)
	var cfg *config.Config
	var err error

	if len(os.Args) > 1 {
		cfg, err = config.New(os.Args[1])
		if err != nil {
			logrus.WithError(err).Fatalf("couldn't read configuration")
		}
	} else {
		fmt.Fprintf(os.Stderr, "usage: %s <config file>\n", os.Args[0])
		os.Exit(1)
	}
	logrus.Infof("configuration loaded")

	be, err := notmuch.New(cfg)
	if err != nil {
		logrus.WithError(err).Fatalf("couldn't initialize notmuch database")
	}
	logrus.Infof("notmuch database initialized")

	s := server.New(be)

	if cfg.Debug {
		s.Debug = os.Stderr
	}
	s.Addr = ":9143"

	if cfg.TLSCertificate != "" && cfg.TLSKey != "" {
		logrus.WithFields(logrus.Fields{
			"tlsCert": cfg.TLSCertificate,
			"tlsKey":  cfg.TLSKey,
		}).Infof("starting with TLS enabled")

		certs, err := tls.LoadX509KeyPair(cfg.TLSCertificate, cfg.TLSKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not load certificates: %s", err.Error())
			os.Exit(1)
		}

		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certs},
			MinVersion:   tls.VersionTLS10,
			MaxVersion:   tls.VersionTLS13,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			PreferServerCipherSuites: true,
		}
	} else {
		s.AllowInsecureAuth = true
	}
	if err := s.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
