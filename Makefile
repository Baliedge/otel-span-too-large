JAEGER_AGENT_HOST=javelin.ethn.home

build:
	go build -v .

run:
	OTEL_EXPORTER_JAEGER_AGENT_HOST=${JAEGER_AGENT_HOST} OTEL_TRACES_SAMPLER=always_on go run -v .
