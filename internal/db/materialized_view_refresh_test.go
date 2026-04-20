package db

import (
	"errors"
	"testing"
)

func TestShouldRetryNonConcurrent(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "concurrently-not-populated",
			err:  errors.New(`ERROR: CONCURRENTLY cannot be used when the materialized view is not populated (SQLSTATE 0A000)`),
			want: true,
		},
		{
			name: "cannot-refresh-concurrently",
			err:  errors.New(`ERROR: cannot refresh materialized view "public.triple_plays" concurrently (SQLSTATE 55000)`),
			want: true,
		},
		{
			name: "other-error",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := shouldRetryNonConcurrent(tc.err)
			if got != tc.want {
				t.Fatalf("shouldRetryNonConcurrent(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsMaterializedViewDeferredDependencyError(t *testing.T) {
	if !isMaterializedViewDeferredDependencyError(errors.New(`ERROR: materialized view "x" has not been populated`)) {
		t.Fatal("expected deferred dependency error to be detected")
	}
	if isMaterializedViewDeferredDependencyError(errors.New("other")) {
		t.Fatal("did not expect unrelated error to be treated as deferred dependency")
	}
}
