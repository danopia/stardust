package entries

import (
	"log"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

var stringOutputShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("output-shape",
		inmem.NewString("type", "String"),
	))

// Directory containing the clone function
func getGitDriver() *inmem.Folder {
	return inmem.NewFolderOf("git",

		inmem.NewFolderOf("clone",
			inmem.NewFunction("invoke", gitClone),
			//inmem.NewLink("input-shape", "/rom/shapes/git-opts"),
		).Freeze(),

		inmem.NewFolderOf("init",
			inmem.NewFunction("invoke", gitInit),
			//inmem.NewLink("input-shape", "/rom/shapes/git-opts"),
		).Freeze(),

		inmem.NewFolderOf("open",
			inmem.NewFunction("invoke", gitOpen),
			//inmem.NewLink("input-shape", "/rom/shapes/git-opts"),
		).Freeze(),
	).Freeze()
}

func inflateGitInput(ctx base.Context, input base.Entry) (*filesystem.Storage, *billyAdapter) {
	inputFolder := input.(base.Folder)
	workingPath, _ := helpers.GetChildString(inputFolder, "working-path")
	repoPath, _ := helpers.GetChildString(inputFolder, "repo-path")
	if repoPath == "" {
		repoPath = workingPath + "/.git"
	}

	repoTree := newBillyAdapter(ctx, repoPath)
	workTree := newBillyAdapter(ctx, workingPath)

	repoStore, err := filesystem.NewStorage(repoTree)
	if err != nil {
		log.Println("[git] clone storage problem:", err)
		return repoStore, nil
	}

	return repoStore, workTree
}

func gitClone(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	originUri, _ := helpers.GetChildString(inputFolder, "origin-uri")
	repoPath, _ := helpers.GetChildString(inputFolder, "repo-path")
	//idempotent, _ := helpers.GetChildString(inputFolder, "idempotent")

	// delete the target
	ctx.Put(repoPath, nil)
	ctx.Put(repoPath, inmem.NewFolder("git"))

	repoStore, workTree := inflateGitInput(ctx, input)
	if repoStore == nil || workTree == nil {
		log.Println("[git] clone storage problem:")
		return nil
	}

	repo, err := git.Clone(repoStore, workTree, &git.CloneOptions{
		URL: originUri,
	})
	if err != nil {
		log.Println("[git] clone problem:", err)
		return nil
	}

	return &gitApi{
		repo: repo,
	}
}

func gitInit(ctx base.Context, input base.Entry) (output base.Entry) {
	repoStore, workTree := inflateGitInput(ctx, input)
	if repoStore == nil || workTree == nil {
		log.Println("[git] init storage problem:")
		return nil
	}

	repo, err := git.Init(repoStore, workTree)
	if err != nil {
		log.Println("[git] init problem:", err)
		return nil
	}

	return &gitApi{
		repo: repo,
	}
}

func gitOpen(ctx base.Context, input base.Entry) (output base.Entry) {
	repoStore, workTree := inflateGitInput(ctx, input)
	if repoStore == nil || workTree == nil {
		log.Println("[git] open storage problem:")
		return nil
	}

	repo, err := git.Open(repoStore, workTree)
	if err != nil {
		log.Println("[git] open problem:", err)
		return nil
	}

	return &gitApi{
		repo: repo,
	}
}

// Combines a working tree with a git repo
// Presents a set of APIs to invoke Git functionality
type gitApi struct {
	repo *git.Repository
}

var _ base.Folder = (*gitApi)(nil)

func (e *gitApi) Name() string {
	return "git"
}

func (e *gitApi) Children() []string {
	return []string{"status", "add"}
}

func (e *gitApi) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "status":
		return inmem.NewFolderOf("status",
			&gitStatusFunc{e.repo},
			gitStatusShape,
			stringOutputShape,
		).Freeze(), true

	case "add":
		return inmem.NewFolderOf("add",
			&gitAddFunc{e.repo},
			gitAddShape,
			stringOutputShape,
		).Freeze(), true

	default:
		return
	}
}

func (e *gitApi) Put(name string, entry base.Entry) (ok bool) {
	return false
}

var gitStatusShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props"),
	))

type gitStatusFunc struct {
	repo *git.Repository
}

var _ base.Function = (*gitStatusFunc)(nil)

func (e *gitStatusFunc) Name() string {
	return "invoke"
}

func (e *gitStatusFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	w, err := e.repo.Worktree()
	if err != nil {
		panic(err)
	}

	status, err := w.Status()
	if err != nil {
		panic(err)
	}

	return inmem.NewString("status", status.String())
}

var gitAddShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("path", "String"),
		),
	))

type gitAddFunc struct {
	repo *git.Repository
}

var _ base.Function = (*gitAddFunc)(nil)

func (e *gitAddFunc) Name() string {
	return "invoke"
}

func (e *gitAddFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	w, err := e.repo.Worktree()
	if err != nil {
		panic(err)
	}

	inputFolder := input.(base.Folder)
	path, _ := helpers.GetChildString(inputFolder, "path")

	hash, err := w.Add(path)
	if err != nil {
		panic(err)
	}

	return inmem.NewString("hash", hash.String())
}
