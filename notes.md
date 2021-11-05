### Schema-parsing-related TODOs

- EDMX can reference other EDMX docs or annotations, etc - <edmx:AnnotationsReference />, <edmx:Reference />
- Function/Action overloading
- Parsing Singleton for /me endpoint, etc
- If (NavigationPropertyBinding) omitted, clients MUST assume that the target entity set or singleton can vary per related entity.

- Abstract types and inheritance

### Model-related TODOs

- Sort out primitive data type names and mappings
- Relations to nested documents in OData requires a relative path and can contain type casts
- Dynamic properties:
    `If an EntityType is an OpenEntityType, the set of properties that are associated with the EntityType can, in addition to declared properties, include dynamic properties.`

- Abstract types and inheritance

## Discussion:
- Function/Action binding context
- Implicit, contained navigation properties

##
- Sort out primitive data type names and mappings
- Duplicates
- Do we need enum values or just their names

## Resources:

XML Schema of the EDMX document
    https://docs.oasis-open.org/odata/odata-csdl-xml/v4.01/os/schemas/edm.xsd

EDMX schema
    https://docs.microsoft.com/en-us/openspecs/windows_protocols/mc-edmx/5dff5e25-56a1-408b-9d44-bff6634c7d16

EDMX spec
    http://docs.oasis-open.org/odata/odata/v4.0/errata03/os/complete/part3-csdl/odata-v4.0-errata03-os-part3-csdl-complete.html#_Toc453752536

OData tutorial trip pin:
https://www.odata.org/blog/trippin-new-odata-v4-sample-service/
https://www.odata.org/getting-started/basic-tutorial/

Examples of function/action bindings:
https://olingo.apache.org/doc/odata4/tutorials/action/tutorial_bound_action.html

