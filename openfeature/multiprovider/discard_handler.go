package multiprovider

import (
	"context"
	"log/slog"
)

// TODO: Remove once we upgrade to go 1.24+

// discardHandler This type is only required until we upgrade to go 1.24 where slog.DiscardHandler is available for use
type discardHandler struct{}

func (d discardHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (d discardHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (d discardHandler) WithAttrs([]slog.Attr) slog.Handler {
	return d
}

func (d discardHandler) WithGroup(string) slog.Handler {
	return d
}

var (
	_              slog.Handler = (*discardHandler)(nil)
	DiscardHandler slog.Handler = discardHandler{}
)
