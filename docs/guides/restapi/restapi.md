# About the REST API service

Your application and the public Runtime Radar API interact via the REST API over HTTPS. The supported request methods are `GET`, `POST`, `PATCH`, and `DELETE`.

Root API URL:

```
https://<web server address>/api/v1
```

For example:

```
https://ptcs.example.com/api/v1
```

In response to your application requests, the service returns messages in JSON format.

***Note.** If you want other requests to be added to the public Runtime Radar API, contact Technical Support.*

## Authorizing requests via the REST API

To send requests, which require user authentication, to the public API, you must create a personal access token with permissions needed to execute the correponding requests. For example, to execute a request to create a response rule, a user access token must have the **Create response rules** permission.

## Managing response rules

Response rules determine how Runtime Radar to respond when threats or vulnerabilities are detected in scanned and checked objects.

You can manage response rules both in the Runtime Radar web interface and using the public API. All requests for managing response rules are authorized using the `access_token` setting ([access token](#10359847435)) with the String data type.

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Creating response rules

You can create response rules in the system using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **Create response rules** permission.

Request method and URL:

```
POST <root API URL>/public-api/rule
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type.

The request body parameters are described in the table below.

<table><caption>Body parameters of a request to create a response rule</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.9%;"/><col style="width: 18.9%;"/><col style="width: 42.0%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Required
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`name`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule name
</td></tr><tr><td align="left">

`type`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule type: `TYPE_RUNTIME`
</td></tr><tr><td align="left">

`rule`
</td><td align="left">

Yes
</td><td align="left">

Array of JSON objects
</td><td align="left">

Rule parameters
</td></tr><tr><td align="left">

`rule` → `version`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule version. The only possible value is `1`
</td></tr><tr><td align="left">

`rule` → `block`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Blocking parameters. If you specify `null`, blocking will not be performed. The request must contain `rule` → `block` → `severity`
</td></tr><tr><td align="left">

`rule` → `block` → `severity`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Block vulnerabilities or threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Notification emailing settings. If you specify `null`, notifications will not be sent. The request must contain `rule` → `notify` → `severity`
</td></tr><tr><td align="left">

`rule` → `notify` → `severity`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Notify about threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify` → `targets`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

List of addresses to which notifications will be sent
</td></tr><tr><td align="left">

`rule` → `whitelist`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

List of exclusions for the rule to skip
</td></tr><tr><td align="left">

`rule` → `whitelist` → `threats`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

[List of threats](#11978146571). Can be empty
</td></tr><tr><td align="left">

`rule` → `whitelist` → `binaries`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

List of executable files. Can be empty
</td></tr><tr><td align="left">

`scope`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Rule scope
</td></tr><tr><td align="left">

`scope` → `version`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule scope version. The only possible value is `1`
</td></tr><tr><td align="left">

`scope` → `registries`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of registries containing images that will be covered by the rule
</td></tr><tr><td align="left">

`scope` → `image_names`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of the images that will be covered by the rule
</td></tr><tr><td align="left">

`scope` → `clusters`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Cluster for which the rule will be created
</td></tr><tr><td align="left">

`scope` → `namespace`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of namespaces containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `pods`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of pods containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `containers`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of containers containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `nodes`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of nodes containing objects that will be covered by the rule. Can be empty
</td></tr></tbody></table>


**Response to a request**

In response to a successful request, the service returns code 200 and a message with the response rule identifier.

In response to an unsuccessful request, the service returns an empty message with an [error code](#9572838667).

**Example**

Request:

```
POST https://runtimeradar.example.com/api/v1/public-api/rule
```

Request body:

```
{
  "name": "runtime rule",
  "type": "TYPE_RUNTIME",
  "rule": {
    "version": "1",
    "block": { "severity": "high" },
    "notify": {
      "severity": "medium",
      "targets": ["9f1c0b7a-2f44-4a1e-8e6d-5c2b7a9e0011"]
    },
    "whitelist": {
      "threats": ["CS_RT_CRYPTOMINER", "CS_RT_RECON_TOOLS"],
      "binaries": ["/usr/bin/curl", "/opt/**/agent"]
    }
  },
  "scope": {
    "version": "1",
    "clusters": ["prod-*"],
    "namespaces": ["default"],
    "image_names": ["registry.local/*"],
    "registries": [],
    "pods": [],
    "containers": [],
    "nodes": []
  }
}
```

Response to a successful request:

```
{
  "id": "17e6a196-dab1-4d49-806d-44c89700084d"
}
```

Response to an unsuccessful request:

```
{
  "code": 3,
  "message": "unexpected EOF",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "BAD_REQUEST"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching response rule parameters

You can fetch the parameters of an existing response rule using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **View response rules** permission.

Request method and URL:

```
GET <root API URL>/public-api/rule/{id}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the `id` path parameter (response rule identifier[^6]) with the String data type.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Scheme of a response to a request to fetch response rule data</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`rule`
</td><td align="left">

Array of JSON objects
</td><td align="left">

Rule
</td></tr><tr><td align="left">

`rule` → `id`
</td><td align="left">

String
</td><td align="left">

Rule identifier
</td></tr><tr><td align="left">

`rule` → `name`
</td><td align="left">

String
</td><td align="left">

Rule name
</td></tr><tr><td align="left">

`rule` → `type`
</td><td align="left">

String
</td><td align="left">

Rule type: `TYPE_RUNTIME`
</td></tr><tr><td align="left">

`rule` →` rule`
</td><td align="left">

Array of JSON objects
</td><td align="left">

Rule parameters
</td></tr><tr><td align="left">

`rule` → `rule` → `version`
</td><td align="left">

String
</td><td align="left">

Rule version. The only possible value is `1`
</td></tr><tr><td align="left">

`rule` → `rule` → `block`
</td><td align="left">

JSON object
</td><td align="left">

Blocking parameters. If the received value is `null`, blocking will not be performed. The response body also contains `rule` → `block` → `severity`
</td></tr><tr><td align="left">

`rule` → `rule` → `block` → `severity`
</td><td align="left">

String
</td><td align="left">

Block threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `rule` → `notify`
</td><td align="left">

JSON object
</td><td align="left">

Notification emailing settings. If the received value is `null`, a notification will not be sent. The response body also contains `rule` → `notify` → `severity`
</td></tr><tr><td align="left">

`rule` → `rule` → `notify` → `severity`
</td><td align="left">

String
</td><td align="left">

Notify about threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `rule` → `notify` → `targets`
</td><td align="left">

Array of strings
</td><td align="left">

List of addresses to which notifications will be sent
</td></tr><tr><td align="left">

`rule` → `rule` → `whitelist` → `threats`
</td><td align="left">

Array of strings
</td><td align="left">

[List of threats](#11978146571). Can be empty
</td></tr><tr><td align="left">

`rule` → `scope`
</td><td align="left">

JSON object
</td><td align="left">

Rule scope
</td></tr><tr><td align="left">

`rule` → `scope` → `version`
</td><td align="left">

String
</td><td align="left">

Rule scope version. The only possible value is `1`
</td></tr><tr><td align="left">

`rule` → `scope` → `registries`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of registries containing images that are be covered by the rule
</td></tr><tr><td align="left">

`rule` → `scope` → `image_names`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of the images covered by the rule
</td></tr><tr><td align="left">

`rule` → `scope` → `clusters`
</td><td align="left">

Array of strings
</td><td align="left">

Cluster for which the rule will be created
</td></tr><tr><td align="left">

`rule` → `scope` → `namespace`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of namespaces containing objects covered by the rule. Can be empty
</td></tr><tr><td align="left">

`rule` → `scope` → `pods`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of pods containing objects covered by the rule. Can be empty
</td></tr><tr><td align="left">

`rule` → `scope` → `containers`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of containers containing objects covered by the rule. Can be empty
</td></tr><tr><td align="left">

`rule` → `scope` → `nodes`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of nodes containing objects covered by the rule. Can be empty
</td></tr><tr><td align="left">

`rule` → `deleted`
</td><td align="left">

Boolean
</td><td align="left">

Shows if a rule is deleted. Possible values: `true` (deleted) and `false` (not deleted)
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns an empty message with an [error code](#9572838667).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/rule-123
```

Response to a successful request:

```
{
  "rule": {
    "name": "runtime rule",
    "rule": {
      "version": "1",
      "block": { "severity": "high" },
      "notify": { "severity": "medium", "targets": ["9f1c0b7a-2f44-4a1e-8e6d-5c2b7a9e0011"] },
      "whitelist": { "threats": ["CS_RT_CRYPTOMINER"], "binaries": ["/usr/bin/curl"] }
    },
    "type": "TYPE_RUNTIME",
    "scope": {
      "version": "1",
      "image_names": [],
      "registries": [],
      "clusters": ["prod-*"],
      "namespaces": ["default"],
      "pods": [],
      "containers": [],
      "nodes": []
    }
  },
  "deleted": false
}
```

Response to an unsuccessful request:

```
{
  "code": 16,
  "message": "access token not found",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "UNAUTHENTICATED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

[^4]: You can find out the identifier of a response rule by sending a [request for the list of response rules](#11949896587).

### Changing response rule settings

You can change the settings of an existing response rule using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **Edit response rules** permission.

Request method and URL:

```
PATCH <root API URL>/public-api/rule/{id}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the `id` path parameter (response rule identifier[^6]) with the String data type as well as request body parameters described in the table below.

<table><caption>Body parameters of a request to edit the response rule settings</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.9%;"/><col style="width: 18.9%;"/><col style="width: 42.0%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Required
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`name`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule name
</td></tr><tr><td align="left">

`type`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule type: `TYPE_RUNTIME`
</td></tr><tr><td align="left">

`rule`
</td><td align="left">

Yes
</td><td align="left">

Array of JSON objects
</td><td align="left">

Rule parameters
</td></tr><tr><td align="left">

`rule` → `version`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule version. The only possible value is `1`
</td></tr><tr><td align="left">

`rule` → `block`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Blocking parameters. If you specify `null`, blocking will not be performed. The request must contain `rule` → `block` → `severity`
</td></tr><tr><td align="left">

`rule` → `block` → `severity`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Block vulnerabilities or threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Notification emailing settings. If you specify `null`, notifications will not be sent. The request must contain `rule` → `notify` → `severity`
</td></tr><tr><td align="left">

`rule` → `notify` → `severity`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Notify about vulnerabilities or threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify` → `targets`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

List of addresses to which notifications will be sent
</td></tr><tr><td align="left">

`rule` → `whitelist`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

List of exclusions for the rule to skip
</td></tr><tr><td align="left">

`rule` → `whitelist` → `threats`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

[List of threats](#11978146571). Can be empty
</td></tr><tr><td align="left">

`rule` → `whitelist` → `binaries`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

List of executable files. Can be empty 
</td></tr><tr><td align="left">

`scope`
</td><td align="left">

Yes
</td><td align="left">

JSON object
</td><td align="left">

Rule scope
</td></tr><tr><td align="left">

`scope` → `version`
</td><td align="left">

Yes
</td><td align="left">

String
</td><td align="left">

Rule scope version. The only possible value is `1`
</td></tr><tr><td align="left">

`scope` → `image_names`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of the images that will be covered by the rule
</td></tr><tr><td align="left">

`scope` → `registries`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of registries containing images that will be covered by the rule
</td></tr><tr><td align="left">

`scope` → `clusters`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Cluster for which the rule will be created
</td></tr><tr><td align="left">

`scope` → `namespace`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of namespaces containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `pods`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of pods containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `containers`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of containers containing objects that will be covered by the rule. Can be empty
</td></tr><tr><td align="left">

`scope` → `nodes`
</td><td align="left">

Yes
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of nodes containing objects that will be covered by the rule. Can be empty
</td></tr></tbody></table>


**Response to a request**

In response to a successful request, the service returns an empty message with code 200.

**Example**

Request:

```
PATCH https://runtimeradar.example.com/api/v1/public-api/rule/rule-123
```

Request body:

```
{
  "name": "runtime rule",
  "rule": {
    "version": "1",
    "block": { "severity": "critical" },
    "notify": {
      "severity": "high",
      "targets": ["9f1c0b7a-2f44-4a1e-8e6d-5c2b7a9e0011"]
    },
    "whitelist": { "threats": [], "binaries": ["/usr/bin/curl"] }
  },
  "scope": { "version": "1", "clusters": ["prod-*"], "namespaces": ["default", "kube-system"] }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

[^6]: You can find out the identifier of a response rule by sending a request for the list of response rules.

### Deleting a response rule

You can delete a response rule using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **Delete response rules** permission.

Request method and URL:

```
DELETE <root API URL>/public-api/rule/{id}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the `id` path parameter (response rule identifier[^6]) with the String data type.

**Response to a request**

In response to a successful request, the service returns an empty message with code 200.

In response to an unsuccessful request, the service returns a message with an [error code](#9572838667).

**Example**

Request:

```
DELETE https://runtimeradar.example.com/api/v1/public-api/rule/rule-123
```

[^6]: You can find out the identifier of a response rule by sending a request for the list of response rules.

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### <a name="11949896587"></a>Listing response rules

You can fetch the paginated list of existing response rules in the system using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **View response rules** permission.

Request method and URL:

```
GET <root API URL>/public-api/rule/page/{page_num}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL has the required path parameter `page-num` (page number) with the String data type and the optional query parameters `page_size` (number of page entries) with the Int64 data type and `order` (sorting parameter) with the String data type.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Scheme of a response to a request to fetch a list of response rules</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`id`
</td><td align="left">

String
</td><td align="left">

Rule identifier
</td></tr><tr><td align="left">

`name`
</td><td align="left">

String
</td><td align="left">

Rule name
</td></tr><tr><td align="left">

`type`
</td><td align="left">

String
</td><td align="left">

Rule type: `TYPE_RUNTIME`
</td></tr><tr><td align="left">

`rule`
</td><td align="left">

Array of JSON objects
</td><td align="left">

Rule parameters
</td></tr><tr><td align="left">

`rule` → `version`
</td><td align="left">

String
</td><td align="left">

Rule version. The only possible value is `1`
</td></tr><tr><td align="left">

`rule` → `block`
</td><td align="left">

JSON object
</td><td align="left">

Blocking parameters. If the received value is `null`, blocking will not be performed. The response body also contains `rule` → `block` → `severity`
</td></tr><tr><td align="left">

`rule` → `block` → `severity`
</td><td align="left">

String
</td><td align="left">

Block vulnerabilities or threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify`
</td><td align="left">

JSON object
</td><td align="left">

Notification emailing settings. If the received value is `null`, a notification will not be sent. The response body also contains `rule` → `notify` → `severity`
</td></tr><tr><td align="left">

`rule` → `notify` → `severity`
</td><td align="left">

String
</td><td align="left">

Notify about vulnerabilities or threats of this severity and higher. Possible values:

* `low`. Low.
* `medium`. Medium.
* `high`. High.
* `critical`. Critical</td></tr><tr><td align="left">

`rule` → `notify` → `targets`
</td><td align="left">

Array of strings
</td><td align="left">

List of addresses to which notifications will be sent
</td></tr><tr><td align="left">

`rule` → `whitelist`
</td><td align="left">

JSON object
</td><td align="left">

List of exclusions
</td></tr><tr><td align="left">

`rule` → `whitelist` → `threats`
</td><td align="left">

Array of strings
</td><td align="left">

List of threats
</td></tr><tr><td align="left">

`rule` → `whitelist` → `binaries`
</td><td align="left">

Array of strings
</td><td align="left">

List of executable files
</td></tr><tr><td align="left">

`scope`
</td><td align="left">

JSON object
</td><td align="left">

Rule scope
</td></tr><tr><td align="left">

`scope` → `version`
</td><td align="left">

String
</td><td align="left">

Rule scope version. The only possible value is `1`
</td></tr><tr><td align="left">

`scope` → `image_names`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of the images covered by the rule
</td></tr><tr><td align="left">

`scope` → `registries`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of registries containing images that are be covered by the rule
</td></tr><tr><td align="left">

`scope` → `clusters`
</td><td align="left">

Array of strings
</td><td align="left">

Cluster for which the rule was created
</td></tr><tr><td align="left">

`scope` → `namespace`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of namespaces containing objects covered by the rule
</td></tr><tr><td align="left">

`scope` → `pods`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of pods containing objects covered by the rule
</td></tr><tr><td align="left">

`scope` → `containers`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of containers containing objects covered by the rule
</td></tr><tr><td align="left">

`scope` → `nodes`
</td><td align="left">

Array of strings
</td><td align="left">

Names or name templates of nodes containing objects covered by the rule
</td></tr><tr><td align="left">

`deleted`
</td><td align="left">

Boolean
</td><td align="left">

Shows if a rule is deleted. Possible values: `true` (deleted) and `false` (not deleted) 
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns an empty message with an [error code](#9572838667).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/rule/page/1
```

Response to a successful request:

```
{
  "total": 2,
  "rules": [
    {
      "id": "1a5f...",
      "name": "runtime rule",
      "rule": {
        "...": "..."
      },
      "type": "TYPE_RUNTIME",
      "scope": {
        "...": "..."
      }
    },
    {
      "id": "3c7e...",
      "name": "notify only",
      "rule": {
        "version": "1",
        "block": null,
        "notify": {
          "severity": "low",
          "targets": []
        },
        "whitelist": null
      },
      "type": "TYPE_RUNTIME",
      "scope": {
        "version": "1",
        "image_names": [],
        "registries": [],
        "clusters": [],
        "namespaces": [],
        "pods": [],
        "containers": [],
        "nodes": []
      }
    }
  ]
}
```

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching information about use of a notification template in a response rule

You can fetch information about use of a notification template in a response rule using a request to the public Runtime Radar API. To execute a request, you must have an access token with the **View response rules** permission.

Request method and URL:

```
GET <root API URL>/public-api/rule/notify-targets-in-use
```

The request header contains the `Access_Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the query parameter `targets` (list of target notification templates) with the Array of strings data type.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Scheme of a response to a request to fetch response rule data</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`in-use`
</td><td align="left">

Boolean
</td><td align="left">

Shows whether a notification template is used in a response rule. Possible values: `true` (used) and `false` (not used)
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns an empty message with an [error code](#9572838667).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/rule/notify-targets-in-use
```

Response to a successful request:

```
{
  "in_use": true
}
```

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### <a name="11978146571"></a>List of exclusions

This section contains a list of threats that you can specify as exclusions from response rules.

<table><caption>Threats that you can specify in the exclusion list for response rules with the TYPE_RUNTIME type</caption><colgroup><col style="width: 16.3%;"/><col style="width: 29.2%;"/><col style="width: 54.4%;"/></colgroup><thead><tr><th align="left">

Value
</th><th align="left">

Name
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`CS_RT_CRYPTOMINER`
</td><td align="left">

Cryptominer startup
</td><td align="left">

The detector detects if any known cryptominers were started or stopped. The latter may indicate an attacker's attempt to get rid of competing loads
</td></tr><tr><td align="left">

`CS_RT_CVE_2022_0492`
</td><td align="left">

Container isolation vulnerability exploitation (CVE-2022-0492)
</td><td align="left">

The detector detects attempts to modify the `notify_on_release` and `release_agent` files, which may indicate an attempt to exploit vulnerability CVE-2022-0492. Vulnerability exploitation may result in privilege escalation and breakout from the isolated environment of a container
</td></tr><tr><td align="left">

`CS_RT_DOWNLOAD_TOOLS`
</td><td align="left">

Network activity: file downloading
</td><td align="left">

The detector detects suspicious network activity of utilities used for file transferring and downloading
</td></tr><tr><td align="left">

`CS_RT_HACK_TOOLS`
</td><td align="left">

Hacking tools
</td><td align="left">

The detector detects startups of tools typically associated with malicious activities, such as automated exploitation of known vulnerabilities, brute-force attacks, reconnaissance, and information collection
</td></tr><tr><td align="left">

`CS_RT_K8S_SA_TOKEN_READ`
</td><td align="left">

Reading of Kubernetes authentication token
</td><td align="left">

The detector detects reading of the Kubernetes authentication token, which may indicate that the Kubernetes account was compromised
</td></tr><tr><td align="left">

`CS_RT_KERNEL_MODULE`
</td><td align="left">

Kernel module uploading
</td><td align="left">

The detector detects loading of kernel modules that may contain malicious code to attack the target system
</td></tr><tr><td align="left">

`CS_RT_LD_PRELOAD`
</td><td align="left">

Code injection using LD_PRELOAD
</td><td align="left">

The detector detects if the `/etc/ld.so.preload` file was edited, which may indicate an attacker's attempt to inject code using dynamic library preloading
</td></tr><tr><td align="left">

`CS_RT_MOUNT_DEV`
</td><td align="left">

Device mounting from `/dev` directory
</td><td align="left">

The detector detects the `mount()` system calls to mount devices from the `/dev` directory, which may indicate an attacker's attempt to gain access to the file system of the container host. The detector assigns low severity to unsuccessful attempts and high severity to successful ones
</td></tr><tr><td align="left">

`CS_RT_PRIV_ECS`
</td><td align="left">

Privilege escalation
</td><td align="left">

The detector detects if a root user started a process or if a process called the `commit_creds` function with the `UID/EUID == 0` or `GID/EGID == 0` parameter to gain root privileges. This behavior may indicate an attacker's attempt to elevate their privileges
</td></tr><tr><td align="left">

`CS_RT_PROCESS_CAP_RAISE`
</td><td align="left">

Privilege escalation: capabilities
</td><td align="left">

This detector detects processes running with excessive capabilities, which may indicate an attacker's attempt to escalate their privileges in the system
</td></tr><tr><td align="left">

`CS_RT_RAW_SOCKETS`
</td><td align="left">

Creation of raw network socket
</td><td align="left">

The detector detects if a raw network socket was created, which may indicate network traffic interception or reconnaissance
</td></tr><tr><td align="left">

`CS_RT_RECON_TOOLS`
</td><td align="left">

Use of network reconnaissance tools
</td><td align="left">

The detector detects attempts to start applications that use the libpcap library. Such attempts may indicate use of network reconnaissance tools
</td></tr><tr><td align="left">

`CS_RT_REVERSE_SHELL_CREATE`
</td><td align="left">

Reverse shell creation
</td><td align="left">

The detector detects signs that a reverse shell was created
</td></tr><tr><td align="left">

`CS_RT_REVERSE_SHELL_USE`
</td><td align="left">

Reverse shell use
</td><td align="left">

The detector detects signs that a reverse shell is used
</td></tr><tr><td align="left">

`CS_RT_SCHED_TASK_MOD`
</td><td align="left">

Suspicious changes in task scheduler configuration files
</td><td align="left">

The detector detects changes in configuration files of a task scheduler (for example, cron) if it was installed in a container or added by an attacker
</td></tr><tr><td align="left">

`CS_RT_SSH_TUNNEL_CREATE`
</td><td align="left">

SSH tunnel creation
</td><td align="left">

The detector detects if a local or remote network tunnel was created using an SSH service
</td></tr><tr><td align="left">

`CS_RT_SSH_TUNNEL_USE`
</td><td align="left">

Use of SSH tunnel
</td><td align="left">

The detector detects network activity related to network tunneling of a local SSH service
</td></tr><tr><td align="left">

`CS_RT_SUSP_FILE_READ`
</td><td align="left">

Suspicious reading of sensitive system files
</td><td align="left">

The detector detects suspicious reading of system files with utilities from uncommon directories, such as `/tmp` and `/home`, which may indicate that system configuration data is being collected
</td></tr><tr><td align="left">

`CS_RT_SUSP_FILE_WRITE`
</td><td align="left">

Suspicious changes in sensitive system files
</td><td align="left">

The detector detects suspicious changes in system files, such as `/etc/passwd` and `/etc/shadow`, which may indicate an attacker's attempt to alter the system configuration or gain privileged access
</td></tr><tr><td align="left">

`CS_RT_SUSP_SHELL`
</td><td align="left">

Suspicious startup of command shell
</td><td align="left">

The detector detects suspicious startups of the command shell. They may indicate attempts to exploit RCE vulnerabilities, open a remote communication channel, or use the GTFOBins utilities for privilege escalation
</td></tr><tr><td align="left">

`CS_RT_BASE64_DECODE_RUN`
</td><td align="left">

Base64: data decoding
</td><td align="left">

The detector detects if utilities for decoding arbitrary text from Base64 are used
</td></tr><tr><td align="left">

`CS_RT_BIN_PERM_RAISE`
</td><td align="left">

Change of file access permissions
</td><td align="left">

The detector detects if permissions to execute files in the `boot`, `dev`, `home`, `media`, `mnt`, `run`, `sys`, `tmp`, and `var` directories were granted
</td></tr><tr><td align="left">

`CS_RT_CORE_PATTERN_MODIFY`
</td><td align="left">

Change of generation settings for process dump file
</td><td align="left">

The detector detects if the `core_pattern` file was edited in the `procfs` file system, which may indicate an attacker's attempt to elevate their privileges and leave the isolated container environment
</td></tr><tr><td align="left">

`CS_RT_CVE_2025_1974`
</td><td align="left">

Ingress Nightmare vulnerability (malicious code execution)
</td><td align="left">

The detector detects if a shared library was loaded from the directory `/tmp/nginx/client-body/`, which may indicate that malicious code was loaded and executed to exploit the Ingress Nightmare vulnerability (CVE-2025-1974)
</td></tr><tr><td align="left">

`CS_RT_IO_URING_CREATE`
</td><td align="left">

Creation of `io_uring` interface instance
</td><td align="left">

The detector detects if an `io_uring` interface instance was created, which may indicate an attacker's attempt to conceal their actions by executing file operations and network operations without the corresponding system calls
</td></tr><tr><td align="left">

`CS_RT_SSH_KEY_MODIFY`
</td><td align="left">

Suspicious change of SSH keys
</td><td align="left">

The detector detects suspicious changes in files that contain SSH keys, which may indicate that an existing account was compromised or that an attacker attempted to achieve persistence on the system
</td></tr><tr><td align="left">

`CS_RT_FILELESS_EXEC`
</td><td align="left">

Fileless process execution
</td><td align="left">

The detector detects if a process was executed without an executable file in the file system, which may indicate an attacker's attempt to conceal their activity
</td></tr><tr><td align="left">

`CS_RT_HARDLINK_CREATE`
</td><td align="left">

Creation of hard links to system files
</td><td align="left">

The detector detects if hard links to system files were created, which may indicate an attacker's attempt to bypass the existing file monitoring rules
</td></tr><tr><td align="left">

`CS_RT_FIFO_FILE_CREATE`
</td><td align="left">

Creation of named pipe file
</td><td align="left">

The detector detects signs that a named pipe file was created. Applications can use such a file to bypass existing audit policies when exchanging data—for example, to create a reverse shell
</td></tr><tr><td align="left">

`CS_RT_OPENSSL_USE`
</td><td align="left">

OpenSSL use for malicious purposes
</td><td align="left">

The detector detects if the OpenSSL utility was used to perform an attacker's scenarios: malicious code loading, reading or writing of arbitrary files, or data exchange via a TLS or SSL server or client
</td></tr><tr><td align="left">

`CS_RT_PTRACE_USE`
</td><td align="left">

Code injection into executable process through ptrace
</td><td align="left">

The detector detects code injection into an executable process through the ptrace() system call, which may indicate attempts to masquerade as a legitimate process, access the process memory and system or network resources, and escalate privileges
</td></tr><tr><td align="left">

`CS_RT_SUSP_TOOLS`
</td><td align="left">

Start of suspicious utilities and executable files
</td><td align="left">

The detector detects if suspicious system utilities were started and files were executed from potentially harmful directories (for example, `/tmp`, `/sys`, or `/boot`). Such activity may indicate an attempt to carry out an LOLBins attack as well as execution of malicious code loaded by an attacker
</td></tr><tr><td align="left">

`CS_RT_CVE_2026_31431`
</td><td align="left">

Creation of AF_ALG socket
</td><td align="left">

The detector detects if an AF_ALG socket was created, which may indicate an attempt to exploit the Copy Fail (CVE-2026-31431) vulnerability
</td></tr></tbody></table>


### <a name="9572838667"></a>Error messages

This section describes general errors that may occur when you manage Runtime Radar response rules using the REST API. You can find errors specific to a particular type of request in sections with the request descriptions.

Possible error codes and their meaning:
* 400 (Bad Request). Syntax error in the request.
* 401 (Unauthorized). Access token not found.
* 403 (Forbidden). Access denied. Not enough permissions to execute a request.
* 404 (Not Found). Requested URL not found.
* 500 (Internal Server Error). Internal server error.
* 502 (Bad Gateway). Server unavailable or invalid server response.

## Fetching information about runtime events

Request method and URL:

```
GET <root API URL>/public-api/runtime-event/slice/{direction}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL has the required path parameter `direction` with the String type and the optional query parameters `cursor` (event set cursor) with the String (date-time) type and `slice_size` (number of page entries) with the Int64 type.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

In response to an unsuccessful request, the service returns an empty message with an error code.

Possible error codes and their meaning:
* 400 (Bad Request). Syntax error in the request.
* 500 (Internal Server Error). Internal server error.

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/runtime-event/slice/right?cursor=2025-08-11T05:02:26.271Z&slice_size=15
```

Response to a successful request:

```
{
    "runtime_events": [
        {
            "id": "74d23240-a09a-41d8-a816-9c7a171ca1b3",
            "tetragon_version": "v1.3.0",
            "event": {
                "process_kprobe": {
                    "process": {
                        "exec_id": "YXNoL...E5NzYyODg=",
                        "pid": 1976288,
                        "uid": 0,
                        "cwd": "/",
                        "binary": "/usr/bin/ls",
                        "arguments": "-la",
                        "flags": "execve rootcwd clone",
                        "start_time": "2025-08-28T04:09:29.327694437Z",
                        "auid": 4294967295,
                        "pod": {
                            "namespace": "cs",
                            "name": "admission-controller-5...4-b...p",
                            "container": {
                                "id": "containerd://f653...efcbac5",
                                "name": "admission-controller",
                                "image": {
                                    "id": "docker-cs-dev.devops.example.com/admission-controller@sha256:a6d92...115237",
                                    dev.devops.example.com/admission-controller:v0.1.0"
                                },
                                "start_time": "2025-08-27T11:14:20Z",
                                "pid": 30,
                                "maybe_exec_probe": false
                            },
                            "pod_labels": {
                                "app.kubernetes.io/instance": "cs",
                                "app.kubernetes.io/managed-by": "Helm",
                                "app.kubernetes.io/name": "admission-controller",
                                "app.kubernetes.io/version": "v0.1.0",
                                "helm.sh/chart": "admission-controller-v0.1.0",
                                "pod-template-hash": "5fc57d9594"
                            },
                            "workload": "admission-controller",
                            "workload_kind": "Deployment"
                        },
                        "docker": "f6539c3cbafe30bd25dc4a8260aeb27",
                        "parent_exec_id": "YXNoLW...xODY=",
                        "refcnt": 1,
                        "cap": {
                            "permitted": [
                                "CAP_CHOWN",
                                ...,
                                "CAP_SETFCAP"
                            ],
                            "effective": [
                                "CAP_CHOWN",
                                ...,
                                "CAP_SETFCAP"
                            ],
                            "inheritable": []
                        },
                        "ns": {
                            "uts": {
                                "inum": 4026536148,
                                "is_host": false
                            },
                            "ipc": {
                                "inum": 4026536149,
                                "is_host": false
                            },
                            "mnt": {
                                "inum": 4026536223,
                                "is_host": false
                            },
                            "pid": {
                                "inum": 4026536224,
                                "is_host": false
                            },
                            "pid_for_children": {
                                "inum": 4026536224,
                                "is_host": false
                            },
                            "net": {
                                "inum": 4026534370,
                                "is_host": false
                            },
                            "time": {
                                "inum": 4026531834,
                                "is_host": true
                            },
                            "time_for_children": {
                                "inum": 4026531834,
                                "is_host": true
                            },
                            "cgroup": {
                                "inum": 4026536225,
                                "is_host": false
                            },
                            "user": {
                                "inum": 4026531837,
                                "is_host": true
                            }
                        },
                        "tid": 1976288,
                        "process_credentials": {
                            "uid": 0,
                            "gid": 0,
                            "euid": 0,
                            "egid": 0,
                            "suid": 0,
                            "sgid": 0,
                            "fsuid": 0,
                            "fsgid": 0,
                            "securebits": [],
                            "caps": null,
                            "user_ns": null
                        },
                        "binary_properties": null,
                        "user": null,
                        "in_init_tree": false
                    },
                    "parent": {
                        "exec_id": "YXNoL...xODY=",
                        "pid": 1976186,
                        "uid": 0,
                        "cwd": "/",
                        "binary": "/usr/bin/bash",
                        "arguments": "",
                        "flags": "execve rootcwd clone",
                        "start_time": "2025-08-28T04:09:25.977838295Z",
                        "auid": 4294967295,
                        "pod": {
                            "namespace": "cs",
                            "name": "admission-controller-5...4-b...p",
                            "container": {
                                "id": "containerd://f6539...cbac5",
                                "name": "admission-controller",
                                "image": {
                                    "id": "docker-cs-dev.devops.example.com/admission-controller@sha256:a6d92...15237",
                                    "name": "docker-cs-dev.devops.example.com/admission-controller:v0.1.0"
                                },
                                "start_time": "2025-08-27T11:14:20Z",
                                "pid": 29,
                                "maybe_exec_probe": false
                            },
                            "pod_labels": {
                                "app.kubernetes.io/instance": "cs",
                                "app.kubernetes.io/managed-by": "Helm",
                                "app.kubernetes.io/name": "admission-controller",
                                "app.kubernetes.io/version": "v0.1.0",
                                "helm.sh/chart": "admission-controller-v0.1.0",
                                "pod-template-hash": "5fc57d9594"
                            },
                            "workload": "admission-controller",
                            "workload_kind": "Deployment"
                        },
                        "docker": "f6539...aeb27",
                        "parent_exec_id": "YXNoL...xODU=",
                        "refcnt": 0,
                        "cap": {
                            "permitted": [
                                "CAP_CHOWN",
                                ...,
                                "CAP_SETFCAP"
                            ],
                            "effective": [
                                "CAP_CHOWN",
                                ...,
                                "CAP_SETFCAP"
                            ],
                            "inheritable": []
                        },
                        "ns": {
                            "uts": {
                                "inum": 4026536148,
                                "is_host": false
                            },
                            "ipc": {
                                "inum": 4026536149,
                                "is_host": false
                            },
                            "mnt": {
                                "inum": 4026536223,
                                "is_host": false
                            },
                            "pid": {
                                "inum": 4026536224,
                                "is_host": false
                            },
                            "pid_for_children": {
                                "inum": 4026536224,
                                "is_host": false
                            },
                            "net": {
                                "inum": 4026534370,
                                "is_host": false
                            },
                            "time": {
                                "inum": 4026531834,
                                "is_host": true
                            },
                            "time_for_children": {
                                "inum": 4026531834,
                                "is_host": true
                            },
                            "cgroup": {
                                "inum": 4026536225,
                                "is_host": false
                            },
                            "user": {
                                "inum": 4026531837,
                                "is_host": true
                            }
                        },
                        "tid": 1976186,
                        "process_credentials": {
                            "uid": 0,
                            "gid": 0,
                            "euid": 0,
                            "egid": 0,
                            "suid": 0,
                            "sgid": 0,
                            "fsuid": 0,
                            "fsgid": 0,
                            "securebits": [],
                            "caps": null,
                            "user_ns": null
                        },
                        "binary_properties": null,
                        "user": null,
                        "in_init_tree": false
                    },
                    "function_name": "security_file_permission",
                    "args": [
                        {
                            "file_arg": {
                                "mount": "",
                                "path": "/proc/filesystems",
                                "flags": "",
                                "permission": "-r--r--r--"
                            },
                            "label": ""
                        },
                        {
                            "int_arg": 4,
                            "label": ""
                        }
                    ],
                    "return": {
                        "int_arg": 0,
                        "label": ""
                    },
                    "action": "KPROBE_ACTION_POST",
                    "kernel_stack_trace": [],
                    "policy_name": "file-monitoring",
                    "return_action": "KPROBE_ACTION_POST",
                    "message": "",
                    "tags": [],
                    "user_stack_trace": []
                },
                "node_name": "cs.example.com",
                "time": "2025-08-28T04:09:29.335381806Z",
                "aggregation_info": null,
                "cluster_name": ""
            },
            "threats": [],
            "detect_errors": [],
            "is_incident": false,
            "incident_severity": "none",
            "block_by": [],
            "notify_by": []
        }
    ],
    "left_cursor": "2025-08-28T04:09:29.335381806Z",
    "right_cursor": "2025-08-28T04:09:25.950950032Z"
}
```

Response to an unsuccessful request:

```
{
  "code": 16,
  "message": "access token not found",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "UNAUTHENTICATED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### <a name="9706701835"></a>Derector schema

Describes a detector.

<table><caption>Fields in the detector schema</caption><colgroup><col style="width: 23.8%;"/><col style="width: 24.1%;"/><col style="width: 52.0%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`id`
</td><td align="left">

String
</td><td align="left">

Unique detector identifier
</td></tr><tr><td align="left">

`name`
</td><td align="left">

String
</td><td align="left">

Detector name
</td></tr><tr><td align="left">

`version`
</td><td align="left">

Int64
</td><td align="left">

Detector version
</td></tr><tr><td align="left">

`description`
</td><td align="left">

String
</td><td align="left">

Detector description
</td></tr><tr><td align="left">

`author`
</td><td align="left">

String
</td><td align="left">

Detector author
</td></tr><tr><td align="left">

`contact`
</td><td align="left">

String
</td><td align="left">

Contact details
</td></tr></tbody></table>


### Detect_errors schema

Describes found errors.

<table><caption>Fields in the detect_errors schema</caption><colgroup><col style="width: 23.8%;"/><col style="width: 24.1%;"/><col style="width: 52.0%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`detector`
</td><td align="left">

JSON object
</td><td align="left">

Detector used for a check. Described using the [detector schema](#9706701835)
</td></tr><tr><td align="left">

`error`
</td><td align="left">

String
</td><td align="left">

Error
</td></tr></tbody></table>


## Fetching information about cluster objects

You can fetch information about nodes and pods in a cluster using the public Runtime Radar API. All requests for managing response rules are authorized using the `access_token` setting ([access token](#10359847435)) setting with the String data type.

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching information about a cluster node

You can fetch information about a cluster node using a request to the public Runtime Radar API. To send a request, you must create an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/node?name=<node_name>
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the query parameter `name` (node name) with the Array of strings data type.

**Response to a request**

In response to a successful request, the service returns code 200. For details about fields in a response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#node-v1-core).

In response to an unsuccessful request, the service returns a message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/node?name=node-1
```

Response to a successful request:

To view an example of a successful response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#node-v1-core).

Response to an unsuccessful request:

```
{
  "code": 16,
  "message": "access token not found",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "UNAUTHENTICATED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching cluster node metadata

You can fetch sorted and filtered metadata of cluster nodes using a request to the public Runtime Radar API. To send a request, you need an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/node/list
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the following query parameters:
* `names` (list of node names) with the Array of strings data type.
* `sort.field` (field for sorting) with the String data type. Possible values: `created_at` (creation date and time) and `name` (node name).
* `sort.key` (sorting order) with the String data type. Possible values: `desc` (reverse sorting); otherwise, forward sorting.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Scheme of a response to a request to fetch node metadata</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`nodes`
</td><td align="left">

Array of JSON objects
</td><td align="left">

List of nodes returned by the request
</td></tr><tr><td align="left">

`nodes` → `hostname`
</td><td align="left">

String
</td><td align="left">

Node name
</td></tr><tr><td align="left">

`nodes` → `ip`
</td><td align="left">

String
</td><td align="left">

Node IP address
</td></tr><tr><td align="left">

`nodes` → `name`
</td><td align="left">

String
</td><td align="left">

Short node name
</td></tr><tr><td align="left">

`nodes` → `uid`
</td><td align="left">

String
</td><td align="left">

Node UUID
</td></tr><tr><td align="left">

`total`
</td><td align="left">

Integer
</td><td align="left">

Number of nodes in the received response
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns a message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/node/list?names=node-1&names=node-2&sort.field=name&sort.key=asc
```

Response to a successful request:

```
{
  "nodes": [
    {
      "hostname": "worker-node-01",
      "ip": "192.0.2.0",
      "name": "node-01",
      "uid": "c3f5b9e2-7a4d-11ee-b962-0242ac120002"
    },
    {
      "hostname": "worker-node-02",
      "ip": "192.0.2.1",
      "name": "node-02",
      "uid": "d1a9c3f8-7a4d-11ee-b962-0242ac120002"
    },
    {
      "hostname": "worker-node-03",
      "ip": "192.0.2.2",
      "name": "node-03",
      "uid": "e4b1d6a5-7a4d-11ee-b962-0242ac120002"
    }
  ],
  "total": 3
}
```

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching a paginated list of cluster nodes

You can fetch a paginated list of cluster nodes using a request to the public Runtime Radar API. To send a request, you need an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/node/page/{page_num}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the required path setting `page-num` (page number) with the Integer data type and the following query parameters:
* `page_size` (number of records per page) with the Integer data type.
* `names` (list of node names) with the Array of strings data type.
* `sort.field` (field for sorting) with the String data type. Possible values: `created_at` (creation date and time) and `name` (node name).
* `sort.key` (sorting order) with the String data type. Possible values: `desc` (reverse sorting); otherwise, forward sorting.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Response scheme to a request to fetch a paginated list of nodes</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`nodes`
</td><td align="left">

Array of JSON objects
</td><td align="left">

List of nodes returned by the request
</td></tr><tr><td align="left">

`nodes` → `node`
</td><td align="left">

JSON object
</td><td align="left">

For details about node fields, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#node-v1-core)
</td></tr><tr><td align="left">

`total`
</td><td align="left">

Integer
</td><td align="left">

Number of nodes in the received response
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns a message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/node/page/1
```

Response to a successful request:

```
{
  "nodes": [
    node{
    ...
    },
    node{
    ...
    },
    node{
    ...
    },
  ],
  "total": 3
}
```

For details about node parameters returned in a response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#node-v1-core).

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching information about a cluster pod

You can fetch information about a cluster pod using a request to the public Runtime Radar API. To send a request, you must create an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/pod?name=<pod name>&namespace=<namespace>
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the query parameters `name` (pod name) with the String data type and `namespace` with the String data type.

**Response to a request**

In response to a successful request, the service returns code 200. For details about fields in a response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#pod-v1-core).

In response to an unsuccessful request, the service returns a message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/pod?name=nfs-subdir-external-provisioner-74d6f96865-5l86s&namespace=default
```

Response to a successful request:

To view an example of a successful response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#pod-v1-core).

Response to an unsuccessful request:

```
{
  "code": 16,
  "message": "access token not found",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "UNAUTHENTICATED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching cluster pod metadata

You can fetch sorted and filtered metadata of cluster pods using a request to the public Runtime Radar API. To send a request, you need an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/pod/list
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the following query parameters:
* `namespaces` (list of namespaces) with the Array of strings data type.
* `nodes` (list of node names) with the Array of strings data type.
* `pods` (list of pod names) with the Array of strings data type.
* `containers` (list of container names) with the Array of strings data type.
* `sort.field` (field for sorting) with the String data type. Possible values: `created_at` (creation date and time) and `name` (list of pod names).
* `sort.key` (sorting order) with the String data type. Possible values: `desc` (reverse sorting); otherwise, forward sorting.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Scheme of a response to a request to fetch metadata on pods in the cluster</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`pods`
</td><td align="left">

Array of JSON objects
</td><td align="left">

List of pods returned by the request
</td></tr><tr><td align="left">

`pods` → `containers`
</td><td align="left">

Array of strings
</td><td align="left">

Names of containers running in a pod
</td></tr><tr><td align="left">

`pods` → `name`
</td><td align="left">

String
</td><td align="left">

Pod name
</td></tr><tr><td align="left">

`pods` → `namespace`
</td><td align="left">

String
</td><td align="left">

Pod namespace
</td></tr><tr><td align="left">

`pods` → `nodeName`
</td><td align="left">

String
</td><td align="left">

Name of the node on which the pod is running
</td></tr><tr><td align="left">

`pods` → `phase`
</td><td align="left">

String
</td><td align="left">

Current pod status
</td></tr><tr><td align="left">

`pods` → `uid`
</td><td align="left">

String
</td><td align="left">

Pod UUID
</td></tr><tr><td align="left">

`total`
</td><td align="left">

Integer
</td><td align="left">

Number of pods in the received response
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns an empty message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/pod/list?namespaces=default&nodes=node-1&pods=pod-1&containers=nginx&sort.field=name&sort.key=asc
```

Response to a successful request:

```
{
  "pods": [
    {
      "containers": [
        "nginx",
        "sidecar-exporter"
      ],
      "name": "frontend-7d9c8b9b5f-kt2hf",
      "namespace": "production",
      "nodeName": "node-3a2b1c",
      "phase": "Running",
      "uid": "c1d2e3f4-5678-90ab-cdef-1234567890ab"
    },
    {
      "containers": [
        "redis",
        "redis-exporter"
      ],
      "name": "cache-redis-0",
      "namespace": "staging",
      "nodeName": "node-7e8f9g",
      "phase": "Pending",
      "uid": "a1b2c3d4-1111-2222-3333-444455556666"
    },
    {
      "containers": [
        "api-server"
      ],
      "name": "backend-58d4f7c9f9-vz9kp",
      "namespace": "default",
      "nodeName": "node-0d1e2f",
      "phase": "Succeeded",
      "uid": "f9e8d7c6-b5a4-3210-9abc-def012345678"
    }
  ],
  "total": 3
}
```

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### Fetching a paginated list of cluster pods

You can fetch a paginated list of cluster pods using a request to the public Runtime Radar API. To send a request, you need an access token with the **View information about cluster objects** permission.

Request method and URL:

```
GET <root API URL>/public-api/pode/page/{page_num}
```

The request header contains the `Access-Token` parameter ([access token](#10359847435)) with the String data type. The request URL contains the required path setting `page-num` (page number) with the Integer data type and the following query parameters:
* `page_size` (number of records per page) with the Integer data type.
* `namespaces` (list of namespaces) with the Array of strings data type.
* `nodes` (list of node names) with the Array of strings data type.
* `pods` (list of pod names) with the Array of strings data type.
* `containers` (list of container names) with the Array of strings data type.
* `sort.field` (field for sorting) with the String data type. Possible values: `created_at` (creation date and time) and `name` (pod name).
* `sort.key` (sorting order) with the String data type. Possible values: `desc` (reverse sorting); otherwise, forward sorting.

**Response to a request**

In response to a successful request, the service returns code 200. Response fields are described in the table below.

<table><caption>Response scheme to a request to fetch a paginated list of pods</caption><colgroup><col style="width: 20.0%;"/><col style="width: 18.8%;"/><col style="width: 61.1%;"/></colgroup><thead><tr><th align="left">

Field
</th><th align="left">

Type or format
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`pods`
</td><td align="left">

Array of JSON objects
</td><td align="left">

List of pods returned by the request
</td></tr><tr><td align="left">

`pods`→ `pod`
</td><td align="left">

JSON object
</td><td align="left">

For details about pod fields, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#objectmeta-v1-meta)
</td></tr><tr><td align="left">

`total`
</td><td align="left">

Integer
</td><td align="left">

Number of pods in the received response
</td></tr></tbody></table>


In response to an unsuccessful request, the service returns a message with an [error code](#11006622731).

**Example**

Request:

```
GET https://runtimeradar.example.com/api/v1/public-api/pod/page/1
```

Response to a successful request:

```
{
  "pods": [
    pod{
    ...
    },
    pod{
    ...
    },
    pod{
    ...
    },
  ],
  "total": 3
}
```

For details about pod parameters returned in a response to a request, visit [kubernetes.io/docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#pod-v1-core).

Response to an unsuccessful request:

```
{
  "code": 7,
  "message": "permission denied",
  "details": {
    "@type": "type.googleapis.com/google.rpc.ErrorInfo",
    "reason": "PERMISSION_DENIED"
  }
}
```

[^1]: You can create an access token in the Runtime Radar web interface. For more information about managing access tokens, see the Administrator Guide or the User Guide.

### <a name="11006622731"></a>Error messages

This section describes general errors that may occur when you manage nodes and pods using the Runtime Radar REST API. You can find errors specific to a particular type of request in sections with the request descriptions.

Possible error codes and their meaning:
* 400 (Bad Request). Syntax error in the request.
* 401 (Unauthorized). Access token not found.
* 403 (Forbidden). Access denied. Not enough permissions to execute a request.
* 500 (Internal Server Error). Internal server error.
