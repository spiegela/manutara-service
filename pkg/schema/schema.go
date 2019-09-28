package schema

import (
	"strings"

	"github.com/graphql-go/graphql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/manutara/manutara/client/api"
)

const (
	namespaceLabelKey = "api.manutara.org/serviceNamespace"
	nameLabelKey      = "api.manutara.org/serviceName"
)

func labelSelector(namespace string, name string) string {
	labels := []string{
		namespaceLabelKey + "=" + namespace,
		nameLabelKey + "=" + name,
	}
	return strings.Join(labels, ",")
}

// ServiceSchema generates a GraphQL schema for a defined manutara service
type ServiceSchema struct {
	Client    api.Client
	listOpts  metav1.ListOptions
	namespace string
}

// Generate creates a GraphQL schema for a given service Namespace and Name
func (s *ServiceSchema) Generate(namespace string, name string) (graphql.Schema, error) {
	selector := labelSelector(namespace, name)
	s.namespace = namespace
	s.listOpts = metav1.ListOptions{LabelSelector: selector}
	mutation, err := s.rootField("RootMutation",
		"Root field containing mutation fields", mutationFields)
	if err != nil {
		return graphql.Schema{}, err
	}
	query, err := s.rootField("RootQuery",
		"Root field containing query fields", queryFields)
	if err != nil {
		return graphql.Schema{}, err
	}
	subscription, err := s.rootField("RootQuery",
		"Root field containing query fields", subscriptionFields)
	if err != nil {
		return graphql.Schema{}, err
	}
	return graphql.NewSchema(graphql.SchemaConfig{
		Mutation:     mutation,
		Query:        query,
		Subscription: subscription,
	})
}

func (s *ServiceSchema) rootSubscription() (*graphql.Object, error) {
	return &graphql.Object{}, nil
}

func (s *ServiceSchema) rootQuery() (*graphql.Object, error) {
	return &graphql.Object{}, nil
}

type fieldListFn func(*ServiceSchema) (graphql.Fields, error)

func (s *ServiceSchema) rootField(name string, description string, listFn fieldListFn) (*graphql.Object, error) {
	fields, err := listFn(s)
	if err != nil {
		return nil, err
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:        name,
		Description: description,
		Fields:      fields,
	}), nil
}

func mutationFields(s *ServiceSchema) (graphql.Fields, error) {
	list, err := s.Client.Mutations(s.namespace).List(s.listOpts)
	if err != nil {
		return nil, err
	}
	fields := graphql.Fields{}
	for _, mut := range list.Items {
		fields[mut.Spec.FieldName] = mut.Field()
	}
	return fields, nil
}

func queryFields(s *ServiceSchema) (graphql.Fields, error) {
	list, err := s.Client.Queries(s.namespace).List(s.listOpts)
	if err != nil {
		return nil, err
	}
	fields := graphql.Fields{}
	for _, query := range list.Items {
		fields[query.Spec.Name] = query.Field()
	}
	return fields, nil
}

func subscriptionFields(s *ServiceSchema) (graphql.Fields, error) {
	list, err := s.Client.Subscriptions(s.namespace).List(s.listOpts)
	if err != nil {
		return nil, err
	}
	fields := graphql.Fields{}
	for _, subscription := range list.Items {
		fields[subscription.Spec.Name] = subscription.Field()
	}
	return fields, nil
}
