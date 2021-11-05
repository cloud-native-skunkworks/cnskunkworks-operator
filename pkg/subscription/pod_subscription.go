package subscription

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	log "github.com/siruspen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type PodSubscription struct {
	watcherInterface watch.Interface
	ClientSet        kubernetes.Interface
	Ctx              context.Context
	Completion       chan bool
	// This specific reference to ConfigMapSubscription is for updating pod annotations
	ConfigMapSubscriptRef *ConfigMapSubscription
}

func (p *PodSubscription) applyConfigMapChanges(pod *v1.Pod, event watch.EventType) {

	if p.ConfigMapSubscriptRef != nil {
		if p.ConfigMapSubscriptRef.PlatformConfig != nil {
			updatedPod := pod.DeepCopy()
			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string)
			}
			// Loop through and apply
			for _, annotation := range p.ConfigMapSubscriptRef.PlatformConfig.Annotations {
				updatedPod.Annotations[annotation.Name] = annotation.Value
			}
			// Update the pod
			_, err := p.ClientSet.CoreV1().Pods(pod.Namespace).Update(p.Ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				log.Error(err)
			}
		}
	}
}
func (p *PodSubscription) Reconcile(object runtime.Object, event watch.EventType) {
	rootSpan := opentracing.GlobalTracer().StartSpan("podSubscriptionReoncile")
	defer rootSpan.Finish()

	pod := object.(*v1.Pod)
	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
	}).Info(fmt.Sprintf("PodSubscription event type %s for %s", event, pod.Name))
	switch event {
	case watch.Added:
		watchEventAdd := opentracing.GlobalTracer().StartSpan(
			"watchEventAdd", opentracing.ChildOf(rootSpan.Context()),
		)
		defer watchEventAdd.Finish()
		// Fetch the required PlatformConfig annotations
		p.applyConfigMapChanges(pod, event)
	case watch.Deleted:
	case watch.Modified:
		watchEventModified := opentracing.GlobalTracer().StartSpan(
			"watchEventModified", opentracing.ChildOf(rootSpan.Context()),
		)
		defer watchEventModified.Finish()
		p.applyConfigMapChanges(pod, event)
	}
}

func (p *PodSubscription) Subscribe() (watch.Interface, error) {
	var err error
	p.watcherInterface, err = p.ClientSet.CoreV1().Pods("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	log.Info("Started watch stream for PodSubscription")
	return p.watcherInterface, nil
}

func (p *PodSubscription) IsComplete() <-chan bool {

	return p.Completion
}
