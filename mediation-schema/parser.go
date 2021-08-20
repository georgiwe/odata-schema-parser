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
	case "Edm.DateTimeOffset":
		return "datetime", nil
	case "Edm.Single":
		return "float32", nil
	case "Edm.Double":
		return "float64", nil
	case "Edm.Int32":
		return "int32", nil
	case "Edm.Int64":
		return "int64", nil
	case "Edm.Decimal":
		return "decimal", nil
	default:
		if strings.HasPrefix(edmType, "Edm.") {
			panic(edmType)
		}
		return "", fmt.Errorf("unknown Type: %s", edmType)
	}
}

func mapType(edmType string, objects *edmObjects) string {
	if mappedType, err := mapEdmType(edmType); err == nil {
		return mappedType
	}
	if strings.HasPrefix(edmType, collectionPrefix) {
		actualType := edmType[len(collectionPrefix) : len(edmType)-1]
		return fmt.Sprintf("[]%s", mapType(actualType, objects))
	} else if entityType, ok := objects.entityTypes[edmType]; ok {
		return entityType.Name
	} else if _, ok := objects.complexTypes[edmType]; ok {
		return "structure"
	} else if _, ok := objects.enumTypes[edmType]; ok {
		return "structure"
	}
	return fmt.Sprintf("<unknown %s>", edmType)
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

func addStructuralProperties(typeName string, objects *edmObjects, result map[string]Property) {
	properties := getTypeProperties(typeName, objects)
	for _, property := range properties {
		prop := Property{
			Type: mapType(property.Type, objects),
		}
		if ref := getStructureName(property.Type, objects); ref != nil && typeName != property.Type {
			prop.StructureRef = ref
		}
		result[property.Name] = prop
	}
}

func addNavProperties(entitySet *ods.EntitySet, objects *edmObjects, result map[string]Property) {
	properties := getTypeNavProperties(entitySet.EntityType, objects)

	for _, property := range properties {
		propDescriptor := Property{
			Type: mapType(property.Type, objects),
		}

		if ref := getStructureName(property.Type, objects); ref != nil {
			propDescriptor.StructureRef = ref
		}

		if bindingTarget, found := getBindingTarget(entitySet, property.Name); found {
			propDescriptor.RelatedTo = &bindingTarget
		}

		result[property.Name] = propDescriptor
	}

	if len(entitySet.NavigationPropertyBindings) != len(properties) {
		fmt.Printf("Nav property count mismatch. EntitySet: %s, EntitySet count: %d, NavProp count: %d\n", entitySet.Name, len(entitySet.NavigationPropertyBindings), len(properties))
	}
}

func getStructureName(propType string, objects *edmObjects) *string {
	if complexType, ok := objects.complexTypes[propType]; ok {
		return &complexType.Name
	} else if enumType, ok := objects.enumTypes[propType]; ok {
		return &enumType.Name
	}
	return nil
}

func addComplexTypeNavProperties(typeName string, objects *edmObjects, result map[string]Property) {
	properties := getTypeNavProperties(typeName, objects)

	for _, property := range properties {
		prop := Property{
			Type: mapType(property.Type, objects),
		}
		if ref := getStructureName(property.Type, objects); ref != nil && typeName != property.Type {
			prop.StructureRef = ref
		}

		result[property.Name] = prop
	}
}

func mapEntitySet(entitySet *ods.EntitySet, objects *edmObjects) *Collection {
	entityType := objects.entityTypes[entitySet.EntityType]
	keys := make([]string, len(*entityType.Key))

	for i, keyRef := range *entityType.Key {
		keys[i] = keyRef.Name
	}

	collection := Collection{
		Key: keys,
		Structure: Structure{
			Name:       entitySet.Name,
			Properties: make(map[string]Property),
		},
	}

	addStructuralProperties(entitySet.EntityType, objects, collection.Properties)
	addNavProperties(entitySet, objects, collection.Properties)

	return &collection
}

func mapComplexType(complexTypeName string, complexType *ods.ComplexType, objects *edmObjects) *Structure {
	cType := Structure{
		Name:       complexType.Name,
		Properties: make(map[string]Property),
	}

	addStructuralProperties(complexTypeName, objects, cType.Properties)
	addComplexTypeNavProperties(complexTypeName, objects, cType.Properties)
	return &cType
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
		Collections: []Collection{},
		Endpoints:   []Endpoint{},
		Structures:  []Structure{},
	}

	objects := extractObjects(edm)

	for _, entitySet := range objects.entityContainer.EntitySets {
		collection := mapEntitySet(&entitySet, objects)
		schema.Collections = append(schema.Collections, *collection)
	}

	for typeName, complexType := range objects.complexTypes {
		schema.Structures = append(schema.Structures, *mapComplexType(typeName, complexType, objects))
	}

	schema.Name = objects.entityContainer.Name
	return schema, nil
}
