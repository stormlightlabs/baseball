package core

import (
	"context"
	"testing"
)

func TestCountModeDefaultsToLightweight(t *testing.T) {
	mode := CountModeFromContext(context.Background())
	if mode != CountModeLightweight {
		t.Fatalf("expected default count mode %q, got %q", CountModeLightweight, mode)
	}

	effective := EffectiveCountModeFromContext(context.Background())
	if effective != CountModeLightweight {
		t.Fatalf("expected default effective mode %q, got %q", CountModeLightweight, effective)
	}
}

func TestCountModeStrictFallbackTracking(t *testing.T) {
	ctx := WithCountMode(context.Background(), CountModeStrict)
	if got := CountModeFromContext(ctx); got != CountModeStrict {
		t.Fatalf("expected requested mode %q, got %q", CountModeStrict, got)
	}
	if got := EffectiveCountModeFromContext(ctx); got != CountModeStrict {
		t.Fatalf("expected effective mode %q before fallback, got %q", CountModeStrict, got)
	}

	MarkCountModeFallback(ctx)
	if got := EffectiveCountModeFromContext(ctx); got != CountModeFallback {
		t.Fatalf("expected effective mode %q after fallback, got %q", CountModeFallback, got)
	}
}

func TestCountModeLightweightIgnoresFallbackMark(t *testing.T) {
	ctx := WithCountMode(context.Background(), CountModeLightweight)
	MarkCountModeFallback(ctx)
	if got := EffectiveCountModeFromContext(ctx); got != CountModeLightweight {
		t.Fatalf("expected effective mode to remain %q, got %q", CountModeLightweight, got)
	}
}
