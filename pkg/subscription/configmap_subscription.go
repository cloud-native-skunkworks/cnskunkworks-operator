package subscription

import (
	"context"
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
}

func (p *ConfigMapSubscription) Reconcile(object runtime.Object, event watch.EventType) {

	configMap := object.(*v1.ConfigMap)
	klog.Infof("ConfigMapSubscription event type %s for %s", event, configMap.Name)

	switch event {
	case watch.Added:
	case watch.Deleted:
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
