package main

import (
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/vcs"
)

type Commit struct {
	Message string
	Hash    string
}

type ModuleRepo interface {
	GoModule(rev string) *modfile.File
	Log(from, to string, callback func(Commit))
	URL() string
	HashFor(rev string) string
}

func NewRepo(target string) ModuleRepo {
	// Local repo
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		repo, err := git.PlainOpen(target)
		errcheck(err, "failed to open local repo")
		return &gitRepo{repo, target}
	}

	// Normal Repo
	root, err := vcs.RepoRootForImportDynamic(target, false)
	errcheck(err, "failed to resolve module to repo")

	dir, err := ioutil.TempDir("", "clone-")
	errcheck(err, "failed to setup temp dir for cloning")

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: root.Repo,
	})

	errcheck(err, "failed to clone %s", root.Repo)
	return &gitRepo{repo, root.Repo}
}

type gitRepo struct {
	repo *git.Repository
	url  string
}

func (g *gitRepo) URL() string {
	return g.url
}

func (g *gitRepo) HashFor(rev string) string {
	hash, err := g.repo.ResolveRevision(plumbing.Revision(rev))
	errcheck(err, "failed to resolve %q to hash for repo %s", rev, g.url)
	return hash.String()
}

func (g *gitRepo) GoModule(rev string) *modfile.File {
	if rev == "dirty" {
		tree, err := g.repo.Worktree()
		errcheck(err, "failed to open local go.mod file")
		file, err := tree.Filesystem.Open("go.mod")
		errcheck(err, "failed to open local go.mod file")

		defer file.Close()

		content, err := ioutil.ReadAll(file)
		errcheck(err, "failed to read local go.mod file")

		modfile, err := modfile.Parse("go.mod", content, nil)
		errcheck(err, "failed to parse go.mod")

		return modfile
	}

	hash, err := g.repo.ResolveRevision(plumbing.Revision(rev))
	errcheck(err, "failed to resolve %q to hash for repo %s", rev, g.url)

	commit, err := g.repo.CommitObject(*hash)
	errcheck(err, "failed to find commit for %s@%s", g.url, rev)

	tree, err := commit.Tree()
	errcheck(err, "failed to get tree")

	treefile, err := tree.File("go.mod")
	errcheck(err, "failed to get go.mod %s@%s", g.url, rev)

	contents, err := treefile.Contents()
	errcheck(err, "failed to get go.mod contents %s@%s", g.url, rev)

	content := []byte(contents)
	file, err := modfile.Parse("go.mod", content, nil)
	errcheck(err, "failed to parse go.mod %s@%s", g.url, rev)
	return file
}

func (g *gitRepo) Log(from, to string, callback func(Commit)) {
	end := plumbing.NewHash(to)
	start := plumbing.NewHash(from)

	// In the case 'start' is not a ancestor commit of 'end' we'll
	// print parent commits that are unique to 'end'
	exclude := make(map[plumbing.Hash]struct{})
	excludeIter, err := g.repo.Log(&git.LogOptions{From: start})
	errcheck(err, "failed to create iterator for repo %s", g.url)
	defer excludeIter.Close()

	err = excludeIter.ForEach(func(c *object.Commit) error {
		exclude[c.Hash] = struct{}{}
		return nil
	})

	errcheck(err, "failed to iterate over commits")

	var validFunc object.CommitFilter = func(c *object.Commit) bool {
		_, ok := exclude[c.Hash]
		return !ok
	}

	commit, err := g.repo.CommitObject(end)
	errcheck(err, "failed to get commit for hash")

	iter := object.NewFilterCommitIter(commit, &validFunc, nil)
	defer iter.Close()

	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == start {
			return storer.ErrStop
		}

		callback(Commit{
			Message: c.Message,
			Hash:    c.Hash.String(),
		})

		return nil
	})
	errcheck(err, "failed to iterate over commits")
}
