package tracer

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vnafikov/foundation/logger"
	"github.com/vnafikov/foundation/logger/parameters"
)

const (
	loggerAttrKeyArguments = "arguments"
	loggerAttrKeyRows      = "rows"
	loggerAttrKeyPID       = "pid"
	loggerAttrKeyDuration  = "duration"
)

var (
	ctxKeyQuery    = contextKeyQuery{}
	ctxKeyBatch    = contextKeyBatch{}
	ctxKeyCopyFrom = contextKeyCopyFrom{}
)

type (
	contextKeyQuery    struct{}
	contextKeyBatch    struct{}
	contextKeyCopyFrom struct{}
)

type Tracer struct{}

func (t *Tracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if t.shouldSkipLog() {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyQuery, traceQueryData{
		sql:       data.SQL,
		args:      data.Args,
		startTime: time.Now(),
	})
}

func (*Tracer) shouldSkipLog() bool {
	return !logger.ShouldLogDebug()
}

func (t *Tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if t.shouldSkipLog() {
		return
	}

	pid := uint64(conn.PgConn().PID())
	queryData := t.queryData(ctx, pid)
	msg := parameters.PGSQLMessagePrefix + " " + queryData.sql
	attrs := t.argAttrs(queryData.args)
	duration := time.Since(queryData.startTime)
	logger.Debug(
		ctx,
		msg,
		slog.Group(loggerAttrKeyArguments, attrs...),
		slog.Int64(loggerAttrKeyRows, data.CommandTag.RowsAffected()),
		slog.Uint64(loggerAttrKeyPID, pid),
		slog.Duration(loggerAttrKeyDuration, duration),
	)
}

func (*Tracer) queryData(ctx context.Context, pid uint64) traceQueryData {
	value := ctx.Value(ctxKeyQuery)
	if value == nil {
		logger.Error(ctx, errContextValueNotFound, "context.(Context).Value", slog.Uint64(loggerAttrKeyPID, pid))

		return traceQueryData{}
	}

	queryData, ok := value.(traceQueryData)
	if !ok {
		logger.Error(
			ctx,
			errContextValueTypeAssertionFailed,
			"context.(Context).Value.(traceQueryData)",
			slog.Uint64(loggerAttrKeyPID, pid),
		)
	}
	return queryData
}

func (t *Tracer) argAttrs(args []any) []any {
	attrs := make([]any, len(args))
	for i, arg := range args {
		attrs[i] = slog.String("$"+strconv.Itoa(i+1), t.attrValue(arg))
	}
	return attrs
}

func (*Tracer) attrValue(arg any) string {
	switch arg := arg.(type) {
	case string:
		return arg
	case []byte:
		return string(arg)
	}
	return fmt.Sprintf("%+v", arg)
}

func (t *Tracer) TraceBatchStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceBatchStartData) context.Context {
	if t.shouldSkipLog() {
		return ctx
	}

	now := time.Now()
	return context.WithValue(ctx, ctxKeyBatch, &traceBatchData{
		batchStartTime: now,
		queryStartTime: now,
	})
}

func (t *Tracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	if t.shouldSkipLog() {
		return
	}

	pid := uint64(conn.PgConn().PID())
	batchData := t.batchData(ctx, pid)
	defer batchData.setQueryStartTime()

	msg := parameters.PGSQLMessagePrefix + " " + data.SQL
	attrs := t.argAttrs(data.Args)
	duration := time.Since(batchData.queryStartTime)
	logger.Debug(
		ctx,
		msg,
		slog.Group(loggerAttrKeyArguments, attrs...),
		slog.Int64(loggerAttrKeyRows, data.CommandTag.RowsAffected()),
		slog.Uint64(loggerAttrKeyPID, pid),
		slog.Duration(loggerAttrKeyDuration, duration),
	)
}

func (*Tracer) batchData(ctx context.Context, pid uint64) *traceBatchData {
	value := ctx.Value(ctxKeyBatch)
	if value == nil {
		logger.Error(ctx, errContextValueNotFound, "context.(Context).Value", slog.Uint64(loggerAttrKeyPID, pid))

		return new(traceBatchData)
	}

	batchData, ok := value.(*traceBatchData)
	if !ok {
		logger.Error(
			ctx,
			errContextValueTypeAssertionFailed,
			"context.(Context).Value.(*traceBatchData)",
			slog.Uint64(loggerAttrKeyPID, pid),
		)

		return new(traceBatchData)
	}
	return batchData
}

func (t *Tracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, _ pgx.TraceBatchEndData) {
	if t.shouldSkipLog() {
		return
	}

	pid := uint64(conn.PgConn().PID())
	batchData := t.batchData(ctx, pid)
	duration := time.Since(batchData.batchStartTime)
	logger.Debug(
		ctx,
		"PG batch execution ended.",
		slog.Uint64(loggerAttrKeyPID, pid),
		slog.Duration(loggerAttrKeyDuration, duration),
	)
}

func (t *Tracer) TraceCopyFromStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	if t.shouldSkipLog() {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyCopyFrom, traceCopyFromData{
		tableName:   data.TableName,
		columnNames: data.ColumnNames,
		startTime:   time.Now(),
	})
}

func (t *Tracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	if t.shouldSkipLog() {
		return
	}

	pid := uint64(conn.PgConn().PID())
	copyFromData := t.copyFromData(ctx, pid)
	columns := strings.Join(copyFromData.columnNames, ", ")
	duration := time.Since(copyFromData.startTime)
	logger.Debug(
		ctx,
		"PG COPY FROM execution ended.",
		slog.String("table", copyFromData.tableName.Sanitize()),
		slog.String("columns", columns),
		slog.Int64(loggerAttrKeyRows, data.CommandTag.RowsAffected()),
		slog.Uint64(loggerAttrKeyPID, pid),
		slog.Duration(loggerAttrKeyDuration, duration),
	)
}

func (*Tracer) copyFromData(ctx context.Context, pid uint64) traceCopyFromData {
	value := ctx.Value(ctxKeyCopyFrom)
	if value == nil {
		logger.Error(ctx, errContextValueNotFound, "context.(Context).Value", slog.Uint64(loggerAttrKeyPID, pid))

		return traceCopyFromData{}
	}

	copyFromData, ok := value.(traceCopyFromData)
	if !ok {
		logger.Error(
			ctx,
			errContextValueTypeAssertionFailed,
			"context.(Context).Value.(traceCopyFromData)",
			slog.Uint64(loggerAttrKeyPID, pid),
		)
	}
	return copyFromData
}
