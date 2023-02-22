package reset_zerolog_timestamps

import (
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Unix(0, 0)
	}

	// Blank out line numbers, as otherwise our test expectations change whenever we edit production code.
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return file + ":XXX"
	}
}
