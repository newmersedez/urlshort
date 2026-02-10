package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type queryTracer struct {
	logger *slog.Logger
}

func newQueryTracer(logger *slog.Logger) *queryTracer {
	return &queryTracer{
		logger: logger,
	}
}

func (t *queryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	t.logger.Debug("running query", "sql", data.SQL)
	return ctx
}

func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.logger.Debug("query finished", "result", data.CommandTag)
}
