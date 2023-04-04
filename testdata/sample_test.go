package testdata

import (
	"testing"
)

// Test case for function sum
package testdata_test

import (
	"testing"

	"github.com/path/to/testdata"
)

func TestSum(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 3},
		{10, -5, 5},
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := testdata.Sum(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Sum(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

