package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	clubscore "clubscore"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr *os.File) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}
	command := args[0]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	filePath := fs.String("file", "activities.yaml", "archive path")
	id := fs.String("id", "", "activity number")
	name := fs.String("name", "", "activity name")
	leader := fs.String("leader", "", "person responsible")
	score := fs.Float64("score", 0, "score from 0 to 100")
	participants := fs.Int("participants", 0, "participant count")
	newID := fs.String("new-id", "", "replacement activity number")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	list, err := loadOrCreate(*filePath)
	if err != nil {
		return err
	}
	switch command {
	case "add":
		if err := list.Append(clubscore.Activity{ID: *id, Name: *name, Leader: *leader, Score: *score, Participants: *participants}); err != nil {
			return err
		}
		return list.SaveFile(*filePath)
	case "delete":
		if err := list.Delete(*id); err != nil {
			return err
		}
		return list.SaveFile(*filePath)
	case "update":
		if *newID == "" {
			*newID = *id
		}
		if err := list.Modify(*id, clubscore.Activity{ID: *newID, Name: *name, Leader: *leader, Score: *score, Participants: *participants}); err != nil {
			return err
		}
		return list.SaveFile(*filePath)
	case "get":
		activity, found := list.Query(*id)
		if !found {
			return fmt.Errorf("activity %s not found", *id)
		}
		fmt.Fprintf(stdout, "%s\t%s\t%s\t%.2f\t%d\n", activity.ID, activity.Name, activity.Leader, activity.Score, activity.Participants)
		return nil
	case "sort-id":
		list.SortByID()
		return list.SaveFile(*filePath)
	case "sort-score":
		list.SortByScore()
		return list.SaveFile(*filePath)
	case "list":
		for _, activity := range list.ToSlice() {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%.2f\t%d\n", activity.ID, activity.Name, activity.Leader, activity.Score, activity.Participants)
		}
		return nil
	case "help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func loadOrCreate(path string) (*clubscore.ActivityList, error) {
	list, err := clubscore.LoadFile(path)
	if err == nil {
		return list, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return clubscore.NewList(), nil
	}
	return nil, err
}

func printUsage(out *os.File) {
	const usage = `clubscore manages club activity scores.

Commands:
  add --file activities.yaml --id A1 --name "Campus Run" --leader Lin --score 92 --participants 80
  delete --file activities.yaml --id A1
  update --file activities.yaml --id A1 --name "Campus Run" --leader Lin --score 95 --participants 90
  get --file activities.yaml --id A1
  list --file activities.yaml
  sort-id --file activities.yaml
  sort-score --file activities.yaml
`
	_, _ = out.WriteString(usage)
}
