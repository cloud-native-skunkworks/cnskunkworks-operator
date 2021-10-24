package subscription

import (
	"context"
	"errors"
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
	platformConfigMapName      = "platform-default-configmap"
	platformConfigMapNamespace = "kube-system"
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

	configMap := object.(*v1.ConfigMap)
	if ok, err := isPlatformConfigMap(configMap); !ok {
		if err != nil {
			klog.Error(err)
		}
		return
	}

	klog.Infof("ConfigMapSubscription event type %s for %s", event, configMap.Name)
	switch event {
	case watch.Added:
		// Populate the platformConfigMapAnnotations map
		p.platformConfigPhase = string(event)
		rawDefaultsString := configMap.Data["platform-defaults"]
		var unMarshalledData platformConfig
		err := yaml.Unmarshal([]byte(rawDefaultsString), &unMarshalledData)
		if err != nil {
			klog.Error(err)
			return
		}
		p.PlatformConfig = &unMarshalledData
	case watch.Deleted:
		// wipe out the platformConfigMapAnnotations
		// On a delete how do we know which annotations to remove from the pod?
		p.platformConfigPhase = string(event)
		p.PlatformConfig = nil
	case watch.Modified:
	}

}

func (p *ConfigMapSubscription) Subscribe() (watch.Interface, error) {
	var err error
	p.watcherInterface, err = p.ClientSet.CoreV1().ConfigMaps("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	klog.Info("Started watch stream for ConfigMapSubscription")
	return p.watcherInterface, nil
}

func (p *ConfigMapSubscription) IsComplete() <-chan bool {

	return p.Completion
}
