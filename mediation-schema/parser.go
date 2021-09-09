package mediationschema

import (
	"fmt"
	"strings"

	ods "github.com/kinvey/odata-schema/odata-schema"
)

const collectionPrefix = "Collection("

type edmObjects struct {
	entityTypes     map[string]*ods.EntityType
	complexTypes    map[string]*ods.ComplexType
	enumTypes       map[string]*ods.EnumType
	entityContainer *ods.EntityContainer
}

func formQualifiedName(schema *ods.Schema, objectName string) (string, string) {
	namespacedName := fmt.Sprintf("%s.%s", schema.Namespace, objectName)
	aliasedName := ""

	if schema.Alias != nil {
		aliasedName = fmt.Sprintf("%s.%s", *schema.Alias, objectName)
	}

	return namespacedName, aliasedName
}

func addToEntityTypes(objects edmObjects, schema *ods.Schema, entityType ods.EntityType) {
	namespacedName, aliasedName := formQualifiedName(schema, entityType.Name)
	objects.entityTypes[namespacedName] = &entityType
	if aliasedName != "" {
		objects.entityTypes[aliasedName] = &entityType
	}
}

func addToComplexTypes(objects edmObjects, schema *ods.Schema, complexType ods.ComplexType) {
	namespacedName, aliasedName := formQualifiedName(schema, complexType.Name)
	objects.complexTypes[namespacedName] = &complexType
	if aliasedName != "" {
		objects.complexTypes[aliasedName] = &complexType
	}
}

func addToEnumTypes(objects edmObjects, schema *ods.Schema, enumtype ods.EnumType) {
	namespacedName, aliasedName := formQualifiedName(schema, enumtype.Name)
	objects.enumTypes[namespacedName] = &enumtype
	if aliasedName != "" {
		objects.enumTypes[aliasedName] = &enumtype
	}
}

func getBindingTarget(entitySet *ods.EntitySet, navPropertyName string) (string, bool) {
	for _, navBindings := range entitySet.NavigationPropertyBindings {
		if navBindings.Path == navPropertyName {
			return navBindings.Target, true
		}
	}
	return "", false
}

func mapEdmType(edmType string) (string, error) {
	switch edmType {
	case "Edm.String", "Edm.Guid":
		return "string", nil
	case "Edm.Boolean":
		return "boolean", nil
	case "Edm.DateTime": // TODO: consolidate somehow
		return "datetime", nil
	case "Edm.DateTimeOffset":
		return "datetime", nil
	case "Edm.Single":
		return "float32", nil
	case "Edm.Double":
		return "float64", nil
	case "Edm.Int16":
		return "int16", nil
	case "Edm.Int32":
		return "int32", nil
	case "Edm.Int":
		return "int32", nil
	case "Edm.Int64":
		return "int64", nil
	case "Edm.Decimal":
		return "decimal", nil
	case "Edm.Binary":
		return "binary", nil
	case "Edm.Stream":
		return "stream", nil
	case "Edm.GeographyPoint":
		return "geopoint", nil
	default:
		if strings.HasPrefix(edmType, "Edm.") {
			panic(edmType)
		}
		return "", fmt.Errorf("unknown Type: %s", edmType)
	}
}

func mapType(typeName string, objects *edmObjects) Property {
	result := Property{
		PropertyType: "unknown",
		Type:         fmt.Sprintf("unknown %s", typeName),
		IsCollection: false,
	}
	if mappedType, err := mapEdmType(typeName); err == nil {
		result.PropertyType = "primitive"
		result.Type = mappedType
	} else if strings.HasPrefix(typeName, collectionPrefix) {
		actualType := typeName[len(collectionPrefix) : len(typeName)-1]
		mapped := mapType(actualType, objects)
		mapped.IsCollection = true
		result = mapped
	} else if entityType, ok := objects.entityTypes[typeName]; ok {
		result.Type = entityType.Name
		result.PropertyType = "relation"
	} else if complexType, ok := objects.complexTypes[typeName]; ok {
		result.PropertyType = "structure"
		result.Type = complexType.Name
	} else if enumType, ok := objects.enumTypes[typeName]; ok {
		result.PropertyType = "enum"
		result.Type = enumType.Name
	}
	return result
}

func getTypeProperties(typeName string, objects *edmObjects) []ods.Property {
	var properties []ods.Property
	var baseType *string
	if foundType, ok := objects.entityTypes[typeName]; ok {
		properties = foundType.Properties
		baseType = foundType.BaseType
	} else if foundType, ok := objects.complexTypes[typeName]; ok {
		properties = foundType.Properties
		baseType = foundType.BaseType
	}
	if baseType != nil {
		baseProps := getTypeProperties(*baseType, objects)
		properties = append(properties, baseProps...)
	}
	return properties
}

func getTypeNavProperties(typeName string, objects *edmObjects) []ods.NavigationProperty {
	var properties []ods.NavigationProperty
	var baseType *string
	if foundType, ok := objects.entityTypes[typeName]; ok {
		properties = foundType.NavigationProperties
		baseType = foundType.BaseType
	} else if foundType, ok := objects.complexTypes[typeName]; ok {
		properties = foundType.NavigationProperties
		baseType = foundType.BaseType
	}
	if baseType != nil {
		baseProps := getTypeNavProperties(*baseType, objects)
		properties = append(properties, baseProps...)
	}
	return properties
}

