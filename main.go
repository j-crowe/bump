package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/haya14busa/go-versionsort"
)

type Level int

const (
	PATCH Level = iota
	MINOR
	MAJOR
	CURRENT // Do not bump and show the latest version.
)

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:	bump [major,minor,patch (default=patch)]
bump returns next semantic version tag`)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "https://github.com/haya14busa/bump")
	os.Exit(2)
}

func run(w io.Writer) error {
	flag.Usage = usage
	flag.Parse()
	ctx := context.Background()
	tags, err := tags(ctx)
	if err != nil {
		return err
	}
	if len(tags) == 0 {
		return errors.New("existing tag not found")
	}
	latest := latestSemVer(tags)
	next, err := nextTag(latest, bumpLevel(flag.Args()))
	if err != nil {
		return err
	}
	fmt.Fprintln(w, next)
	return nil
}

func nextTag(latest string, level Level) (string, error) {
	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return "", err
	}
	next := nextSemVer(latestVer, level)
	return next.Original(), nil
}

func tags(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "tag")
	b, err := cmd.CombinedOutput()
	fmt.Fprintln(os.Stdout, b)

	cmd2 := exec.CommandContext(ctx, "git", "branch")
	b2, _ := cmd2.CombinedOutput()
	fmt.Fprintln(os.Stdout, b2)

	if err != nil {
		return nil, fmt.Errorf("failed to run `git tag`: %v", err)
	}
	if len(b) == 0 {
		return nil, nil
	}
	return strings.Split(string(b), "\n"), nil
}

func latestSemVer(tags []string) string {
	latest := ""
	for _, tag := range tags {
		if versionsort.Less(latest, tag) {
			latest = tag
		}
	}
	return latest
}

func nextSemVer(v *semver.Version, level Level) semver.Version {
	switch level {
	case PATCH:
		return v.IncPatch()
	case MINOR:
		return v.IncMinor()
	case MAJOR:
		return v.IncMajor()
	case CURRENT:
		// Do nothing.
		return *v
	}
	log.Fatalf("unknown level: %v", level)
	return v.IncPatch()
}

func bumpLevel(args []string) Level {
	if len(args) == 0 {
		return PATCH
	}
	switch args[0] {
	case "patch":
		return PATCH
	case "minor":
		return MINOR
	case "major":
		return MAJOR
	case "current":
		return CURRENT
	}
	return PATCH
}
