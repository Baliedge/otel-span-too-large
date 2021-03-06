package main

// Trace a code block as a scoped span.
// * Use instead of manual instrumentation: `tracer.Start()`/`span.End()`.
// * Must call `InitTracing()` first.
// * Automates start/end of span.
// * Tags file and line number where span started.
// * If function returned error:
//   * Span is tagged as error.
//   * Sets span attributes `otel.status_code` and `otel.status_description`
//     with error details.
//   * Logs error details to span.

import (
	"context"
	"runtime"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ScopeAction func(ctx context.Context) error

// Start a scope with span named after fully qualified caller function.
func StartScope(ctx context.Context, opts ...trace.SpanStartOption) context.Context {
	spanName, fileTag := getCallerSpanName(2)
	return startSpan(ctx, spanName, fileTag, opts...)
}

// Start a scope with user-provided span name.
func StartNamedScope(ctx context.Context, spanName string, opts ...trace.SpanStartOption) context.Context {
	fileTag := getFileTag(2)
	return startSpan(ctx, spanName, fileTag, opts...)
}

// End scope created by `StartScope()`/`StartNamedScope()`.
// Logs error return value and ends span.
func EndScope(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)

	// If scope returns an error, mark span with error.
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}

// Scope calls action function within a tracing span named after the calling
// function.
// Equivalent to wrapping a code block with `StartScope()`/`EndScope()`.
// Must call `InitTracing()` first.
func Scope(ctx context.Context, action ScopeAction, opts ...trace.SpanStartOption) error {
	spanName, fileTag := getCallerSpanName(2)
	ctx = startSpan(ctx, spanName, fileTag, opts...)
	err := action(ctx)
	EndScope(ctx, err)
	return err
}

// NamedScope calls action function within a tracing span.
// Equivalent to wrapping a code block with `StartNamedScope()`/`EndScope()`.
// Must call `InitTracing()` first.
func NamedScope(ctx context.Context, spanName string, action ScopeAction, opts ...trace.SpanStartOption) error {
	fileTag := getFileTag(2)
	ctx = startSpan(ctx, spanName, fileTag, opts...)
	err := action(ctx)
	EndScope(ctx, err)
	return err
}

func startSpan(ctx context.Context, spanName, fileTag string, opts ...trace.SpanStartOption) context.Context {
	// Initialize span.
	tracer, ok := ctx.Value(tracerKey{}).(trace.Tracer)
	if !ok {
		// No tracer embedded.  Fall back to default tracer.
		tracer = GetDefaultTracer()
	}
	// Else, omit tracing.
	if tracer == nil {
		return ctx
	}

	opts = append(opts, trace.WithAttributes(
		attribute.String("file", fileTag),
	))

	ctx, _ = tracer.Start(ctx, spanName, opts...)
	return ctx
}

func getCallerSpanName(skip int) (string, string) {
	pc, file, line, ok := runtime.Caller(skip)

	// Determine source file and line number.
	var fileTag, spanName string
	if ok {
		fileTag = file + ":" + strconv.Itoa(line)
		spanName = runtime.FuncForPC(pc).Name()
	} else {
		// Rare condition.  Probably a bug in caller.
		fileTag = "unknown"
	}

	return spanName, fileTag
}

func getFileTag(skip int) string {
	_, file, line, ok := runtime.Caller(1)

	// Determine source file and line number.
	if !ok {
		// Rare condition.  Probably a bug in caller.
		return "unknown"
	}

	return file + ":" + strconv.Itoa(line)
}
