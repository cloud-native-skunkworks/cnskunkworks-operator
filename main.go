/*
Copyright © 2020 alexsimonjones@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"context"
	"flag"

	"math/rand"
	"net/http"
	"time"

	"github.com/cloud-native-skunkworks/cnskunkworks-operator/pkg/runtime"
	"github.com/cloud-native-skunkworks/cnskunkworks-operator/pkg/subscription"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	minWatchTimeout = 5 * time.Minute
	timeoutSeconds  = int64(minWatchTimeout.Seconds() * (rand.Float64() + 1.0))
	masterURL       string
	kubeconfig      string
	addr            = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
)

func main() {

	flag.Parse()
	// Metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*addr, nil))
	}()
	// Tracing
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, err := cfg.New(
		"cnskunkworks_operator",
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	// Logs
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	// Run ....
	log.Info("Got watcher client...")

	kubeCfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	log.Info("Built config from flags...")

	defaultKubernetesClientSet, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		log.Fatalf("Error building watcher clientset: %s", err.Error())
	}

	// Context
	context := context.TODO()

	configMapSubscription := &subscription.ConfigMapSubscription{
		ClientSet:  defaultKubernetesClientSet,
		Ctx:        context,
		Completion: make(chan bool),
	}
	podSubscription := &subscription.PodSubscription{
		ClientSet:             defaultKubernetesClientSet,
		Ctx:                   context,
		Completion:            make(chan bool),
		ConfigMapSubscriptRef: configMapSubscription,
	}

	if err := runtime.RunLoop([]subscription.ISubscription{
		configMapSubscription,
		podSubscription,
	}); err != nil {
		log.Fatalf(err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

}
