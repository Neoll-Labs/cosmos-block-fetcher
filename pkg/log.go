package pkg

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Info().Msg("hello, cosmos-block-fetcher")

	// Set default level to info
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
