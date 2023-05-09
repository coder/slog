# logfmt

logfmt provides an implementation that supports nested objects and arrays. It
is meant to be a drop-in replacement for JSON, which other logfmt implements
do not support.

This package makes the trade-off of being more difficult for computers to parse
in favor of the human.

See these examples:

JSON:
```json
{ "user": { "id": 123, "name": "foo", "age": 20, "hobbies": ["basketball", "football"] } }
```

flat logfmt:
```
user.id=123 user.name=foo user.age=20 user.hobbies="basketball,football"
```

nested logfmt:
```
user={ id=123 name=foo age=20 hobbies=[basketball football] }
```
