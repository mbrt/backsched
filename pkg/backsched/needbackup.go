package backsched

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// NeedBackup prints to console the outdated backups.
func NeedBackup(config Config, state State, wmNotify bool) error {
	outdated := ComputeOutdated(config, state)
	if wmNotify {
		return reportOutdated(outdated, notifySend{osExecutor{}, "backsched"})
	}
	return reportOutdated(outdated, nil)
}

// ComputeOutdated returns the outdated backups info.
func ComputeOutdated(config Config, state State) []OutdatedBackup {
	res := []OutdatedBackup{}
	now := time.Now()
	for _, backup := range config.Backups {
		lastBackup, ok := state.LastBackupOf(backup.Name)
		if !ok {
			res = append(res, OutdatedBackup{backup.Name, 0, true})
		} else {
			sinceDays := int(now.Sub(lastBackup).Hours() / 24)
			if sinceDays >= backup.EveryDays {
				res = append(res, OutdatedBackup{backup.Name, sinceDays, false})
			}
		}
	}
	return res
}

// OutdatedBackup contains info about an out-to-date backup.
type OutdatedBackup struct {
	Name      string
	SinceDays int
	Never     bool
}

// Notifier notifies the user with a message
type Notifier interface {
	Notify(summary, msg string) error
}

type notifySend struct {
	exec  Executor
	title string
}

func (n notifySend) Notify(summary, msg string) error {
	args := []string{"--urgency=normal", fmt.Sprintf("--app-name=%s", n.title), summary, msg}
	if err := n.exec.Exec("notify-send", args...); err != nil {
		return errors.Wrap(err, "failed notification")
	}
	return nil
}

func (b OutdatedBackup) String() string {
	if b.Never {
		return fmt.Sprintf("%s: never done", b.Name)
	}
	return fmt.Sprintf("%s: last backup was %d days ago", b.Name, b.SinceDays)
}

func reportOutdated(outdated []OutdatedBackup, notifier Notifier) error {
	if len(outdated) == 0 {
		fmt.Println("Backups are up to date")
		// don't annoy the user if everything is fine
	} else {
		summary := "The following backups are outdated:"
		msgs := make([]string, len(outdated)+1)
		for i, b := range outdated {
			msgs[i] = fmt.Sprintf("  - %v", b)
		}
		msg := strings.Join(msgs, "\n")

		fmt.Println(summary)
		fmt.Println(msg)
		if notifier != nil {
			return notifier.Notify(summary, msg)
		}
	}
	return nil
}
