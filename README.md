# Sensu Squadcast Handler

[![Bonsai Asset Badge](https://img.shields.io/badge/Sensu%20Squadcast%20Handler-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/SquadcastHub/sensu-squadcast-handler)
![Go Test](https://github.com/SquadcastHub/sensu-squadcast-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/SquadcastHub/sensu-squadcast-handler/workflows/goreleaser/badge.svg)

The Sensu Go Squadcast handler is a [Sensu Event Handler][1] that sends event data to
a Squadcast  endpoint.

## Installation

Create an executable script from this source.

From the local path of the sensu-squadcast-handler repository:
```
go build -o /usr/local/bin/sensu-squadcast-handler
```

## Configuration

Example Sensu Go handler definition:


squadcast-handler.yaml

```yaml
type: Handler
api_version: core/v2
metadata:
  name: squadcast
  namespace: default
spec:
  command: sensu-squadcast-handler
  env_vars:
  - SENSU_SQUADCAST_APIURL= <Squadcast Alert Source Url>
  filters:
  - is_incident
  timeout: 10
  type: pipe
```

`sensuctl create -f squadcast-handler.yaml`

Example Sensu Go check definition:

```yaml
api_version: core/v2
type: CheckConfig
metadata:
  namespace: default
  name: health-check
spec:
  command: check-http -u http://localhost:8080/health
  subscriptions:
  - test
  publish: true
  interval: 10
  handlers:
  - squadcast

```

## Usage examples

Help:

```
The Sensu Go Squadcast handler sends Sensu events to Squadcast


Usage:
  sensu-squadcast-handler [flags]


Flags:
  -a, --api-url string   The URL for the Squadcast API
  -h, --help             help for sensu-squadcast-handler
```

[1]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work