package reset_zerolog_timestamps

import (
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Unix(0, 0)
	}
}
