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

type PodSubscription struct {
	watcherInterface watch.Interface
	ClientSet        kubernetes.Interface
	Ctx              context.Context
	Completion       chan bool
}

func (p *PodSubscription) Reconcile(object runtime.Object, event watch.EventType) {

	pod := object.(*v1.Pod)
	klog.Infof("PodSubscription event type %s for %s", event, pod.Name)

	switch event {
	case watch.Added:
		if _, ok := pod.Annotations["type"]; !ok {
			updatedPod := pod.DeepCopy()
			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string)
			}
			updatedPod.Annotations["type"] = "sre"
			// Update the pod
			_, err := p.ClientSet.CoreV1().Pods(pod.Namespace).Update(p.Ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(err)
			}
		}

	case watch.Deleted:
	case watch.Modified:

		if pod.Annotations["type"] == "sre" {
			klog.Info("This could be some custom behaviour beyond just a CRUD")
		}

	}

}

func (p *PodSubscription) Subscribe() (watch.Interface, error) {
	var err error
	p.watcherInterface, err = p.ClientSet.CoreV1().Pods("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	klog.Info("Started watch stream for PodSubscription")
	return p.watcherInterface, nil
}

func (p *PodSubscription) IsComplete() <-chan bool {

	return p.Completion
}
