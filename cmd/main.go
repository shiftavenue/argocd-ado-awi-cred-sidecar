package main

import (
	"flag"
	"time"

	"github.com/shiftavenue/argocd-ado-awi-cred-sidecar/pkg"
	"go.uber.org/zap/zapcore"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	kzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var logLevel int

func main() {
	flag.IntVar(&logLevel, "log-level", 0,
		"A zap log level should be multiplied by -1 to get the logr verbosity. For example, to get logr verbosity of 3, pass zapcore.Level(-3) to this Opts. See https://pkg.go.dev/github.com/go-logr/zapr for how zap level relates to logr verbosity.")
	flag.Parse()
	log := kzap.New(kzap.Level(zapcore.Level(logLevel)))

	klog.SetLogger(log)
	coordinator, err := pkg.NewCoordinator(log)
	if err != nil {
		log.Error(err, "Failed to create coordinator")
		return
	}

	for {
		err := coordinator.EvaluateAccessTokenExpiration()
		if err != nil {
			log.Error(err, "Failed to evaluate access token expiration time")
		}
		time.Sleep(60 * time.Second)
	}
}