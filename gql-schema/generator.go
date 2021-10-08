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

func typeToArray(typeName string) string {
	return fmt.Sprintf("[%s]", typeName)
}

func propertyToFieldType(prop mschema.Property, types map[string]mschema.Type) string {
	var fieldType string
	if prop.PropertyType == "primitive" {
		fieldType = strings.Title(prop.ValueType)
		if strings.HasPrefix(fieldType, "Int") || strings.HasPrefix(fieldType, "Float") {
			fieldType = replaceLastDigitsRegexp.ReplaceAllString(fieldType, "")
		}
		if fieldType == "Datetime" {
			fieldType = "String" // TODO: Maybe create a custom GQL type?
		} else if fieldType == "Decimal" {
			fieldType = "Float"
		}
	} else {
		fieldType = getTypeName(prop.ValueType, types)
	}

	if prop.IsCollection {
		fieldType = typeToArray(fieldType)
	}

	return fieldType
}

func propToField(propName string, prop mschema.Property, types map[string]mschema.Type) Field {
	field := Field{
		Type:     propertyToFieldType(prop, types),
		Required: prop.Required,
		Element: Element{
			Name: propName,
		},
	}

	return field
}

func propsToFields(properties map[string]mschema.Property, types map[string]mschema.Type) *[]Field {
	fields := make([]Field, len(properties))
	i := 0
	for propName, prop := range properties {
		fields[i] = propToField(propName, prop, types)
		i += 1
	}

	return &fields
}

func createDefinition(name string, properties map[string]mschema.Property, types map[string]mschema.Type) Definition {
	def := Definition{
		Type:   "type",
		Fields: propsToFields(properties, types),
		Element: Element{
			Name: name,
		},
	}

	return def
}

func createInputType(entityType *mschema.EntityType, types map[string]mschema.Type) Definition {
	typeDef := createDefinition(entityType.Name, entityType.Properties, types)
	typeDef.Type = "input"
	typeDef.Name = fmt.Sprintf("%sInput", typeDef.Name)
	fields := []Field{}
	for _, field := range *typeDef.Fields {
		isStructuralProp := entityType.Properties[field.Name].PropertyType != "relation"
		if field.Type != "ID" && isStructuralProp {
			field.Required = false
			fields = append(fields, field)
		}
	}
	typeDef.Fields = &fields
	return typeDef
}

func findCollectionForType(entityTypeName string, collections map[string]mschema.Collection) (string, bool) {
	for name, collection := range collections {
		if collection.EntityType == entityTypeName {
			return name, true
		}
	}
	return "", false
}

func entityTypeToDefinition(entityTypeName string, service *mschema.Service) (Definition, Definition) {
	entityType := service.Types[entityTypeName].EntityType
	typeDef := createDefinition(entityType.Name, entityType.Properties, service.Types)

	if collectionForType, found := findCollectionForType(entityTypeName, service.Collections); found {
		typeDef.Directives = &[]Directive{newBackendDirective("sitefinity", collectionForType, "", "")}
	}
	addKey(entityType, typeDef.Fields)
	inputDef := createInputType(entityType, service.Types)
	return typeDef, inputDef
}

