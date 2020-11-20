You may think that BFF stands for "Best Friends Forever", although that's very apt but that's not what this tool is built for.

BFF was built out of the requirement to have an API-aware proxy that is cabaple of routing, filtering, verifying and modifing HTTP request and response.

It is built on top of [Google's Martian](https://github.com/google/martian) proxy development kit and written in Golang.

It can be composed to fit most use-cases and additional modifiers can be introduced and composed together to provide a customized functionality.

In other words, BFF stands for ["Backend For Frontend"](https://samnewman.io/patterns/architectural/bff/).

## Quick Start

Download the latest release for your platform and architecture from [here](/-/releases)

```sh
mkdir -p /tmp/quickstart-bff
tar -xvzf bff_<version>_<platform>_<arch>.tar.gz /tmp/quickstart-bff
cd /tmp/quickstart-bff
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
./bff -v=3 -u http://jsonplaceholder.typicode.com
curl -v http://localhost:5000
```
