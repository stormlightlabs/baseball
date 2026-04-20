package core

import "context"

// CountMode controls how expensive row count operations should be.
type CountMode string

const (
	CountModeLightweight CountMode = "lightweight"
	CountModeStrict      CountMode = "strict"
	CountModeFallback    CountMode = "fallback"
)

type countModeContextKey struct{}
type countModeContextState struct {
	requested CountMode
	fallback  bool
}

// WithCountMode adds a count mode hint to context.
func WithCountMode(ctx context.Context, mode CountMode) context.Context {
	state := &countModeContextState{
		requested: normalizeCountMode(mode),
	}
	return context.WithValue(ctx, countModeContextKey{}, state)
}

// CountModeFromContext returns the requested count mode; defaults to lightweight.
func CountModeFromContext(ctx context.Context) CountMode {
	if state := countModeStateFromContext(ctx); state != nil {
		return state.requested
	}
	return CountModeLightweight
}

// MarkCountModeFallback marks the current request as having used fallback count behavior.
func MarkCountModeFallback(ctx context.Context) {
	state := countModeStateFromContext(ctx)
	if state == nil {
		return
	}
	if state.requested != CountModeStrict {
		return
	}
	state.fallback = true
}

// EffectiveCountModeFromContext returns the effective mode used during request execution.
func EffectiveCountModeFromContext(ctx context.Context) CountMode {
	state := countModeStateFromContext(ctx)
	if state == nil {
		return CountModeLightweight
	}
	if state.requested == CountModeStrict && state.fallback {
		return CountModeFallback
	}
	return state.requested
}

func countModeStateFromContext(ctx context.Context) *countModeContextState {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(countModeContextKey{}); v != nil {
		if state, ok := v.(*countModeContextState); ok {
			return state
		}
		if mode, ok := v.(CountMode); ok {
			return &countModeContextState{requested: normalizeCountMode(mode)}
		}
	}
	return nil
}

func normalizeCountMode(mode CountMode) CountMode {
	if mode == CountModeStrict {
		return CountModeStrict
	}
	return CountModeLightweight
}
