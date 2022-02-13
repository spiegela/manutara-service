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
// Field, or Subscription
type ElementType int

// Enum constants for ElementType
const (
	ElementTypeDataType ElementType = iota
	ElementTypeField
)

// Watch subscribes to updates for all schema elements which could result in the
// update of a schema
func (w *Watcher) Watch(namespace string, name string) (chan ElementType, error) {
	w.updates = make(chan ElementType)
	opts := metav1.ListOptions{LabelSelector: labelSelector(namespace, name)}
	var wg sync.WaitGroup
	wg.Add(3)

	typeWatcher, err := client.DataTypes(w.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	go func(dataTypeChan <-chan watch.Event) {
		for _ = range dataTypeChan {
			w.updates <- ElementTypeDataType
		}
		wg.Done()
	}(typeWatcher.ResultChan())

	fieldWatcher, err := client.DataTypes(w.Namespace).Watch(opts)
	if err != nil {
		return nil, err
	}
	go func(mutationsChan <-chan watch.Event) {
		for _ = range mutationsChan {
			w.updates <- ElementTypeField
		}
		wg.Done()
	}(fieldWatcher.ResultChan())

	go func() {
		wg.Wait()
		close(w.updates)
	}()

	return w.updates, nil
}
