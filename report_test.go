package unter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/353solutions/unter"
)

// FIXME: Must be in sync with report.go
const (
	minFee  = 250 // ¢
	perMile = 250
	perHour = 3000
)

var feeCases = []struct {
	duration time.Duration
	distance float64
	shared   bool
	expected int
}{
	{time.Second, 0.1, false, minFee},
	{3 * time.Minute, 3, false, 750},
	{7 * time.Hour, 3, false, 7 * perHour},
	{3 * time.Minute, 3, true, 675},
}

// Homework: Run a fuzz test on RideFee
// https://go.dev/doc/tutorial/fuzz#fuzz_test

func TestRideFee(t *testing.T) {
	for _, tc := range feeCases {
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			fee := unter.RideFee(tc.duration, tc.distance, tc.shared)
			require.Equal(t, tc.expected, fee)
		})
	}
}
