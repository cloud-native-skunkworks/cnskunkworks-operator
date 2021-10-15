package runtime

import (
	"github.com/cloud-native-skunkworks/cnskunkworks-operator/pkg/subscription"
)

func RunLoop(subscriptions []subscription.ISubscription) error{


	for _, subscription := range subscriptions {

		wiface, err := subscription.Subscribe()
		if err != nil {
			return err
		}

		go func() {

			for {
				select {
					case msg := <- wiface.ResultChan():
						subscription.Reconcile(msg.Object,msg.Type)
					// TODO: Do we want a way of escaping from our go rountines?
				}
			}

		}() // Could signal handler into them??
	}

	for _, subscription := range subscriptions {


		select {
			case  _ = <- subscription.IsComplete():
				break
		}
	}
	return nil
}