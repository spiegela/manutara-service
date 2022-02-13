package schema

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/manutara/manutara/api/v1"
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
	Client       api.Client
	listOpts     metav1.ListOptions
	namespace    string
	graphQLTypes map[string]graphql.Type
	rawDataTypes []v1.DataType
}

// Generate creates a GraphQL schema for a given service Namespace and Name
func (s *ServiceSchema) Generate(namespace string, name string) (graphql.Schema, error) {
	selector := labelSelector(namespace, name)
	s.namespace = namespace
	s.listOpts = metav1.ListOptions{LabelSelector: selector}

	err := s.GenerateDataTypes()
	if err != nil {
		return graphql.Schema{}, err
	}
	fmt.Printf("%+v\n", s.graphQLTypes)

	return graphql.NewSchema(graphql.SchemaConfig{
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

func (s *ServiceSchema) GenerateDataTypes() error {
	dataTypeList, err := s.Client.DataTypes(s.namespace).List(s.listOpts)
	if err != nil {
		return err
	}
	s.rawDataTypes = dataTypeList.Items
	for _, dataType := range dataTypeList.Items {
		if _, ok := s.graphQLTypes[dataType.Name]; ok {
			// skip if already defined, since elements can be generated due
			// to dependencies
			continue
		}
		graphQLType, err := s.asGraphQLType(dataType)
		if err != nil {
			return err
		}
		s.graphQLTypes[dataType.Name] = graphQLType
	}
	return nil
}

func (s *ServiceSchema) asGraphQLType(dataType v1.DataType) (graphql.Type, error) {
	fields, err := s.asGraphQLFields(dataType.Spec.Fields)
	if err != nil {
		return nil, err
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:        dataType.Name,
		Description: dataType.Spec.Description,
		Fields:      fields,
	}), nil
}

func (s *ServiceSchema) asGraphQLFields(dataTypeFields v1.DataTypeFields) (graphql.Fields, error) {
	fields := map[string]*graphql.Field{}
	for fieldName, dataTypeField := range dataTypeFields {
		fieldDataType, err := s.asGraphQLFieldType(dataTypeField)
		if err != nil {
			return nil, err
		}
		fields[fieldName] = &graphql.Field{
			Description: dataTypeField.Description,
			Type:        fieldDataType,
		}
	}
	return fields, nil
}

func (s *ServiceSchema) asGraphQLFieldType(dataTypeField v1.DataTypeField) (graphql.Type, error) {
	switch {
	case dataTypeField.BasicType != "":
		return asGraphQLBasicType(dataTypeField)
	case dataTypeField.UserDefinedType != "":
		if graphQLTypeDep, ok := s.graphQLTypes[dataTypeField.UserDefinedType]; ok {
			return graphQLTypeDep, nil
		}
		for _, dataType := range s.rawDataTypes {
			if dataType.Name == dataTypeField.UserDefinedType {
				graphQLTypeDep, err := s.asGraphQLType(dataType)
				if err == nil {
					return nil, err
				}
				s.graphQLTypes[dataType.Name] = graphQLTypeDep
			}
		}
		return nil, fmt.Errorf("unable to find dependent type: %s", dataTypeField.UserDefinedType)
	}
	return nil, errors.New("unable to detect field type")
}

func asGraphQLBasicType(dataTypeField v1.DataTypeField) (graphql.Type, error) {
	switch dataTypeField.BasicType {
	case v1.DataTypeIDField:
		return graphql.ID, nil
	case v1.DataTypeStringField:
		return graphql.String, nil
	case v1.DataTypeIntField:
		return graphql.Int, nil
	case v1.DataTypeFloatField:
		return graphql.Float, nil
	case v1.DataTypeBooleanField:
		return graphql.Boolean, nil
	case v1.DataTypeDateField:
		return graphql.DateTime, nil
	}
	return nil, errors.New("basic field type unknown")
}
