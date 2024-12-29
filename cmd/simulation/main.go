package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/janmalch/roar/internal/run"
	"github.com/janmalch/roar/models"
	"github.com/janmalch/roar/pkg/git"
)

var commits = []string{
	"feat(users): add user feature",
	"feat: add new feature",
	"fix: important hotfix",
	"feat(products): add product feature",
	"docs: improve docs",
	"fix(users): fix user",
	"test: add good tests",
	"feat!: we had to break it, sorry",
}

func main() {
	runs := flag.Int("r", 10, "how many runs the simulation should perform")
	initial := flag.String("x", "", "use v1.0.0 at start of simulation")
	seed := flag.Int64("s", rand.Int63(), "seed for the RNG")
	flag.Parse()

	rng := rand.New(rand.NewSource(*seed))

	if err := os.MkdirAll(".sim/repos", os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(".sim/logs", os.ModePerm); err != nil {
		panic(err)
	}

	id := fmt.Sprintf("sim-%d", time.Now().Unix())
	repoDir := ".sim/repos/" + id
	if err := os.Mkdir(repoDir, os.ModePerm); err != nil {
		panic(err)
	}

	log, err := os.OpenFile(fmt.Sprintf(".sim/logs/%s.log", id), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	fmt.Printf("Seed: %d\nRepository: %s\nLogs: %s\nRuns: %d\n", *seed, repoDir, log.Name(), *runs)

	log.WriteString(fmt.Sprintf("%s | SEED = %d | BEGIN SIMULATION OF %d RUNS\n", fmtNow(), *seed, runs))

	defer func() {
		if r := recover(); r != nil {
			log.WriteString(fmt.Sprintf("%s | PANIC: %+v\n", fmtNow(), r))
			panic(r)
		}
	}()

	r := git.NewRepo(repoDir)
	if _, err = r.ExecGit("init", "-b", "main"); err != nil {
		panic(err)
	}
	dummy, err := os.ReadFile("testdata/openapi.yml")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(r.PathOf("openapi.yml"), dummy, os.ModePerm)
	if err != nil {
		panic(err)
	}
	if err = r.Add("openapi.yml"); err != nil {
		panic(err)
	}
	if err = r.Commit("Initial commit"); err != nil {
		panic(err)
	}

	if len(*initial) > 0 {
		makeStable(r, *initial)
	}

	c := models.Config{
		Updates: []models.UpdateConfig{
			{
				File:    "openapi.yml",
				Find:    "  version: ",
				Replace: "  version: {{version}}",
			},
		},
		Changelog: models.ChangelogConfig{
			Include:          []string{"feat", "fix", "refactor"},
			UrlCommit:        "https://github.com/JanMalch/roar/commit/{{hash}}",
			UrlBrowseAtTag:   "https://github.com/JanMalch/roar/tree/v{{version}}",
			UrlCompareTags:   "https://github.com/JanMalch/roar/compare/v{{previous}}...v{{version}}",
			UrlCommitsForTag: "https://github.com/JanMalch/roar/commits/v{{version}}",
		},
	}
	today := time.Now()

	for i := 1; i <= *runs; i++ {
		take := rng.Intn(3) + 1
		log.WriteString(fmt.Sprintf("%s | BEGIN RUN %d WITH %d commits\n", fmtNow(), i, take))

		for c := 1; c <= take; c++ {
			message := fmt.Sprintf("%s (run %d, commit %d)", commits[rng.Intn(len(commits))], i, c)
			log.WriteString(fmt.Sprintf("%s | commit: %s\n", fmtNow(), message))
			if _, err = r.ExecGit("commit", "--allow-empty", "-m", message); err != nil {
				panic(err)
			}
		}

		tag, err := run.Programmatic(
			r,
			c,
			nil,
			today,
			false,
			log,
			false,
		)
		if err != nil {
			log.WriteString(fmt.Sprintf("%s | RUN FAILED:\n%+v\n", fmtNow(), err))
			panic(err)
		}
		r.ExecGit("tag", tag)

		log.WriteString(fmt.Sprintf("%s | END RUN: %d\n\n", fmtNow(), i))
	}
}

func makeStable(r *git.Repo, x string) {
	if _, err := r.ExecGit("tag", "v"+x); err != nil {
		panic(err)
	}
}

func fmtNow() string {
	return time.Now().Format(time.RFC3339)
}
