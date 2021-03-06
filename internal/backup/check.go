package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/mbrt/backsched/internal/config"
)

// ComputeOutdated returns the list of outdated backups.
func ComputeOutdated(ctx context.Context, cfg config.Config, env Env) ([]Info, error) {
	var res []Info
	state := loadState(env.Sio)

	for _, bc := range cfg.Backups {
		clog := log.With().Str("backup", bc.Name).Logger()
		t, ok := state.LastBackupOf(bc.Name)
		since := env.Clock.Now().Sub(t)
		if ok && since < time.Duration(bc.Interval) {
			clog.Info().Msgf("Skipping because: last backup was %s ago", fmtDuration(since))
			continue
		}
		if !ok {
			// The backup was never executed, use the special value to signal it.
			since = 0
		}
		res = append(res, Info{
			Since:  since,
			Backup: bc,
		})
	}

	return res, nil
}

// Info represents the state of a backup.
type Info struct {
	// Since contains how long ago the backup was performed.
	// Zero is a special value meaning "never".
	Since  time.Duration
	Backup config.Backup
}

func (i Info) String() string {
	if i.Since == 0 {
		return fmt.Sprintf("%s: last backup was never?", i.Backup.Name)
	}
	return fmt.Sprintf("%s: last backup was %s ago", i.Backup.Name, fmtDuration(i.Since))
}

func fmtDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	if d < time.Hour*48 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	return fmt.Sprintf("%d days", int(d/time.Hour/24))
}
