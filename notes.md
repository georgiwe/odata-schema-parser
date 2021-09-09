## Findings:
- EDMX can reference other EDMX docs or annotations, etc - <edmx:AnnotationsReference />, <edmx:Reference />
- Dynamic properties:
    `If an EntityType is an OpenEntityType, the set of properties that are associated with the EntityType can, in addition to declared properties, include dynamic properties.`

## Resources:

XML Schema of the EDMX document
    https://docs.oasis-open.org/odata/odata-csdl-xml/v4.01/os/schemas/edm.xsd

EDMX schema
    https://docs.microsoft.com/en-us/openspecs/windows_protocols/mc-edmx/5dff5e25-56a1-408b-9d44-bff6634c7d16

EDMX spec
    http://docs.oasis-open.org/odata/odata/v4.0/errata03/os/complete/part3-csdl/odata-v4.0-errata03-os-part3-csdl-complete.html#_Toc453752536
