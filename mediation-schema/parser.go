package mediationschema

import (
	"fmt"
	"strings"

	ods "github.com/kinvey/odata-schema/odata-schema"
)

const collectionPrefix = "Collection("

func mapEntityType(qualifiedName string, objects *edmObjects) (EntityType, error) {
	if _, ok := objects.entityTypes[qualifiedName]; !ok {
		return EntityType{}, fmt.Errorf("unable to map entity type. entity type '%s' was not defined", qualifiedName)
	}

	entityType := objects.entityTypes[qualifiedName]
	typeKeys, err := getTypeKeys(qualifiedName, objects)

	if err != nil {
		return EntityType{}, err
	}

	mappedType := EntityType{
		Key:        typeKeys,
		Streamable: entityType.HasStream,
		BaseType:   entityType.BaseType,
		Structure: Structure{
			Name:       entityType.Name,
			Properties: make(map[string]Property),
			OpenType:   entityType.OpenType,
		},
	}

	addStructuralProperties(qualifiedName, objects, mappedType.Properties)
	addNavProperties(qualifiedName, objects, mappedType.Properties)

	return mappedType, nil
}

func mapCollection(entitySet ods.EntitySet, objects *edmObjects) (Collection, error) {
	if _, ok := objects.entityTypes[entitySet.EntityType]; !ok {
		return Collection{}, fmt.Errorf("unable to map collection. entity type '%s' was not defined", entitySet.EntityType)
	}

	res := Collection{
		Name:       entitySet.Name,
		EntityType: entitySet.EntityType,
		Streamable: objects.entityTypes[entitySet.EntityType].HasStream,
	}

	return res, nil
}

func mapComplexType(qualifiedName string, objects *edmObjects) (Structure, error) {
	if _, ok := objects.complexTypes[qualifiedName]; !ok {
		return Structure{}, fmt.Errorf("unable to map complex type. complex type '%s' was not defined", qualifiedName)
	}

	complexType := objects.complexTypes[qualifiedName]
	mappedType := Structure{
		Name:       complexType.Name,
		Properties: make(map[string]Property),
		OpenType:   complexType.OpenType,
	}

	addStructuralProperties(qualifiedName, objects, mappedType.Properties)
	addNavProperties(qualifiedName, objects, mappedType.Properties)

	return mappedType, nil
}

func mapEnumType(qualifiedName string, objects *edmObjects) (Enum, error) {
	if _, ok := objects.enumTypes[qualifiedName]; !ok {
		return Enum{}, fmt.Errorf("unable to map enum type. enum type '%s' was not defined", qualifiedName)
	}

	enum := objects.enumTypes[qualifiedName]

	eType := Enum{
		Name:        enum.Name,
		Members:     make(map[string]string),
		ValuesType:  enum.UnderlyingType,
		Multiselect: enum.IsFlags,
	}

	if eType.ValuesType == "" {
		mappedType, _ := mapEdmType("Edm.Int32") // Specified in MC-CSDL
		eType.ValuesType = mappedType
	}

	for _, member := range enum.Members {
		eType.Members[member.Name] = member.Value
	}

	return eType, nil
}

func mapFunction(funcName string, function *ods.Function, objects *edmObjects) (Invocation, error) {
	// TODO: use a different type for the result, not Property
	funcResult, err := typeToProperty(function.ReturnType.Type, objects)

	if err != nil {
		return Invocation{}, err
	}

	inv := Invocation{
		Name:             function.Name,
		BindingType:      "unknown",
		BoundDataPointer: function.EntitySetPath,
		Arguments:        make([]InvocationArgument, len(function.Parameters)),
		Result:           &funcResult,
	}

	for i, param := range function.Parameters {
		if prop, err := typeToProperty(param.Type, objects); err != nil {
			return Invocation{}, err
		} else {
			inv.Arguments[i] = InvocationArgument{
				Name:     param.Name,
				Property: prop,
			}
		}
	}

	if _, found := objects.functionImports[funcName]; found && !function.IsBound {
		inv.BindingType = "unbound"
	} else if function.IsBound {
		if entityType, err := typeToProperty(function.Parameters[0].Type, objects); err != nil {
			return Invocation{}, err
		} else {
			if entityType.IsCollection {
				inv.BindingType = "collection"
			} else {
				inv.BindingType = "entity"
			}
		}

		inv.BoundTo = &inv.Arguments[0].ValueType
	}

	return inv, nil
}

