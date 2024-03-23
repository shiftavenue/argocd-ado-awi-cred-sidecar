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

	// TODO: Loop it with a ticker https://pkg.go.dev/time#NewTicker or check accessToken validity
	kuberneteshelper, err := pkg.NewKubernetesHelper(log)
	if err != nil {
		log.Error(err, "Failed to create KubernetesHelper")
		return
	}
	config, err := pkg.ParseConfig()
	if err != nil {
		return
	}
	azurehelper, err := pkg.NewAzureHelper(log)
	if err != nil {
		return
	}

	var expirationTime time.Time

	for {
		current := time.Now()
		remainingTime := int(expirationTime.Sub(current).Seconds())
		bufferTime := int((time.Minute * 5).Seconds())

		if remainingTime < bufferTime {
			log.Info("Refreshing access token")
			secret, err := kuberneteshelper.SearchSecret(config.Namespace, config.MatchUrl)
			if err != nil {
				return
			} else if secret == nil {
				log.Info("No secret found")
				return
			}
			log.Info("Secret found", "secretName", secret.Name, "namespace", secret.Namespace)

			accessToken, err := azurehelper.GetAzureDevOpsAccessToken()
			if err != nil {
				return
			}
			expirationTime = accessToken.ExpiresOn
			log.Info("Access token retrieved", "expiration", accessToken.ExpiresOn)
			err = kuberneteshelper.UpdateSecret(accessToken.Token, secret)
			if err != nil {
				log.Error(err, "Failed to update secret")
				return
			}
		} else {
			log.Info("Access token is still valid", "expiration", expirationTime, "remainingTime", remainingTime)
		}
		time.Sleep(60 * time.Second)
	}
}
