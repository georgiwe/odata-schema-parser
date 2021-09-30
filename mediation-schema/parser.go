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
	functions       map[string]*ods.Function
	functionImports map[string]*ods.FunctionImport
	actions         map[string]*ods.Action
	actionImports   map[string]*ods.ActionImport
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

func addToFunctions(objects edmObjects, schema *ods.Schema, function ods.Function) {
	namespacedName, aliasedName := formQualifiedName(schema, function.Name)
	objects.functions[namespacedName] = &function
	if aliasedName != "" {
		objects.functions[aliasedName] = &function
	}
}

func addToActions(objects edmObjects, schema *ods.Schema, action ods.Action) {
	namespacedName, aliasedName := formQualifiedName(schema, action.Name)
	objects.actions[namespacedName] = &action
	if aliasedName != "" {
		objects.actions[aliasedName] = &action
	}
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

func typeToProperty(typeName string, objects *edmObjects) Property {
	result := Property{
		PropertyType: "unknown",
		ValueType:    fmt.Sprintf("unknown %s", typeName),
		IsCollection: false,
	}
	if mappedType, err := mapEdmType(typeName); err == nil {
		result.PropertyType = "primitive"
		result.ValueType = mappedType
	} else if strings.HasPrefix(typeName, collectionPrefix) {
		actualType := typeName[len(collectionPrefix) : len(typeName)-1]
		mapped := typeToProperty(actualType, objects)
		mapped.IsCollection = true
		result = mapped
	} else if _, ok := objects.entityTypes[typeName]; ok {
		result.ValueType = typeName
		result.PropertyType = "relation"
	} else if _, ok := objects.complexTypes[typeName]; ok {
		result.PropertyType = "structure"
		result.ValueType = typeName
	} else if _, ok := objects.enumTypes[typeName]; ok {
		result.PropertyType = "enum"
		result.ValueType = typeName
	}
	// There is also a type of "entity", not listed here
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

func addStructuralProperties(typeName string, objects *edmObjects, result map[string]Property) {
	properties := getTypeProperties(typeName, objects)
	for _, property := range properties {
		prop := typeToProperty(property.Type, objects)
		result[property.Name] = prop
	}
}

func addNavProperties(entityTypeQualifiedName string, objects *edmObjects, result map[string]Property) {
	properties := getTypeNavProperties(entityTypeQualifiedName, objects)

	for _, property := range properties {
		prop := typeToProperty(property.Type, objects)

		result[property.Name] = prop
	}
}

func addComplexTypeNavProperties(cTypeName string, objects *edmObjects, result map[string]Property) {
	properties := getTypeNavProperties(cTypeName, objects)

	for _, property := range properties {
		prop := typeToProperty(property.Type, objects)

		result[property.Name] = prop
	}
}

func getTypeKeys(srcQualifiedName string, objects *edmObjects) []string {
	src := objects.entityTypes[srcQualifiedName]
	if src.Key != nil {
		keys := make([]string, len(*src.Key))

		for i, keyRef := range *src.Key {
			keys[i] = keyRef.Name
		}

		return keys
	} else if src.BaseType != nil {
		return getTypeKeys(*src.BaseType, objects)
	} else {
		panic("could not find key")
	}
}

func mapEntityType(qualifiedName string, objects *edmObjects) EntityType {
	entityType := objects.entityTypes[qualifiedName]

	mappedEntityType := EntityType{
		Key:       getTypeKeys(qualifiedName, objects),
		HasStream: entityType.HasStream,
		BaseType:  entityType.BaseType,
		Structure: Structure{
			Name:       entityType.Name,
			Properties: make(map[string]Property),
		},
	}

	addStructuralProperties(qualifiedName, objects, mappedEntityType.Properties)
	addNavProperties(qualifiedName, objects, mappedEntityType.Properties)

	return mappedEntityType
}

func mapCollection(entitySet ods.EntitySet, objects *edmObjects) Collection {
	res := Collection{
		Name:         entitySet.Name,
		EntityType:   entitySet.EntityType,
		Downloadable: objects.entityTypes[entitySet.EntityType].HasStream,
	}

	return res
}

func mapComplexType(complexTypeName string, objects *edmObjects) Structure {
	complexType := objects.complexTypes[complexTypeName]
	cType := Structure{
		Name:       complexType.Name,
		Properties: make(map[string]Property),
	}

	addStructuralProperties(complexTypeName, objects, cType.Properties)
	addComplexTypeNavProperties(complexTypeName, objects, cType.Properties)

	return cType
}

func mapEnumType(enumName string, objects *edmObjects) Enum {
	enum := objects.enumTypes[enumName]
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

	return eType
}

func mapFunction(funcName string, function *ods.Function, objects *edmObjects) Invocation {
	// TODO: use a different type for the result, not Property
	funcResult := typeToProperty(function.ReturnType.Type, objects)

	inv := Invocation{
		Name:             function.Name,
		BindingType:      "unknown",
		BoundDataPointer: function.EntitySetPath,
		Arguments:        make([]Property, len(function.Parameters)),
		Result:           &funcResult,
	}

	for i, param := range function.Parameters {
		inv.Arguments[i] = typeToProperty(param.Type, objects)
	}

	if _, found := objects.functionImports[funcName]; found && !function.IsBound {
		inv.BindingType = "unbound"
	} else if function.IsBound {
		entityType := typeToProperty(function.Parameters[0].Type, objects)
		if entityType.IsCollection {
			inv.BindingType = "collection"
		} else {
			inv.BindingType = "entity"
		}

		inv.BoundTo = &inv.Arguments[0].ValueType
	}

	return inv
}

// TODO: consolidate with mapFunction?
func mapAction(actionName string, action *ods.Action, objects *edmObjects) Invocation {
	var result *Property = nil

	if action.ReturnType != nil {
		// TODO: use a different type for the result, not Property
		prop := typeToProperty(action.ReturnType.Type, objects)
		result = &prop
	}

	inv := Invocation{
		Name:             action.Name,
		BindingType:      "unknown",
		BoundDataPointer: action.EntitySetPath,
		Arguments:        make([]Property, len(action.Parameters)),
		Result:           result,
	}

	for i, param := range action.Parameters {
		inv.Arguments[i] = typeToProperty(param.Type, objects)
	}

	if _, found := objects.actionImports[actionName]; found && !action.IsBound {
		inv.BindingType = "unbound"
	} else if action.IsBound {
		entityType := typeToProperty(action.Parameters[0].Type, objects)
		if entityType.IsCollection {
			inv.BindingType = "collection"
		} else {
			inv.BindingType = "entity"
		}

		// inv.BoundTo = &action.Parameters[0].Type
		inv.BoundTo = &inv.Arguments[0].ValueType
	}

	return inv
}

func extractObjects(edm *ods.EdmxDocument) *edmObjects {
	objects := edmObjects{
		entityTypes:     make(map[string]*ods.EntityType),
		complexTypes:    make(map[string]*ods.ComplexType),
		enumTypes:       make(map[string]*ods.EnumType),
		functions:       make(map[string]*ods.Function),
		actions:         make(map[string]*ods.Action),
		functionImports: make(map[string]*ods.FunctionImport),
		actionImports:   make(map[string]*ods.ActionImport),
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
		for _, function := range schema.Functions {
			addToFunctions(objects, &schema, function)
		}
		for _, action := range schema.Actions {
			addToActions(objects, &schema, action)
		}
	}

	for _, functionImport := range objects.entityContainer.FunctionImports {
		objects.functionImports[functionImport.Function] = &functionImport
	}

	for _, actionImport := range objects.entityContainer.ActionImports {
		objects.actionImports[actionImport.Action] = &actionImport
	}

	return &objects
}

func Parse(edm *ods.EdmxDocument) (Service, error) {
	service := Service{
		Collections: make(map[string]Collection),
		Invocations: make(map[string]Invocation),
		Types:       make(map[string]Type),
	}

	objects := extractObjects(edm)

	for _, entitySet := range objects.entityContainer.EntitySets {
		collection := mapCollection(entitySet, objects)
		service.Collections[entitySet.Name] = collection
	}

	for name := range objects.entityTypes {
		et := mapEntityType(name, objects)
		service.Types[name] = Type{
			Kind:       "EntityType",
			EntityType: &et,
		}
	}

	for name := range objects.complexTypes {
		ct := mapComplexType(name, objects)
		service.Types[name] = Type{
			Kind:      "Structure",
			Structure: &ct,
		}
	}

	for name := range objects.enumTypes {
		enum := mapEnumType(name, objects)
		service.Types[name] = Type{
			Kind: "Enum",
			Enum: &enum,
		}
	}

	for funcName, function := range objects.functions {
		service.Invocations[funcName] = mapFunction(funcName, function, objects)
	}

	for actionName, action := range objects.actions {
		service.Invocations[actionName] = mapAction(actionName, action, objects)
	}

	// TODO: singleton for /me
	// TODO: figure out if this "Nav property count mismatch. EntitySet: People, EntitySet count: 6, NavProp count: 3" is ok

	service.Name = objects.entityContainer.Name
	service.Type = "OData4"
	return service, nil
}
