package seed

import "testing"

func TestFormatByteSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{
			name: "bytes",
			size: 999,
			want: "999 B",
		},
		{
			name: "one kilobyte",
			size: 1024,
			want: "1.0 KB",
		},
		{
			name: "fractional kilobytes",
			size: 1536,
			want: "1.5 KB",
		},
		{
			name: "one megabyte",
			size: 1024 * 1024,
			want: "1.0 MB",
		},
		{
			name: "one gigabyte",
			size: 1024 * 1024 * 1024,
			want: "1.0 GB",
		},
		{
			name: "one terabyte",
			size: 1024 * 1024 * 1024 * 1024,
			want: "1.0 TB",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := formatByteSize(tt.size)
			if got != tt.want {
				t.Fatalf("formatByteSize(%d): want %q, got %q", tt.size, tt.want, got)
			}
		})
	}
}
