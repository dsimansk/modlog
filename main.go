package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var single = flag.Bool("s", false, "print logs for a single module")

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, `Usage:
   %s [-s] module from-revision to-revision

	   -s - prints revision logs for a single module
`, path.Base(os.Args[0]))

		os.Exit(1)
	}

	target := args[0]
	fromVersion := args[1]
	toVersion := args[2]

	repo := NewRepo(target)

	if *single {
		log(target, fromVersion, toVersion)
		return
	}

	from := ResolveGoModFile(repo.GoModule(fromVersion))
	to := ResolveGoModFile(repo.GoModule(toVersion))

	for name, info := range to {
		old, ok := from[name]

		if !ok {
			// newly added
			continue
		}

		if old.Version != info.Version {
			log(info.Path, old.Version, info.Version)
		}
	}
}

func GoModVersionToRevision(v string) string {
	if strings.HasPrefix(v, "v0.0.0") {
		return strings.Split(v, "-")[2]
	}

	return v
}

func ResolveGoModFile(f *modfile.File) map[string]module.Version {
	result := make(map[string]module.Version)

	for _, entry := range f.Require {
		result[entry.Mod.Path] = entry.Mod
	}

	for _, entry := range f.Replace {
		result[entry.Old.Path] = entry.New
	}

	return result
}

func log(mod, from, to string) {
	from = GoModVersionToRevision(from)
	to = GoModVersionToRevision(to)

	repo := NewRepo(mod)
	start := repo.HashFor(from)
	end := repo.HashFor(to)

	fmt.Printf("bumping %s %s...%s:\n", mod, start[:7], end[:7])

	repo.Log(start, end, func(c Commit) {
		index := strings.Index(c.Message, "\n")
		if index == -1 {
			index = len(c.Message)
		}

		// We don't want to link the issue #'s since that'll be noisy
		// Github will spam repos with references based on the `#{num}`
		// appearing in the commit message - this replace avoids that
		// by changing it to `# {num}`
		message := ghNum.ReplaceAllString(c.Message[:index], "# $1")

		fmt.Printf("  > %s %s\n", c.Hash[:7], message)
	})
}

func errcheck(err error, format string, args ...interface{}) {
	if err != nil {
		panic(fmt.Errorf(format+": %w", append(args, err)...))
	}
}

var ghNum = regexp.MustCompile(`#(\d+)`)
