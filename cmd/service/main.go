package main

import (
	"flag"

	"github.com/manutara/service/pkg/service"
	"github.com/manutara/service/schema"
)

func main() {
	var (
		name              string
		namespace         string
		bindAddr          string
		metricsAddr       string
		playgroundEnabled bool
		apiHost           string
		kubeconfig        string
	)

	flag.StringVar(&name, "name", "",
		"The name of this service daemon")
	flag.StringVar(&namespace, "namespace", "",
		"The namespace of this service daemon")
	flag.StringVar(&bindAddr, "bind-addr", ":8080",
		"The address to which the service endpoint binds.")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8081",
		"The address to which the metrics endpoint binds.")
	flag.BoolVar(&playgroundEnabled, "playground", true,
		"Indicates that the service endpoint should have playground enabled")
	flag.StringVar(&apiHost, "a", "",
		"Kubernetes API server host when connecting to a cluster externally")
	flag.StringVar(&kubeconfig, "c", "",
		"File path to kubeconfig file when connecting to a cluster externally")

	// Create K8s watcher for Manutara schema entries
	schemaWatcher := schema.Watcher{
		APIHost:       apiHost,
		Config:        kubeconfig,
	}
	schemaUpdates, err := schemaWatcher.Watch(namespace, name)
	if err != nil {
		panic(err)
	}

	// Start GraphQL HTTP Endpoint
	err = service.Daemon{
		Name:              name,
		BindAddr:          bindAddr,
		MetricsAddr:       metricsAddr,
		PlaygroundEnabled: playgroundEnabled,
		SchemaUpdates:     schemaUpdates,
	}.Start()
	if err != nil {
		panic(err)
	}
}
