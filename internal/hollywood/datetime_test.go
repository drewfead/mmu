package hollywood_test

import (
	"testing"
	"time"

	"github.com/drewfead/mmu/internal/hollywood"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Unit_ParseDateTime(t *testing.T) {
	laTZ, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	tests := []struct {
		name        string
		monthAndDay string
		clockTime   string
		tz          *time.Location
		expectTime  time.Time
		expectBool  bool
	}{
		{
			name:        "simple",
			monthAndDay: "Tuesday, June 20",
			clockTime:   "7:30 PM",
			tz:          laTZ,
			expectTime:  time.Date(0, 6, 20, 19, 30, 0, 0, laTZ),
			expectBool:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualTime, actualBool := hollywood.ParseDateTime(tt.monthAndDay, tt.clockTime, tt.tz)

			assert.Equal(t, tt.expectTime, actualTime)
			assert.Equal(t, tt.expectBool, actualBool)
		})
	}
}
