module example/demo

go 1.22.0

replace github.com/alibaba/loongsuite-go-agent => ../../

replace github.com/alibaba/loongsuite-go-agent/test/verifier => ../../test/verifier

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/gomodule/redigo v1.9.2
	github.com/juju/errors v1.0.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0
)
