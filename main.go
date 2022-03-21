package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	launch(context.Background())
	CloseTracing(context.Background())
}

func launch(ctx context.Context) {
	ctx, err := initTracing(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Error in initTracing")
	}

	ctx = StartScope(ctx)
	defer func() {
		EndScope(ctx, err)
	}()

	logrus.WithContext(ctx).Info("Starting...")

	makeTrace(ctx)
}

// Simulate span that gets rejected with "span too large to send" when MaxPacketSize = 1472.
func makeTrace(ctx context.Context) {
	ctx = StartNamedScope(ctx, "github.com/mailgun/turret/v2/client/golang.(*transaction).Close")
	err := errors.New("code:250  message:\"OK\"  utf8_enabled:true  mx_host:\"10.5.0.2\"  secure:true  smtp_log:\"19:15:17.475      0s \u003e- {19:15:17.468, #0, 0}\\n19:15:17.476      0s ** age=8.9931005s, sessionCount=9\\n19:15:17.476      0s \u003c- {#0}\\n19:15:17.476      0s -\u003e MAIL FROM:\u003csender@example.com\u003e BODY=8BITMIME SMTPUTF8\\n19:15:17.477     1ms -\u003c 250 Sender address accepted\\n19:15:17.4771ms -\u003e RCPT TO:\u003crecipient@example.com\u003e\\n19:15:17.478     2ms -\u003c 250 Recipient address accepted\\n19:15:17.478     2ms -\u003e DATA\\n19:15:17.479     3ms -\u003c 354 Continue\\n19:15:17.480     4ms \u003e- {19:15:17.472, #1, 18}\\n19:15:17.480     4ms \u003e- {19:15:17.472, #2, 0, last}\\n19:15:17.481     5ms \u003c- {#1}\\n19:15:17.483     7ms -\u003c 250 Great success\\n19:15:17.484     8ms \u003c- {#2, last}\\n\"  mx_host_ip:\"10.5.0.2\"  tls_version:772  tls_cipher_suite:4865")
	EndScope(ctx, err)
}

func initTracing(ctx context.Context) (context.Context, error) {
	const serviceName = "github.com/mailgun/turret/v2"
	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
	}).Info("Initializing tracing...")

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return ctx, errors.Wrap(err, "error in resource.Merge")
	}

	ctx, _, err = InitTracing(ctx,
		"otel-localhost",
		sdktrace.WithResource(res),
	)
	if err != nil {
		return ctx, errors.Wrap(err, "error in InitTracing")
	}

	return ctx, nil
}
