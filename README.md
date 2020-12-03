## Motivation

The `bff` (Backend for Frontend) proxy was built out of the requirement to have an API-aware proxy that is capable of routing, filtering, verifying and modifing HTTP request and response.

It is built on top of [Google's Martian](https://github.com/google/martian) proxy framework, therefore `bff` supports the same [built-in modifiers](https://github.com/google/martian/wiki/Modifier-Reference) as martian does by default.

In addition to that, it provides a modifier to fetch and aggregate remote resources. Additionally, modifiers to merge and patch request and responses are provided out of the box.

These modifiers can be composed together to solve most use-cases that a BFF service may need.

You can learn more about the BFF cloud pattern here:

- https://www.thoughtworks.com/insights/blog/bff-soundcloud
- https://samnewman.io/patterns/architectural/bff/

## Install

### Executable

```sh
# make a temp directory
cd $(mktemp -d)

# download the bff executable into it
curl -sfL https://github.com/imranismail/bff/releases/download/v0.4.2/bff_0.4.2_Linux_x86_64.tar.gz | tar xvz

# move it into $PATH dir
mv bff /usr/local/bin

# test it
bff --help
```

### Container

- [Github Container Registry](https://github.com/users/imranismail/packages/container/package/bff)

## Supported Flags

```
  -c, --config string      config file (default is $XDG_CONFIG_HOME/bff/config.yaml)
  -h, --help               help for bff
  -i, --insecure           Skip TLS verify
  -p, --port string        Port to run the server on (default "5000")
  -u, --url string         Proxy url
  -v, --verbosity int      Verbosity
```

## Usage

The proxy is configured with a YAML configuration file. The file path can be set using the `--config` flag, it defaults to `$XDG_CONFIG_HOME/bff/config.yaml`

### Executable

```sh
bff --insecure=false --port=5000 --verbosity=3 <<YAML
# fetch resources concurrently
- body.MultiFetcher:
    resources:
      - body.JSONResource:
          method: GET
          url: https://jsonplaceholder.typicode.com/users/1
          behavior: replace # replaces upstream http response
          modifier:
            status.Verifier:
              statusCode: 200 # verify that the resource returns 200 status code
      - body.JSONResource:
          method: GET
          url: https://jsonplaceholder.typicode.com/users/1/todos
          behavior: merge # merge with the previous resource
          group: todos # group this response into "todos" key
          modifier:
            status.Verifier:
              statusCode: 200 # verify that the resource returns 200 status code
- body.JSONPatch:
    scope: [response]
    patch:
      - {op: move, from: /todos, path: /Todos}
YAML
```

### Container

```sh
# make a temp directory
cd $(mktemp -d)

# create config file
cat > config.yml <<EOF
modifiers: |-
  # fetch resources concurrently
  - body.MultiFetcher:
      resources:
        - body.JSONResource:
            method: GET
            url: https://jsonplaceholder.typicode.com/users/1
            behavior: replace # replaces upstream http response
            modifier:
              status.Verifier:
                statusCode: 200 # verify that the resource returns 200 status code
        - body.JSONResource:
            method: GET
            url: https://jsonplaceholder.typicode.com/users/1/todos
            behavior: merge # merge with the previous resource
            group: todos # group this response into "todos" key
            modifier:
              status.Verifier:
                statusCode: 200 # verify that the resource returns 200 status code
  - body.JSONPatch:
      scope: [response]
      patch:
        - {op: move, from: /todos, path: /Todos}
EOF

# run it
docker run --rm -it -v $(pwd)/config.yml:/srv/config.yml ghcr.io/imranismail/bff:latest
```

## Config Reference

### `config.yml`

```yaml
# env: BFF_INSECURE
# flag: --insecure -i
# type: bool
# required: false
# default: false
insecure: false

# env: BFF_PORT
# flag: -p --port
# type: int
# required: false
# default: 5000
port: 5000

# env: BFF_URL
# flag: -u --url
# type: string
# required: false
url: ""

# env: BFF_VERBOSITY
# flag: -v --verbosity
# type: int
# required: false
# default: 2
# options:
#   0: nothiing
#   1: error
#   2: info
#   3: debug
verbosity: 2

# env: BFF_MODIFIERS
# flag: N/A instead it can be set via linux pipe. example: `cat modifiers.yaml | bff` or `bff <<EOF ...config EOF`
# type: string
# required: false
# default: ""
modifiers: |
  - body.MultiFetcher:
      resources:
        - body.JSONResource:
            method: GET
            url: https://jsonplaceholder.typicode.com/users/1
            behavior: replace # replaces upstream http response
            modifier:
              status.Verifier:
                statusCode: 200 # verify that the resource returns 200 status code
        - body.JSONResource:
            method: GET
            url: https://jsonplaceholder.typicode.com/users/1/todos
            behavior: merge # merge with the previous resource
            group: todos # group this response into "todos" key
            modifier:
              status.Verifier:
                statusCode: 200 # verify that the resource returns 200 status code
  - body.JSONPatch:
      scope: [response]
      patch:
        - {op: move, from: /todos, path: /Todos}
```

### Modifiers

This reference is adapted from [Martian's wiki](https://github.com/google/martian/wiki/Modifier-Reference)

Modifiers are able to mutate a request, a response or both.

#### JSONResource

The `body.JSONResource` fetches a remote JSON resource and merges/replaces the upstream response body with the response of the remote request depending on the `behavior` option. Defaults to the `merge` behavior.

The merging is done using [RFC7386: JSON Merge Patch](https://tools.ietf.org/html/rfc7386)

```yaml
body.JSONResource:
  scope: [response]
  method: GET
  url: https://jsonplaceholder.typicode.com/users/1
  behavior: replace # replaces upstream http response
  modifier:
    status.Verifier:
      statusCode: 200
```

#### JSONPatch

The `body.JSONPatch` patches the JSON request or response body using [RFC6902: JSON Patch](https://tools.ietf.org/html/rfc6902)

```yaml
body.JSONPatch:
  scope: [response]
  patch:
    - { op: move, from: /todos, path: /Todos }
```

#### Skip

The `skip.RoundTrip` skips the HTTP roundtrip to the upstream URL that was specified via the `--url` flag

```yaml
skip.RoundTrip:
  scope: [request]
```

#### Cookie

The `cookie.Modifier` injects a cookie into a request or a response.

Example configuration that injects a `Martian-Cookie` cookie into responses:

```yaml
cookie.Modifier:
  scope: [response]
  name: Martian-Cookie
  value: some value
  path: "/some/path"
  domain: example.com
  expires: "2025-04-12T23:20:50.52Z" # RFC 3339
  secure: true
  httpOnly: false
  maxAge: 86400
```

#### Header

The `header.Modifier` injects or modifies headers in a request or a response.

Example configuration that injects an `X-Martian` header with the value of
`true` into requests:

```yaml
header.Modifier:
  scope: [request]
  name: X-Martian
  value: "true"
```

#### Header Blacklist

The `header.Blacklist` deletes headers from a request or a response.

Example configuration that deletes response headers with the names
`X-Martian` and `Y-Martian`:

```yaml
header.Blacklist:
  scope: [response]
  names: [X-Martian, Y-Martian]
```

#### Query String

The `querystring.Modifier` adds or modifies query string parameters on a
request. Any existing parameter values are replaced.

Example configuration that sets the query string parameter `foo` to the value
of `bar` on requests and responses:

```yaml
querystring.Modifier:
  scope: [request, response]
  name: foo
  value: bar
```

#### Status

The `status.Modifier` modifies the HTTP status code on a response.

Example configuration that sets the HTTP status of responses to `200`:

```yaml
status.Modifier:
  scope: [response]
  statusCode: 200
```

#### URL

The `url.Modifier` modifies the URL on a request.

Example configuration that redirects requests to `https://www.google.com/proxy?testing=true`

```yaml
url.Modifier:
  scope: [request]
  scheme: https
  host: www.google.com
  path: "/proxy"
  query: testing=true
```

#### Message Body

The `body.Modifier` modifies the body of a request or response. Additionally, it will modify the following headers to ensure proper transport: `Content-Type`, `Content-Length`, `Content-Encoding`. The body is expected to be uncompressed and Base64 encoded.

```yaml
body.Modifier:
  scope: [request, response]
  contentType: text/plain; charset=utf-8
  body: TWFydGlhbiBpcyBhd2Vzb21lIQ==
```

### Groups

Groups hold lists of modifiers (or filters, or groups) that are executed in a particular order.

#### MultiFetcher

A `body.MultiFetcher` holds a list of data fetchers that are fetched concurrently, the response are modified in first-in, first-out order.

Currently only works with the `body.JSONResource` modifier

```yaml
body.MultiFetcher:
  resources:
    - body.JSONResource:
        method: GET
        url: https://jsonplaceholder.typicode.com/users/1
        behavior: replace # replaces upstream http response
        modifier:
          status.Verifier:
            statusCode: 200
    - body.JSONResource:
        method: GET
        url: https://jsonplaceholder.typicode.com/users/1/todos
        behavior: merge # merge with the first call
        group: todos # group this response into "todos" key
        modifier:
          status.Verifier:
            statusCode: 500
```

#### FIFO

A `fifo.Group` holds a list of modifiers that are executed in first-in,
first-out order.

Example configuration that adds the query string parameter of `foo=bar` on the
request and deletes any `X-Martian` headers on responses:

```yaml
fifo.Group:
  scope: [request, response]
  modifiers:
    - querystring.Modifier:
        scope: [request]
        name: foo
        value: bar
    - header.Blacklist:
        scope: [response]
        names: [X-Martian]
```

#### Priority

A `priority.Group` holds a list of modifiers that are each associated with an
integer. Each integer represents the "priority" of the associated modifier, and
the modifiers are run in order of priority (from highest to lowest).

In the case that two modifiers have the same priority, order of execution of
those modifiers will be determined by the order in which the modifiers with the
same priority were added: the newest modifier added will run first.

Example configuration that adds the query string parameter of `foo=bar` and
deletes any `X-Martian` headers on requests:

```yaml
priority.Group:
  scope: [response]
  modifiers:
    - priority: 0 # will run last
      modifier:
        querystring.Modifier:
          scope: [response]
          name: foo
          value: bar
   - priority: 100 # will run first
      modifier:
        header.Blacklist:
          scope: [response]
          names: [X-Martian]
```

### Filters

Filters execute contained modifiers if the defined conditional is met.

#### Header

The `header.Filter` executes its contained modifier if the a request or
response contains a header that matches the defined `name` and `value`. In the
case that the `value` is undefined, the contained modifier executes if a
request or response contains a header with the defined name.

Example configuration that add the query string parameter `foo=bar` on
responses if the response contains a `Martian-Testing` header with the value of
`true`:

```yaml
header.Filter:
  scope: [response]
  name: Martian-Testing
  value: "true"
  modifier:
    querystring.Modifier:
      scope: [response]
      name: foo
      value: bar
```

#### Query String

The `querystring.Filter` executes its contained modifier if the request or
response contains a query string parameter matches the defined `name` and
`value` in the filter. The `name` and `value` in the filter are regular
expressions in the [RE2](https://github.com/google/re2/wiki/Syntax) syntax. In
the case that a `value` is not defined, the contained modifier is executed if
the query string parameter `name` matches.

Example of a configuration that sets the `Mod-Run` header to `true` on requests
with the query string parameter to `param=true`:

```yaml
querystring.Filter:
  scope: [request]
  name: param
  value: "true"
  modifier:
    header.Modifier:
      scope: [request]
      name: Mod-Run
      value: "true"
```

#### URL

The `url.Filter` executes its contained modifier if the request URL matches all
of the provided parts. Missing parameters are ignored.

Example configuration that sets the `Mod-Run` header to true on all requests
that are made to a URL with the scheme `https`.

```yaml
url.Filter:
  scope: [request]
  scheme: https
  modifier:
    header.Modifier:
      scope: [request]
      name: Mod-Run
      value: "true"
```

### Verifiers

Verifier check network traffic against defined expectations. Failed
verifications are returned as a list of errors.

#### Header

The `header.Verifier` records an error for every request or response that does not contain a header that matches the `name` and `value`. In the case that a `value` is not provided, the an error is recorded for every failed match of the `name`.

Example configuration that records an error if the `Martian-Test` header is not set to `true` on requests and responses:

```yaml
header.Verifier:
  scope: [request, response]
  name: Martian-Test
  value: "true"
```

#### Method

The `method.Verifier` records an error for every request that does not match the expected HTTP method.

Example configuration that records an error for every request that is not a `POST`:

```yaml
method.Verifier:
  scope: [request]
  method: POST
```

#### Pingback

The `pingback.Verifier` records an error for every request that fails to generate a pingback request with the provided url parameters. In the case that certain parameters are not provided, those portions of the URL are not used for matching.

Example configuration that records an error for every request that does not result in a pingback to `https://example.com/testing?test=true`:

```yaml
pingback.Verifier:
  scope: [request]
  scheme: https
  host: example.com
  path: "/testing"
  query: test=true
```

#### Query String

The `querystring.Verifier` records an error for every request or response that does not contain a query string parameter that matches the `name` and `value`. In the case that a `value` is not provided, then an error is recorded of every failed match of `name`, ignoring any `value`.

Example configuration that records an error for each request that does not contain a query string parameter of `param=true`:

```yaml
querystring.Verifier:
  scope: [request]
  name: param
  value: "true"
```

#### Status

The `status.Verifier` records an error for every response that is returned with a HTTP status that does not match the `statusCode` provided.

Example configuration that records an error for each response that does not have the HTTP status of `200 OK`:

```yaml
status.Verifier:
  scope: [response]
  statusCode: 200
```

#### URL

The `url.Verifier` records an error for every request URL that does not match all provided parts of a URL.

Example configuration that records an error for each request that is not for `https://www.martian.proxy/testing?test=true`:

```yaml
url.Verifier:
  scope: [request]
  scheme: https
  host: www.martian.proxy
  path: "/testing"
  query: test=true
```