func addStructuralProperties(typeName string, objects *edmObjects, result map[string]EntityProperty) {
	properties := getTypeProperties(typeName, objects)
	for _, property := range properties {
		prop := EntityProperty{
			Property: mapType(property.Type, objects),
		}
		// if ref := getStructureName(property.Type, objects); ref != nil {
		// 	prop.TypeRef = ref
		// }
		result[property.Name] = prop
	}
}

func addNavProperties(entitySet *ods.EntitySet, objects *edmObjects, result map[string]EntityProperty) {
	properties := getTypeNavProperties(entitySet.EntityType, objects)

	for _, property := range properties {
		prop := EntityProperty{
			Property: mapType(property.Type, objects),
		}

		if bindingTarget, found := getBindingTarget(entitySet, property.Name); found {
			prop.RelatedTo = &bindingTarget
		}

		result[property.Name] = prop
	}

	if len(entitySet.NavigationPropertyBindings) != len(properties) {
		fmt.Printf("Nav property count mismatch. EntitySet: %s, EntitySet count: %d, NavProp count: %d\n", entitySet.Name, len(entitySet.NavigationPropertyBindings), len(properties))
	}
}

// func getStructureName(propType string, objects *edmObjects) *string {
// 	if complexType, ok := objects.complexTypes[propType]; ok {
// 		return &complexType.Name
// 	} else if enumType, ok := objects.enumTypes[propType]; ok {
// 		return &enumType.Name
// 	}
// 	return nil
// }

func mapEntitySet(entitySet *ods.EntitySet, objects *edmObjects) *Collection {
	entityType := objects.entityTypes[entitySet.EntityType]
	keys := make([]string, len(*entityType.Key))

	for i, keyRef := range *entityType.Key {
		keys[i] = keyRef.Name
	}

	collection := Collection{
		Key:       keys,
		HasStream: entityType.HasStream,
		Structure: Structure{
			Name:       entitySet.Name,
			Properties: make(map[string]EntityProperty),
		},
	}

	addStructuralProperties(entitySet.EntityType, objects, collection.Properties)
	addNavProperties(entitySet, objects, collection.Properties)

	return &collection
}

func mapComplexType(complexTypeName string, complexType *ods.ComplexType, objects *edmObjects) *Structure {
	cType := Structure{
		Name:       complexType.Name,
		Properties: make(map[string]EntityProperty),
	}

	addStructuralProperties(complexTypeName, objects, cType.Properties)
	return &cType
}

func mapEnumType(enumName string, enum *ods.EnumType, objects *edmObjects) *Enum {
	eType := Enum{
		Name:       enum.Name,
		Members:    make(map[string]string),
		ValuesType: enum.UnderlyingType,
		IsFlags:    enum.IsFlags,
	}

	if eType.ValuesType == "" {
		mappedType, _ := mapEdmType("Edm.Int32") // Specified in MC-CSDL
		eType.ValuesType = mappedType
	}

	for _, member := range enum.Members {
		eType.Members[member.Name] = member.Value
	}

	return &eType
}

func extractObjects(edm *ods.EdmxDocument) *edmObjects {
	objects := edmObjects{
		entityTypes:     make(map[string]*ods.EntityType),
		complexTypes:    make(map[string]*ods.ComplexType),
		enumTypes:       make(map[string]*ods.EnumType),
		entityContainer: nil,
	}

	for _, schema := range edm.DataServices.Schemas {
		if schema.EntityContainer != nil {
			objects.entityContainer = schema.EntityContainer
		}
		for _, entityType := range schema.EntityTypes {
			addToEntityTypes(objects, &schema, entityType)
		}
		for _, complexType := range schema.ComplexTypes {
			addToComplexTypes(objects, &schema, complexType)
		}
		for _, enumType := range schema.EnumTypes {
			addToEnumTypes(objects, &schema, enumType)
		}
	}

	return &objects
}

func Parse(edm *ods.EdmxDocument) (Service, error) {
	schema := Service{
		// TODO: use make() with appropriate sizes
		Collections: []Collection{},
		Endpoints:   []Endpoint{},
		Structures:  []Structure{},
		Enums:       []Enum{},
		Invocations: []Invocation{},
	}

	objects := extractObjects(edm)

	for _, entitySet := range objects.entityContainer.EntitySets {
		collection := mapEntitySet(&entitySet, objects)
		schema.Collections = append(schema.Collections, *collection)
	}

	for typeName, complexType := range objects.complexTypes {
		schema.Structures = append(schema.Structures, *mapComplexType(typeName, complexType, objects))
	}

	for enumName, enum := range objects.enumTypes {
		schema.Enums = append(schema.Enums, *mapEnumType(enumName, enum, objects))
	}

	schema.Name = objects.entityContainer.Name
	return schema, nil
}
