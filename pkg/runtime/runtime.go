package runtime

import (
	"github.com/cloud-native-skunkworks/cnskunkworks-operator/pkg/subscription"
	"sync"
)

func RunLoop(subscriptions []subscription.ISubscription) error {

	var wg sync.WaitGroup
	for _, sub := range subscriptions {
		wg.Add(1)
		go func(subscription subscription.ISubscription) error {
			wiface, err := subscription.Subscribe()
			if err != nil {
				return err
			}
			for {
				select {
				case msg := <-wiface.ResultChan():
					subscription.Reconcile(msg.Object, msg.Type)
				}
			}
		}(sub)
	}
	wg.Wait()
	return nil
}
