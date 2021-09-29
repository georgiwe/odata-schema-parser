package gqlschema

import (
	"fmt"
	"regexp"
	"strings"

	mschema "github.com/kinvey/odata-schema/mediation-schema"
)

var replaceLastDigitsRegexp = regexp.MustCompile(`\d+$`)

func addKey(entityType *mschema.EntityType, fieldsRef *[]Field) {
	if len(entityType.Key) > 1 {
		panic("how do we handle composite keys")
	}

	fields := *fieldsRef

	for i, field := range fields {
		if entityType.Key[0] == field.Name {
			fields[i].Type = "ID"
			fields[i].Required = false // So we omit the ! in the schema
		}
	}
}

func propsToFields(properties map[string]mschema.Property, service *mschema.Service) *[]Field {
	fields := make([]Field, len(properties))
	i := 0
	for propName, prop := range properties {
		field := Field{
			Required: prop.Required,
			Element: Element{
				Name: propName,
			},
		}
		if prop.PropertyType == "primitive" {
			field.Type = strings.Title(prop.ValueType)
			if strings.HasPrefix(field.Type, "Int") || strings.HasPrefix(field.Type, "Float") {
				field.Type = replaceLastDigitsRegexp.ReplaceAllString(field.Type, "")
			}
		} else if prop.PropertyType == "structure" {
			field.Type = service.Structures[prop.ValueType].Name
		} else if prop.PropertyType == "relation" {
			field.Type = service.EntityTypes[prop.ValueType].Name
		} else if prop.PropertyType == "enum" {
			field.Type = service.Enums[prop.ValueType].Name
		} else {
			panic("unknown type")
		}

		if prop.IsCollection {
			field.Type = fmt.Sprintf("[%s]", field.Type)
		}

		fields[i] = field
		i += 1
	}

	return &fields
}

func createDefinition(name string, properties map[string]mschema.Property, service *mschema.Service) Definition {
	def := Definition{
		Type:   "type",
		Fields: propsToFields(properties, service),
		Element: Element{
			Name: name,
		},
	}

	return def
}

func createInputType(entityType *mschema.EntityType, service *mschema.Service) Definition {
	typeDef := createDefinition(entityType.Name, entityType.Properties, service)
	typeDef.Type = "input"
	typeDef.Name = fmt.Sprintf("%sInput", typeDef.Name)
	fields := make([]Field, 0)
	for _, field := range *typeDef.Fields {
		isStructuralProp := entityType.Properties[field.Name].PropertyType != "relation"
		// isKeyProperty := entityType.Key[0] == field.Name // cant handle composite keys yet, should throw in addKey() before this step
		if field.Type != "ID" && isStructuralProp {
			fields = append(fields, field)
		}
	}
	typeDef.Fields = &fields
	return typeDef
}

// func findCollections(entityTypeQualifiedName string, service *mediationschema.Service) []mediationschema.Collection {
// 	collections := make([]mediationschema.Collection, 0)

// 	i := 0
// 	for _, collection := range service.Collections {
// 		if collection.EntityType == entityTypeQualifiedName {
// 			collections[i] = collection
// 		}
// 		i += 1
// 	}

// 	return collections
// }

func entityTypeToDefinition(entityType *mschema.EntityType, service *mschema.Service) (Definition, Definition) {
	typeDef := createDefinition(entityType.Name, entityType.Properties, service)
	addKey(entityType, typeDef.Fields)
	inputDef := createInputType(entityType, service)
	return typeDef, inputDef
}

// func createQueryFields(collection *mschema.Collection, service) []Field {
// 	collection.
// }

// func createMutationFields(collection *mschema.Collection) []Field {

// }

func enumMembersToFields(members map[string]string, membersType string) *[]Element {
	elements := []Element{}
	for memberName, member := range members {
		element := Element{
			Name: memberName,
			Directives: &[]Directive{
				{
					Name: "meta",
					Fields: []Field{
						{
							Type:    member,
							Element: Element{Name: "value"},
						},
						{
							Type:    membersType,
							Element: Element{Name: "valueType"},
						},
					},
				},
			},
		}
		elements = append(elements, element)
	}
	return &elements
}

func enumToDefinition(enum *mschema.Enum) Definition {
	elements := enumMembersToFields(enum.Members, enum.ValuesType)
	return Definition{
		Type:     "enum",
		Element:  Element{Name: enum.Name},
		Elements: elements,
	}
}

func Parse(backend *mschema.Backend) string {
	schema := Schema{
		Query: Definition{
			Type:    "type",
			Fields:  &[]Field{},
			Element: Element{Name: "Query"},
		},
		Mutation: Definition{
			Type:    "type",
			Fields:  &[]Field{},
			Element: Element{Name: "Mutation"},
		},
		Types: []Definition{},
	}

	for _, service := range backend.Services {
		for _, entityType := range service.EntityTypes {
			typeDef, inputDef := entityTypeToDefinition(&entityType, &service)
			schema.Types = append(schema.Types, typeDef, inputDef)
		}
		for _, structure := range service.Structures {
			typeDef := createDefinition(structure.Name, structure.Properties, &service)
			schema.Types = append(schema.Types, typeDef)
		}
		for _, enum := range service.Enums {
			enumDef := enumToDefinition(&enum)
			schema.Types = append(schema.Types, enumDef)
		}
	}

	// for _, collection := range service.Collections {
	// 	queryFields := createQueryFields(&collection)
	// 	fields := append(*schema.Query.Fields, queryFields...)
	// 	schema.Query.Fields = &fields

	// 	mutationFields := createMutationFields(&collection)
	// 	fields = append(*schema.Mutation.Fields, mutationFields...)
	// 	schema.Mutation.Fields = &fields
	// }

	return schema.String()
}
