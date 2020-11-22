You may think that BFF stands for "Best Friends Forever", although that's very apt but that's not what this tool is built for.

BFF was built out of the requirement to have an API-aware proxy that is cabaple of routing, filtering, verifying and modifing HTTP request and response.

It is built on top of [Google's Martian](https://github.com/google/martian) proxy development tool-kit, therefore `bff` supports the same [built-in modifiers](https://github.com/google/martian/wiki/Modifier-Reference).

It can be composed to fit most use-cases and additional modifiers can be introduced and composed together to provide a customized functionality.

In other words, BFF stands for ["Backend For Frontend"](https://samnewman.io/patterns/architectural/bff/).

## Flags

```
  -c, --config string      config file (default is $XDG_CONFIG_HOME/bff/config.yaml)
  -h, --help               help for bff
  -i, --insecure           Skip TLS verify
  -m, --modifiers string   Modifiers
  -p, --port string        Port to run the server on (default "5000")
  -u, --url string         Proxy url
  -v, --verbosity int      Verbosity
```

## Quick Start

Download the latest release for your platform and architecture from [here](/-/releases)

```sh
# make the temp directory
mkdir -p /tmp/quickstart-bff
cd /tmp/quickstart-bff

# download the bff executable into it
curl -sfL https://github.com/imranismail/bff/releases/download/v0.1.2/bff_0.1.2_Linux_x86_64.tar.gz | tar xvz

# create the config file
cat > config.yml <<YAML
modifiers: |-
  # skip upstream roundtrip
  - skip.RoundTrip:
      scope: [request]
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

# start the proxy
./bff -v=3 -u http://jsonplaceholder.typicode.com

# test it
curl -v http://localhost:5000
```
