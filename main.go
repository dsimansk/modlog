package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, `Usage:
   %s [module] [from-revision] [to-revision]
`, path.Base(os.Args[0]))

		os.Exit(1)
	}

	target := os.Args[1]
	fromVersion := os.Args[2]
	toVersion := os.Args[3]

	repo := NewRepo(target)

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

		fmt.Printf("  > %s %s\n", c.Hash[:7], c.Message[:index])
	})
}

func errcheck(err error, format string, args ...interface{}) {
	if err != nil {
		panic(fmt.Errorf(format+": %w", append(args, err)...))
	}
}