// TODO: consolidate with mapFunction?
func mapAction(actionName string, action *ods.Action, objects *edmObjects) (Invocation, error) {
	var result *Property = nil

	if action.ReturnType != nil {
		// TODO: use a different type for the result, not Property
		if prop, err := typeToProperty(action.ReturnType.Type, objects); err != nil {
			return Invocation{}, err
		} else {
			result = &prop
		}
	}

	inv := Invocation{
		Name:             action.Name,
		BindingType:      "unknown",
		BoundDataPointer: action.EntitySetPath,
		Arguments:        make([]InvocationArgument, len(action.Parameters)),
		Result:           result,
	}

	for i, param := range action.Parameters {
		if prop, err := typeToProperty(param.Type, objects); err != nil {
			return Invocation{}, err
		} else {
			inv.Arguments[i] = InvocationArgument{
				Name:     param.Name,
				Property: prop,
			}
		}
	}

	if _, found := objects.actionImports[actionName]; found && !action.IsBound {
		inv.BindingType = "unbound"
	} else if action.IsBound {
		if entityType, err := typeToProperty(action.Parameters[0].Type, objects); err != nil {
			return Invocation{}, err
		} else {
			if entityType.IsCollection {
				inv.BindingType = "collection"
			} else {
				inv.BindingType = "entity"
			}
		}

		// inv.BoundTo = &action.Parameters[0].Type
		inv.BoundTo = &inv.Arguments[0].ValueType
	}

	return inv, nil
}

func mapEDMObjectsToService(objects *edmObjects) (*Service, error) {
	service := &Service{
		Name:        objects.entityContainer.Name,
		Type:        "OData4",
		Collections: make(map[string]Collection),
		Invocations: make(map[string]Invocation),
		Types:       make(map[string]Type),
	}

	for _, entitySet := range objects.entityContainer.EntitySets {
		if collection, err := mapCollection(entitySet, objects); err != nil {
			return nil, err
		} else {
			service.Collections[entitySet.Name] = collection
		}
	}

	for name := range objects.entityTypes {
		if et, err := mapEntityType(name, objects); err != nil {
			return nil, err
		} else {
			service.Types[name] = Type{
				Kind:       "EntityType",
				EntityType: &et,
			}
		}
	}

	for name := range objects.complexTypes {
		if ct, err := mapComplexType(name, objects); err != nil {
			return nil, err
		} else {
			service.Types[name] = Type{
				Kind:      "Structure",
				Structure: &ct,
			}
		}
	}

	for name := range objects.enumTypes {
		if enum, err := mapEnumType(name, objects); err != nil {
			return nil, err
		} else {
			service.Types[name] = Type{
				Kind: "Enum",
				Enum: &enum,
			}
		}
	}

	for funcName, function := range objects.functions {
		if mapped, err := mapFunction(funcName, function, objects); err != nil {
			return nil, err
		} else {
			service.Invocations[funcName] = mapped
		}
	}

	for actionName, action := range objects.actions {
		if mapped, err := mapAction(actionName, action, objects); err != nil {
			return nil, err
		} else {
			service.Invocations[actionName] = mapped
		}
	}

	// TODO: singleton for /me
	// TODO: figure out if this "Nav property count mismatch. EntitySet: People, EntitySet count: 6, NavProp count: 3" is ok

	return service, nil
}

func formQualifiedName(schema *ods.Schema, objectName string) (string, string) {
	namespacedName := fmt.Sprintf("%s.%s", schema.Namespace, objectName)
	aliasedName := ""

	if schema.Alias != nil {
		aliasedName = fmt.Sprintf("%s.%s", *schema.Alias, objectName)
	}

	return namespacedName, aliasedName
}

func mapEdmType(edmType string) (string, error) {
	switch edmType {
	case "Edm.String", "Edm.Guid":
		return "string", nil
	case "Edm.Boolean":
		return "boolean", nil
	case "Edm.Date":
		return "date", nil
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
	case "Edm.Duration": // TODO: what?
		return "duration", nil
	default:
		if strings.HasPrefix(edmType, "Edm.") {
			panic(edmType)
		}
		return "", fmt.Errorf("unknown Type: %s", edmType)
	}
}

func findCollectionsByEntityType(qualifiedName string, objects *edmObjects) []string {
	result := []string{}
	for _, es := range objects.entityContainer.EntitySets {
		if es.EntityType == qualifiedName {
			result = append(result, es.Name)
		}
	}
	return result
}