func createQueryFields(collection *mschema.Collection, service *mschema.Service) []Field {
	entityTypeName := getName(service.Types[collection.EntityType])
	fields := []Field{
		{
			Type: entityTypeName,
			Arguments: &[]Field{
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
			Type: typeToArray(entityTypeName),
			Arguments: &[]Field{
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

func createMutationFields(collection *mschema.Collection, service *mschema.Service) []Field {
	entityTypeName := getName(service.Types[collection.EntityType])
	fields := []Field{
		{
			Type: entityTypeName,
			Arguments: &[]Field{
				{
					Type:     fmt.Sprintf("%sInput", entityTypeName),
					Required: true,
					Element:  Element{Name: "data"},
				},
			},
			Element: Element{
				Name:       fmt.Sprintf("add%s", utils.UpperFirstLetter(entityTypeName)),
				Directives: &[]Directive{newBackendDirective("sitefinity", collection.Name, "POST", "")},
			},
		},
		{
			Type: "Boolean",
			Arguments: &[]Field{
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
				Name:       fmt.Sprintf("update%s", utils.UpperFirstLetter(entityTypeName)),
				Directives: &[]Directive{newBackendDirective("sitefinity", collection.Name, "PATCH", "")},
			},
		},
		{
			Type: "Boolean",
			Arguments: &[]Field{
				{
					Type:     "ID",
					Required: true,
					Element:  Element{Name: "id"},
				},
			},
			Element: Element{
				Name:       fmt.Sprintf("remove%s", utils.UpperFirstLetter(entityTypeName)),
				Directives: &[]Directive{newBackendDirective("sitefinity", collection.Name, "DELETE", "")},
			},
		},
	}

	return fields
}

func enumMembersToFields(members map[string]string, membersType string) *[]Field {
	elements := []Field{}
	for memberName := range members {
		elements = append(elements, Field{Element: Element{Name: memberName}})
	}
	return &elements
}

func enumToDefinition(enum *mschema.Enum) Definition {
	fields := enumMembersToFields(enum.Members, enum.ValuesType)
	return Definition{
		Type:    "enum",
		Element: Element{Name: enum.Name},
		Fields:  fields,
	}
}

func typeDefToDefinition(service *mschema.Service) []Definition {
	gqlTypes := []Definition{}
	addedTypes := make(map[string]usedTypeDesc)
	var gqlTypeDef Definition
	var inputDef Definition

	for name, typeDef := range service.Types {
		switch typeDef.Kind {
		case "EntityType":
			gqlTypeDef, inputDef = entityTypeToDefinition(name, service)
			gqlTypes = append(gqlTypes, inputDef)
		case "Structure":
			gqlTypeDef = createDefinition(typeDef.Structure.Name, typeDef.Structure.Properties, service.Types)
		case "Enum":
			gqlTypeDef = enumToDefinition(typeDef.Enum)
		}
		if addedType, ok := addedTypes[gqlTypeDef.Name]; !ok {
			addedTypes[gqlTypeDef.Name] = usedTypeDesc{
				QualifiedName: name,
				MedSchemaType: typeDef,
				GqlField:      gqlTypeDef,
			}
		} else {
			fmt.Printf("A type was already added for type '%s'. Attempting to add type '%s'. The existing type is '%s'\n", gqlTypeDef.Name, name, addedType.QualifiedName)
		}
		// TODO: these are the types which are not actually referenced
		if name != "Telerik.Sitefinity.Web.Api.OData.Operations.Media.Models.ThumbnailModel" && name != "Telerik.Sitefinity.Folder" {
			gqlTypes = append(gqlTypes, gqlTypeDef)
		}
	}

	return gqlTypes
}

// func invocationToMutation(invocation mschema.Invocation, types map[string]mschema.Type) Field {
// 	retType := "System__Void"

// 	if invocation.Result != nil {
// 		retType = propertyToFieldType(*invocation.Result, types)
// 	}

// 	field := Field{
// 		Type:      retType,
// 		Arguments: &[]Field{},
// 		Element:   Element{Name: utils.LowerFirstLetter(invocation.Name)},
// 	}

// 	if len(invocation.Arguments) > 0 {
// 		fieldArgs := []Field{}

// 		for _, arg := range invocation.Arguments {
// 			fieldArgs = append(fieldArgs, propToField(arg.Name, arg.Property, types))
// 		}

// 		field.Arguments = &fieldArgs
// 	}

// 	return field
// }

// func invocationsToMutations(invocations map[string]mschema.Invocation, types map[string]mschema.Type) []Field {
// 	result := []Field{}
// 	for _, invocation := range invocations {
// 		result = append(result, invocationToMutation(invocation, types))
// 	}
// 	return result
// }

func Generate(service *mschema.Service) string {
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
		Types: []Definition{
			{
				Type:    "type",
				Element: Element{Name: "System__Void"},
			},
		},
		DirectiveDeclarations: []DirectiveDeclaration{
			{
				Applications: []string{"OBJECT", "FIELD_DEFINITION"},
				Directive:    newBackendDirective("String", "String", "String", "String"),
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

	schema.Types = append(schema.Types, typeDefToDefinition(service)...)

	// TODO: better way to append or not use pointer?
	for _, collection := range service.Collections {
		queryFields := createQueryFields(&collection, service)
		queryFuncs := append(*schema.Query.Fields, queryFields...)
		schema.Query.Fields = &queryFuncs

		mutationFields := createMutationFields(&collection, service)
		mutationFuncs := append(*schema.Mutation.Fields, mutationFields...)
		schema.Mutation.Fields = &mutationFuncs
	}

	// TODO: backend doesn't support these
	// invocations := invocationsToMutations(service.Invocations, service.Types)
	// asMutationFields := append(*schema.Mutation.Fields, invocations...)
	// schema.Mutation.Fields = &asMutationFields

	return schema.String()
}
