package entries

import (
	"log"
	"path"
	"os"
	"time"

	"github.com/stardustapp/core/base"
	"github.com/stardustapp/core/extras"
	"github.com/stardustapp/core/inmem"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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
	workingPath, _ := extras.GetChildString(inputFolder, "working-path")
	repoPath, _ := extras.GetChildString(inputFolder, "repo-path")
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
	originUri, _ := extras.GetChildString(inputFolder, "origin-uri")
	repoPath, _ := extras.GetChildString(inputFolder, "repo-path")
	workingPath, _ := extras.GetChildString(inputFolder, "working-path")
	//idempotent, _ := extras.GetChildString(inputFolder, "idempotent")

	// delete the target
	ctx.Put(workingPath, nil)
	ctx.Put(workingPath, inmem.NewFolder(path.Base(workingPath)))
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

	worktree, err := repo.Worktree()
	if err != nil {
		log.Println("[git] worktree problem:", err)
		return nil
	}

	return &gitApi{
		repo:     repo,
		worktree: worktree,
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

	worktree, err := repo.Worktree()
	if err != nil {
		log.Println("[git] worktree problem:", err)
		return nil
	}

	return &gitApi{
		repo:     repo,
		worktree: worktree,
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

	worktree, err := repo.Worktree()
	if err != nil {
		log.Println("[git] worktree problem:", err)
		return nil
	}

	return &gitApi{
		repo:     repo,
		worktree: worktree,
	}
}

// Combines a working tree with a git repo
// Presents a set of APIs to invoke Git functionality
type gitApi struct {
	repo     *git.Repository
	worktree *git.Worktree
}

var _ base.Folder = (*gitApi)(nil)

func (e *gitApi) Name() string {
	return "git"
}

func (e *gitApi) Children() []string {
	return []string{
		"status",
		"add",
		"remove",
		"commit",
		"push",
		"pull",
	}
}

func (e *gitApi) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "status":
		return inmem.NewFolderOf("status",
			&gitStatusFunc{e, e.worktree},
			gitStatusShape,
			stringOutputShape,
		).Freeze(), true

	case "add":
		return inmem.NewFolderOf("add",
			&gitAddFunc{e, e.worktree},
			gitAddShape,
			stringOutputShape,
		).Freeze(), true

	case "remove":
		return inmem.NewFolderOf("remove",
			&gitRemoveFunc{e, e.worktree},
			gitRemoveShape,
			stringOutputShape,
		).Freeze(), true

	case "commit":
		return inmem.NewFolderOf("commit",
			&gitCommitFunc{e, e.worktree},
			gitCommitShape,
			stringOutputShape,
		).Freeze(), true

	case "push":
		return inmem.NewFolderOf("push",
			&gitPushFunc{e, e.repo},
			gitPushShape,
			stringOutputShape,
		).Freeze(), true

	case "pull":
		return inmem.NewFolderOf("pull",
			&gitPullFunc{e, e.repo},
			gitPullShape,
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
	api      *gitApi
	worktree *git.Worktree
}

var _ base.Function = (*gitStatusFunc)(nil)

func (e *gitStatusFunc) Name() string {
	return "invoke"
}

func (e *gitStatusFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	status, err := e.worktree.Status()
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
	api      *gitApi
	worktree *git.Worktree
}

var _ base.Function = (*gitAddFunc)(nil)

func (e *gitAddFunc) Name() string {
	return "invoke"
}

func (e *gitAddFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	path, _ := extras.GetChildString(inputFolder, "path")

	defer func() {
		if r := recover(); r != nil {
			if r != os.ErrNotExist {
				panic(r)
			}
			log.Println("git add on missing file", path)
			output = nil
		}
	}()

	hash, err := e.worktree.Add(path)
	if err != nil {
		panic(err)
	}

	return inmem.NewString("hash", hash.String())
}

var gitRemoveShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("path", "String"),
		),
	))

type gitRemoveFunc struct {
	api      *gitApi
	worktree *git.Worktree
}

var _ base.Function = (*gitRemoveFunc)(nil)

func (e *gitRemoveFunc) Name() string {
	return "invoke"
}

func (e *gitRemoveFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	path, _ := extras.GetChildString(inputFolder, "path")

	hash, err := e.worktree.Remove(path)
	if err != nil {
		panic(err)
	}

	return inmem.NewString("hash", hash.String())
}

var gitCommitShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewString("message", "String"),
			inmem.NewString("author-name", "String"),
			inmem.NewString("author-email", "String"),
			inmem.NewFolderOf("all",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
		),
	))

type gitCommitFunc struct {
	api      *gitApi
	worktree *git.Worktree
}

var _ base.Function = (*gitCommitFunc)(nil)

func (e *gitCommitFunc) Name() string {
	return "invoke"
}

func (e *gitCommitFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	message, _ := extras.GetChildString(inputFolder, "message")
	authorName, _ := extras.GetChildString(inputFolder, "author-name")
	authorEmail, _ := extras.GetChildString(inputFolder, "author-email")
	allStr, _ := extras.GetChildString(inputFolder, "all")
	all := allStr == "yes"

	author := &object.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}

	hash, err := e.worktree.Commit(message, &git.CommitOptions{
		All:       all,
		Author:    author,
		Committer: author,
	})
	if err != nil {
		panic(err)
	}

	return inmem.NewString("commit-hash", hash.String())
}

var gitPushShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewFolderOf("remote-name",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			//inmem.NewString("remote-name", "String"),
			// TODO: list of refspec strings
		),
	))

type gitPushFunc struct {
	api  *gitApi
	repo *git.Repository
}

var _ base.Function = (*gitPushFunc)(nil)

func (e *gitPushFunc) Name() string {
	return "invoke"
}

func (e *gitPushFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	remoteName, _ := extras.GetChildString(inputFolder, "remote-name")

	// https doesn't have push yet
	err := e.repo.Push(&git.PushOptions{
		RemoteName: remoteName,
	})
	if err != nil {
		log.Println("git push:", err)
		return inmem.NewString("error", err.Error())
	}
	return inmem.NewString("result", "ok")
}

var gitPullShape *inmem.Shape = inmem.NewShape(
	inmem.NewFolderOf("input-shape",
		inmem.NewString("type", "Folder"),
		inmem.NewFolderOf("props",
			inmem.NewFolderOf("remote-name",
				inmem.NewString("type", "String"),
				inmem.NewString("optional", "yes"),
			),
			//inmem.NewString("remote-name", "String"),
			// TODO: list of refspec strings
		),
	))

type gitPullFunc struct {
	api  *gitApi
	repo *git.Repository
}

var _ base.Function = (*gitPullFunc)(nil)

func (e *gitPullFunc) Name() string {
	return "invoke"
}

func (e *gitPullFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
	inputFolder := input.(base.Folder)
	remoteName, _ := extras.GetChildString(inputFolder, "remote-name")

	err := e.repo.Pull(&git.PullOptions{
		RemoteName: remoteName,
	})
	if err != nil {
		log.Println("git pull:", err)
		return inmem.NewString("error", err.Error())
	}
	return inmem.NewString("result", "ok")
}
