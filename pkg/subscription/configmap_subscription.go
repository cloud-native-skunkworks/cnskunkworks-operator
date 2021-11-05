package subscription

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/siruspen/logrus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type ConfigMapSubscription struct {
	watcherInterface watch.Interface
	ClientSet        kubernetes.Interface
	Ctx              context.Context
	Completion       chan bool
	// Specific behaviours for our Operator
	PlatformConfig *platformConfig
	// ADDED
	// DELETED
	platformConfigPhase string
}

var (
	platformConfigMapName                   = "platform-default-configmap"
	platformConfigMapNamespace              = "kube-system"
	prometheusPlatformConfigAnnotationCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cnskunkworks_platform_config_annotation_count",
		Help: "This tell us the number of annotations with  the configmap",
	})
	prometheusPlatformConfigAvailibilityGuage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cnskunkworks_platformconfig_availibity",
		Help: "This tells whether a platform config is available",
	},
		[]string{"configmap_name", "namespace"})
)

type platformAnnotation struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
type platformConfig struct {
	Annotations []platformAnnotation `yaml:"annotations"`
}

func isPlatformConfigMap(configMap *v1.ConfigMap) (bool, error) {
	if configMap == nil {
		return false, errors.New("empty configmap")
	}
	if configMap.Name == platformConfigMapName {
		return true, nil
	}
	return false, nil
}
func (p *ConfigMapSubscription) Reconcile(object runtime.Object, event watch.EventType) {
	rootSpan := opentracing.GlobalTracer().StartSpan("configMapSubscriptionReoncile")
	defer rootSpan.Finish()

	configMap := object.(*v1.ConfigMap)
	isPlatformConfigSpan := opentracing.GlobalTracer().StartSpan(
		"isplatformconfigspan", opentracing.ChildOf(rootSpan.Context()),
	)
	defer isPlatformConfigSpan.Finish()
	if ok, err := isPlatformConfigMap(configMap); !ok {
		if err != nil {
			klog.Error(err)
		}
		return
	}

	klog.Infof("ConfigMapSubscription event type %s for %s", event, configMap.Name)
	switch event {
	case watch.Added:
		watchEventAdd := opentracing.GlobalTracer().StartSpan(
			"watchEventAdd", opentracing.ChildOf(isPlatformConfigSpan.Context()),
		)
		defer watchEventAdd.Finish()
		p.platformConfigPhase = string(event)
		rawDefaultsString := configMap.Data["platform-defaults"]
		var unMarshalledData platformConfig
		err := yaml.Unmarshal([]byte(rawDefaultsString), &unMarshalledData)
		if err != nil {
			log.Error(err)
			return
		}
		p.PlatformConfig = &unMarshalledData
		prometheusPlatformConfigAvailibilityGuage.WithLabelValues(configMap.Name,
			configMap.Namespace).Set(float64(1))
		prometheusPlatformConfigAnnotationCount.Set(float64(len(p.PlatformConfig.Annotations)))
	case watch.Deleted:
		watchEventDeleted := opentracing.GlobalTracer().StartSpan(
			"watchEventDeleted", opentracing.ChildOf(isPlatformConfigSpan.Context()),
		)
		defer watchEventDeleted.Finish()
		// wipe out the platformConfigMapAnnotations
		// On a delete how do we know which annotations to remove from the pod?
		p.platformConfigPhase = string(event)
		p.PlatformConfig = nil
		prometheusPlatformConfigAvailibilityGuage.WithLabelValues(configMap.Name,
			configMap.Namespace).Set(float64(0))
		prometheusPlatformConfigAnnotationCount.Set(0)
	case watch.Modified:
		watchEventModified := opentracing.GlobalTracer().StartSpan(
			"watchEventModified", opentracing.ChildOf(isPlatformConfigSpan.Context()),
		)
		defer watchEventModified.Finish()
	}

}

func (p *ConfigMapSubscription) Subscribe() (watch.Interface, error) {
	var err error
	p.watcherInterface, err = p.ClientSet.CoreV1().ConfigMaps("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	log.Info("Started watch stream for ConfigMapSubscription")
	return p.watcherInterface, nil
}

func (p *ConfigMapSubscription) IsComplete() <-chan bool {

	return p.Completion
}
