package main

import (
	"fmt"
	"time"
)

func NeedBackup(config Config, statePath string) error {
	state, err := LoadState(statePath)
	var outdated []OutdatedBackup
	if err != nil {
		outdated = ComputeOutdated(config, State{})
	} else {
		outdated = ComputeOutdated(config, *state)
	}
	reportOutdated(outdated)
	return nil
}

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

type OutdatedBackup struct {
	Name      string
	SinceDays int
	Never     bool
}

func (b OutdatedBackup) String() string {
	if b.Never {
		return fmt.Sprintf("%s: never done", b.Name)
	}
	return fmt.Sprintf("%s: last backup was %d days ago", b.Name, b.SinceDays)
}

func reportOutdated(outdated []OutdatedBackup) {
	if len(outdated) == 0 {
		fmt.Println("Backups are up to date")
	} else {
		fmt.Printf("The following backups are outdated:\n\n")
		for _, b := range outdated {
			fmt.Printf("  - %v\n", b)
		}
	}
}
