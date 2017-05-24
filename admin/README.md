### API

#### Configuration

##### GET /api/configs

Response:

```json
{
  "configs": [
    {
      "namespaces": {
        "test.namespace": {
          "name": "test.namespace",
          "buckets": {
            ...
          }
        }
      },
      "version": 4,
      "user": "quotaservice",
      "date": 1489427115
    },
    ...
  ]
}
```

##### GET /api

Response:

```json
{
  "namespaces": {
    "test.namespace": {
      "name": "test.namespace",
      "buckets": {
        ...
      }
    }
  },
  "version": 4,
  "user": "quotaservice",
  "date": 1489427115
}
```

##### POST /api

Request:

```json
{
  "namespaces": {
    "test.namespace": {
      "buckets": {
        "xyz": {
          "name": "xyz",
          "size": 1000
        }
      }
    }
  }
}
```

Response:

```
200 OK

{}
```

Error response:

```
500 Internal Server Error

{"description":"invalid character '}' after top-level value","error":"Internal Server Error"}
```

##### GET /api/{namespace}

Response:

```json
{
  "name": "test.namespace",
  "buckets": {
    "xyz": {
      "name": "xyz",
      "namespace": "test.namespace",
      "size": 1000,
      "fill_rate": 50,
      "wait_timeout_millis": 1000,
      "max_idle_millis": -1,
      "max_debt_millis": 10000,
      "max_tokens_per_request": 50
    }
  }
}
```

Error response:

```
404 Not Found

{"description":"Unable to locate namespace null","error":"Not Found"}
```

##### POST /api/{namespace}

Request `POST /api/new.namespace`:

```json

{
  "buckets": {
    "bar": {
      "size": 10000
    }
  }
}
```

Response:

```
200 OK

{}
```

Error response:

```
500 Internal Server Error

{"description":"Namespace new.namespace already exists.","error":"Internal Server Error"}
```

##### PUT /api/{namespace}

Request `PUT /api/new.namespace`:

```json
{
  "buckets": {
    "bar": {
      "size": 10
    }
  }
}
```

Response:

```
200 OK

{}
```

##### DELETE /api/{namespace}

Response:

```
200 OK

{}
```

Error response:

```
400 Bad Request

{"description":"No such namespace new.namespace","error":"Bad Request"}
```

##### GET /api/{namespace}/{bucket}

Response:

```json
{
  "name": "xyz",
  "namespace": "test.namespace2",
  "size": 100,
  "fill_rate": 50,
  "wait_timeout_millis": 1000,
  "max_idle_millis": -1,
  "max_debt_millis": 10000,
  "max_tokens_per_request": 50
}
```

##### PUT /api/{namespace}/{bucket}

Request:

```json
{
  "size": 1000
}
```

Response:

```
200 OK

{}
```

##### POST /api/{namespace}/{bucket}

Request `POST /api/test.namespace2/abc`:

```json
{
  "fill_rate": 10000
}
```

Response:

```
200 OK

{}
```

Error response:

```
500 Internal Server Error

{"description":"Bucket xyz already exists","error":"Internal Server Error"}
```

##### DELETE /api/{namespace}/{bucket}

Response:

```
200 OK

{}
```

#### Stats

##### GET /api/stats/{namespace}

Response:

```json
{
  "namespace": "test.namespace",
  "topHits": [
    {
      "bucket": "x.y.z",
      "value": 1000
    },
    ...
  ],
  "topMisses": [ ]
}
```