func typeToProperty(typeName string, objects *edmObjects) (Property, error) {
	result := Property{
		PropertyType: "unknown",
		ValueType:    fmt.Sprintf("unknown (%s)", typeName),
		IsCollection: false,
	}
	if mappedType, err := mapEdmType(typeName); err == nil {
		result.PropertyType = "primitive"
		result.ValueType = mappedType
	} else if strings.HasPrefix(typeName, collectionPrefix) {
		actualType := typeName[len(collectionPrefix) : len(typeName)-1]
		if mapped, err := typeToProperty(actualType, objects); err != nil {
			return Property{}, err
		} else {
			mapped.IsCollection = true
			result = mapped
		}
	} else if _, ok := objects.entityTypes[typeName]; ok {
		result.ValueType = typeName
		result.PropertyType = "relation"
		collections := findCollectionsByEntityType(typeName, objects)
		if len(collections) == 0 {
			return Property{}, fmt.Errorf("unable to find collection for entity type '%s'", typeName)
		}
		// TODO: return some sort of ambiguity descriptor when len(collections) > 1
		// else if len(collections) > 1 {
		// 	return Property{}, fmt.Errorf("unable to find collection for entity type '%s'", typeName)
		// }
		result.RelationCollection = &collections[0]
	} else if _, ok := objects.complexTypes[typeName]; ok {
		result.PropertyType = "structure"
		result.ValueType = typeName
	} else if _, ok := objects.enumTypes[typeName]; ok {
		result.PropertyType = "enum"
		result.ValueType = typeName
	}
	return result, nil
}

func getTypeStructuralProperties(qualifiedName string, objects *edmObjects) []ods.Property {
	var properties []ods.Property
	var baseType *string
	if foundType, ok := objects.entityTypes[qualifiedName]; ok {
		properties = foundType.Properties
		baseType = foundType.BaseType
	} else if foundType, ok := objects.complexTypes[qualifiedName]; ok {
		properties = foundType.Properties
		baseType = foundType.BaseType
	}
	if baseType != nil {
		baseProps := getTypeStructuralProperties(*baseType, objects)
		properties = append(properties, baseProps...)
	}
	return properties
}

func getTypeNavProperties(qualifiedName string, objects *edmObjects) []ods.NavigationProperty {
	var properties []ods.NavigationProperty
	var baseType *string
	if foundType, ok := objects.entityTypes[qualifiedName]; ok {
		properties = foundType.NavigationProperties
		baseType = foundType.BaseType
	} else if foundType, ok := objects.complexTypes[qualifiedName]; ok {
		properties = foundType.NavigationProperties
		baseType = foundType.BaseType
	}
	if baseType != nil {
		baseProps := getTypeNavProperties(*baseType, objects)
		properties = append(properties, baseProps...)
	}
	return properties
}

func addStructuralProperties(typeName string, objects *edmObjects, result map[string]Property) error {
	properties := getTypeStructuralProperties(typeName, objects)
	for _, property := range properties {
		prop, err := typeToProperty(property.Type, objects)
		if err != nil {
			return err
		}
		if property.Nullable != nil {
			prop.Required = !*property.Nullable
		} else {
			prop.Required = false
		}
		result[property.Name] = prop
	}
	return nil
}

func addNavProperties(qualifiedName string, objects *edmObjects, result map[string]Property) error {
	for _, property := range getTypeNavProperties(qualifiedName, objects) {
		if prop, err := typeToProperty(property.Type, objects); err != nil {
			return err
		} else {
			result[property.Name] = prop
		}
	}
	return nil
}

func getTypeKeys(qualifiedName string, objects *edmObjects) ([]string, error) {
	src := objects.entityTypes[qualifiedName]
	if src.Key != nil {
		keys := make([]string, len(*src.Key))

		for i, keyRef := range *src.Key {
			keys[i] = keyRef.Name
		}

		return keys, nil
	} else if src.BaseType != nil {
		return getTypeKeys(*src.BaseType, objects)
	} else {
		return []string{}, fmt.Errorf("unable to find keys for type '%s'", qualifiedName)
	}
}

func Parse(backendName string, edm *ods.EdmxDocument) (*Service, error) {
	if objects, err := extractObjects(backendName, edm); err != nil {
		return nil, err
	} else {
		return mapEDMObjectsToService(objects)
	}
}
