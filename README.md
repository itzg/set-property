Implements a properties file manipulator that copies over a value from an environment variable optionally applying constraints and mappings.

Bulk JSON definitions can specify a `fallback` for a property that is absent
from the file and whose environment variable is unset:

```json
{
  "server-ip": {"env": "SERVER_IP", "fallback": ""},
  "motd": {"env": "MOTD", "fallback": "Welcome"}
}
```

The fallback may be an empty string, and `env` may be omitted. Fallbacks are
literal values; `mappings` and `allowed` do not apply to them. Existing
properties, including empty ones, are not replaced by fallbacks. As with other
definitions, an environment variable set to an empty string does not change the
property, and it prevents the fallback from being used.