encountered issues when doing data model -> GQL schema:
- Primitive type names
    - do we need to know more than the gql types do - int16 vs Int
    - GQL doesn't have date, duration, etc - shall we create some custom types?
- Can we have all types in a single palce, not separated into entity types, structures, enums
- Composite keys don't seem to be supported in GQL
- Entity (and other) type names - do we need the qualified names? Make names valid (omit dots) and more readable and keep unique names in a directive
- Cross-backend name collisions


Things to drop:
- type casts
- actions bound to entities