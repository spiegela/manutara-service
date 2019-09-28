package schema

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/manutara/manutara/client"
)

// Watcher combines updates to any schema elements into a single channel to
// be consumed by a client to respond to schema changes
type Watcher struct {
	APIHost   string
	Config    string
	Namespace string
	updates   chan ElementType
}

// ElementType is an enum that signifies the schema element type: Query,
// Mutation, or Subscription
type ElementType int

// Enum constants for ElementType
const (
	ElementTypeMutation ElementType = iota
	ElementTypeQuery
	ElementTypeSubscription
)

// Watch subscribes to updates for all schema elements which could result in the
// update of a schema
func (w *Watcher) Watch(namespace string, name string) (chan ElementType, error) {
	w.updates = make(chan ElementType)
	opts := metav1.ListOptions{LabelSelector: labelSelector(namespace, name)}
	var wg sync.WaitGroup
	wg.Add(3)

	mutationsWatcher, err := client.Mutations(w.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	go func(mutationsChan <-chan watch.Event) {
		for _ = range mutationsChan {
			w.updates <- ElementTypeMutation
		}
		wg.Done()
	}(mutationsWatcher.ResultChan())

	queriesWatcher, err := client.Queries(w.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	go func(queriesChan <-chan watch.Event) {
		for _ = range queriesChan {
			w.updates <- ElementTypeQuery
		}
		wg.Done()
	}(queriesWatcher.ResultChan())

	subscriptionsWatcher, err := client.Subscriptions(w.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	go func(subscriptionsChan <-chan watch.Event) {
		for _ = range subscriptionsChan {
			w.updates <- ElementTypeSubscription
		}
		wg.Done()
	}(subscriptionsWatcher.ResultChan())

	go func() {
		wg.Wait()
		close(w.updates)
	}()

	return w.updates, nil
}
