package unter_test

import (
	"fmt"
	"math/rand"
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

var (
	rides []unter.Ride
	// Numbers from from statistics of real data Jan-June 2020
	nDrivers = 1000
	nRides   = 100 // per driver
)

func init() {
	for d := 0; d < nDrivers; d++ {
		driver := fmt.Sprintf("drv-%d", d)
		for i := 0; i < nRides; i++ {
			start := time.Now()
			r := unter.Ride{
				Driver:   driver,
				Start:    start,
				End:      start.Add(time.Duration(rand.Intn(100)) * time.Minute),
				Distance: rand.Float64() * 30,
			}
			rides = append(rides, r)
		}
	}
	rand.Shuffle(len(rides), func(i, j int) { rides[i], rides[j] = rides[j], rides[i] })
}

func BenchmarkByDriver(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rs := unter.ByDriver(rides)
		if len(rs) != nDrivers {
			b.Fatal(rs)
		}
	}
}
