package service

import (
	"net/http"

	"github.com/fvbock/endless"
	"github.com/graphql-go/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/manutara/manutara/client/api"
	"github.com/manutara/service/pkg/schema"
)

type Daemon struct {
	Name              string
	Namespace         string
	BindAddr          string
	MetricsAddr       string
	PlaygroundEnabled bool
	PrettyEnabled     bool
	SchemaUpdates     chan schema.ElementType
	Client            api.Client
}

// Start creates an HTTP listener for the GraphQL service
func (d *Daemon) Start() error {
	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName(d.Name)
	serviceSchema, err := schema.ServiceSchema{
		Client: d.Client,
	}.Generate(d.Namespace, d.Name)
	if err != nil {
		return err
	}
	http.Handle("/graphql", handler.New(&handler.Config{
		Schema:     &serviceSchema,
		Playground: d.PlaygroundEnabled,
		Pretty:     d.PrettyEnabled,
	}))
	log.Info("GraphQL server starting",
		"name", d.Name,
		"bind address", d.BindAddr)
	return endless.ListenAndServe(d.BindAddr, nil)
}
