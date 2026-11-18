# Policy schema

Policy is one JSON document. `intra compile` validates it and writes a
deterministic compiled view. Edit the source only.

See the runnable [example policy](../../examples/policy.example.json).

## Top level

| Field | Type | Required | Meaning |
|---|---|---:|---|
| `version` | integer | yes | Must be at least `1` |
| `actors` | object | yes | Actors by stable name |
| `services` | object | yes | Services by stable name |
| `rules` | array | yes | Allow rules in evaluation order |
| `templates` | array | no | Bounded read or egress templates |
| `egress` | array | no | Host and port exit rules |

Unknown actor, service, action, or template references fail validation.

## Actor

| Field | Type | Meaning |
|---|---|---|
| `name` | string | Must match the object key in generated demo policy |
| `kind` | string | `human`, `robot`, `service`, or `drill` by convention |
| `owner` | string | Accountable person or team |

## Service

| Field | Type | Meaning |
|---|---|---|
| `name` | string | Must match the object key |
| `kind` | string | `http`, `postgres`, `logs`, `metrics`, `saas`, or `egress` |
| `sensitive` | boolean | Applies sensitive TTL and approval rules |
| `upstream` | string | HTTP upstream URL; empty for Postgres and egress |
| `actions` | string array | Actions the service offers |
| `owner` | string | Required accountable owner |
| `objects` | string array | Database tables a read role may select |
| `routes` | route array | Optional method and prefix action map |

Actions are `read`, `write`, `admin`, `export`, `connect`, and `egress`.

## Route

| Field | Type | Meaning |
|---|---|---|
| `method` | string | HTTP method, or empty for every method |
| `prefix` | string | Service-local path prefix |
| `action` | action | Action assigned when this route wins |

The longest matching prefix wins. Prefixes match only at a path segment
boundary. An `/admin` segment upgrades to `admin` unless the winning
route itself names that admin path.

## Rule

| Field | Type | Meaning |
|---|---|---|
| `actor` | string | Actor name |
| `service` | string | Service name |
| `actions` | action array | Actions this rule can allow |
| `access` | string | `static`, `grant`, or `template` |
| `template_id` | string | Required for `template` access |
| `max_ttl_seconds` | integer | Optional rule-level TTL cap |

Rules evaluate in file order. First match wins.

## Template

Templates may grant only `read` and `egress` actions. Their TTL must be
between 1 and 7200 seconds. Every template must be referenced by a rule.
A request must fit inside the template services and actions exactly.

## Egress rule

| Field | Type | Meaning |
|---|---|---|
| `actor` | string | Actor allowed to use the proxy |
| `hosts` | string array | Exact host or `*.suffix` wildcard |
| `ports` | integer array | Allowed destination ports |

Policy match alone never dials. The proxy also rejects every non-global
resolved address and dials the validated IP.

## Compiled output

`compiled.json` contains version, source hash, mesh route notes, guard
upstreams, sorted broker rules, and sorted services. Treat it as build
output. Regenerate it after every policy edit.
