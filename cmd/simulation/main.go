package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/janmalch/roar/internal/run"
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
	runs := 10
	if err := os.MkdirAll(".sim/repos", 0644); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(".sim/logs", 0644); err != nil {
		panic(err)
	}

	id := fmt.Sprintf("sim-%d", time.Now().Unix())
	repoDir := ".sim/repos/" + id
	if err := os.Mkdir(repoDir, 0644); err != nil {
		panic(err)
	}

	log, err := os.OpenFile(fmt.Sprintf(".sim/logs/%s.log", id), os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer log.Close()
	log.WriteString(fmt.Sprintf("%s | BEGIN SIMULATION OF %d RUNS\n", fmtNow(), runs))

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
	err = os.WriteFile(r.PathOf("openapi.yml"), dummy, 0644)
	if err != nil {
		panic(err)
	}
	if err = r.Add("openapi.yml"); err != nil {
		panic(err)
	}
	if err = r.Commit("Initial commit"); err != nil {
		panic(err)
	}

	makeStable(r) // comment out for v0

	for i := 1; i <= runs; i++ {
		take := rand.Intn(3) + 1
		log.WriteString(fmt.Sprintf("%s | BEGIN RUN %d WITH %d commits\n", fmtNow(), i, take))

		for c := 1; c <= take; c++ {
			message := fmt.Sprintf("%s (run %d, commit %d)", commits[rand.Intn(len(commits))], i, c)
			log.WriteString(fmt.Sprintf("%s | commit: %s\n", fmtNow(), message))
			if _, err = r.ExecGit("commit", "--allow-empty", "-m", message); err != nil {
				panic(err)
			}
		}

		tag, err := run.Programmatic(
			r,
			"openapi.yml",
			"  version: ",
			"  version: {{version}}",
			nil,
			[]string{"feat", "fix", "refactor"},
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

func makeStable(r *git.Repo) {
	if _, err := r.ExecGit("tag", "v1.0.0"); err != nil {
		panic(err)
	}
}

func fmtNow() string {
	return time.Now().Format(time.RFC3339)
}
