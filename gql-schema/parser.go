package gqlschema

import (
	"fmt"
	"regexp"
	"strings"

	mschema "github.com/kinvey/odata-schema/mediation-schema"
	"github.com/kinvey/odata-schema/utils"
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

func getName(def mschema.Type) string {
	switch def.Kind {
	default:
		panic("should not happen")
	case "EntityType":
		return def.EntityType.Name
	case "Structure":
		return def.Structure.Name
	case "Enum":
		return def.Enum.Name
	}
}

func getTypeName(typeName string, types map[string]mschema.Type) string {
	return getName(types[typeName])
}

func typeToCollection(typeName string) string {
	return fmt.Sprintf("[%s]", typeName)
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
		} else {
			field.Type = getTypeName(prop.ValueType, service.Types)
		}

		if prop.IsCollection {
			field.Type = typeToCollection(field.Type)
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

func createQueryFields(collection *mschema.Collection, service *mschema.Service) []Function {
	entityTypeName := getName(service.Types[collection.EntityType])
	fields := []Function{
		{
			ReturnType: entityTypeName,
			Arguments: []Field{
				{
					Type:     "ID",
					Required: true,
					Element:  Element{Name: "id"},
				},
			},
			Element: Element{
				Name: utils.LowerFirstLetter(entityTypeName),
			},
		},
		{
			ReturnType: typeToCollection(entityTypeName),
			Arguments: []Field{
				{
					Type:    "String",
					Element: Element{Name: "filter"},
				},
				{
					Type:    "String",
					Element: Element{Name: "sort"},
				},
			},
			Element: Element{
				Name: fmt.Sprintf("%ss", utils.LowerFirstLetter(entityTypeName)),
			},
		},
	}

	return fields
}

func createMutationFields(collection *mschema.Collection, service *mschema.Service) []Function {
	entityTypeName := getName(service.Types[collection.EntityType])
	fields := []Function{
		{
			ReturnType: entityTypeName,
			Arguments: []Field{
				{
					Type:     fmt.Sprintf("%sInput", entityTypeName),
					Required: true,
					Element:  Element{Name: "data"},
				},
			},
			Element: Element{
				Name: fmt.Sprintf("add%s", utils.UpperFirstLetter(entityTypeName)),
			},
		},
		{
			ReturnType: "Boolean",
			Arguments: []Field{
				{
					Type:     "ID",
					Required: true,
					Element:  Element{Name: "id"},
				},
				{
					Type:     fmt.Sprintf("%sInput", entityTypeName),
					Required: true,
					Element:  Element{Name: "data"},
				},
			},
			Element: Element{
				Name: fmt.Sprintf("update%s", utils.UpperFirstLetter(entityTypeName)),
			},
		},
		{
			ReturnType: "Boolean",
			Arguments: []Field{
				{
					Type:     "ID",
					Required: true,
					Element:  Element{Name: "id"},
				},
			},
			Element: Element{
				Name: fmt.Sprintf("remove%s", utils.UpperFirstLetter(entityTypeName)),
			},
		},
	}

	return fields
}

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

func typeDefToDefinition(service *mschema.Service) []Definition {
	gqlTypes := []Definition{}

	for _, ttype := range service.Types {
		switch ttype.Kind {
		case "EntityType":
			{
				typeDef, inputDef := entityTypeToDefinition(ttype.EntityType, service)
				gqlTypes = append(gqlTypes, typeDef, inputDef)
			}
		case "Structure":
			{
				typeDef := createDefinition(ttype.Structure.Name, ttype.Structure.Properties, service)
				gqlTypes = append(gqlTypes, typeDef)
			}
		case "Enum":
			{
				enumDef := enumToDefinition(ttype.Enum)
				gqlTypes = append(gqlTypes, enumDef)
			}
		}
	}

	return gqlTypes
}

func Parse(service *mschema.Service) string {
	schema := Schema{
		Query: Definition{
			Type:      "type",
			Fields:    &[]Field{},
			Functions: &[]Function{},
			Elements:  &[]Element{},
			Element:   Element{Name: "Query"},
		},
		Mutation: Definition{
			Type:      "type",
			Fields:    &[]Field{},
			Functions: &[]Function{},
			Element:   Element{Name: "Mutation"},
		},
		Types: []Definition{},
		DirectiveDeclarations: []DirectiveDeclaration{
			{
				Applications: []string{"OBJECT", "FIELD_DEFINITION"},
				Directive: Directive{
					Name: "backend",
					Fields: []Field{
						{
							Type:    "String",
							Element: Element{Name: "product"},
						},
						{
							Type:    "String",
							Element: Element{Name: "collection"},
						},
						{
							Type:    "String",
							Element: Element{Name: "method"},
						},
						{
							Type:    "String",
							Element: Element{Name: "endpoint"},
						},
					},
				},
			},
			{
				Applications: []string{"FIELD_DEFINITION"},
				Directive: Directive{
					Name: "connection",
					Fields: []Field{
						{
							Type:    "String",
							Element: Element{Name: "primaryKey"},
						},
						{
							Type:    "String",
							Element: Element{Name: "foreignKey"},
						},
					},
				},
			},
		},
	}

	schema.Types = typeDefToDefinition(service)

	for _, collection := range service.Collections {
		queryFields := createQueryFields(&collection, service)
		queryFuncs := append(*schema.Query.Functions, queryFields...)
		schema.Query.Functions = &queryFuncs

		mutationFields := createMutationFields(&collection, service)
		mutationFuncs := append(*schema.Mutation.Functions, mutationFields...)
		schema.Mutation.Functions = &mutationFuncs
	}

	return schema.String()
}
