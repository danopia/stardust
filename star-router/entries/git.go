package entries

import (
	"log"

  "gopkg.in/src-d/go-git.v4"
  "gopkg.in/src-d/go-git.v4/storage/filesystem"
	"github.com/danopia/stardust/star-router/base"
	//"github.com/danopia/stardust/star-router/helpers"
	"github.com/danopia/stardust/star-router/inmem"
)

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

func gitClone(ctx base.Context, input base.Entry) (output base.Entry) {
  repoTree := &billyAdapter{ctx, "/n/aws-ns/app-suite.git"}
  workTree := &billyAdapter{ctx, "/n/aws-ns/app-suite"}

  repoStore, err := filesystem.NewStorage(repoTree)
  if err != nil {
    log.Println("[git] clone storage problem:", err)
    return nil
  }

	repo, err := git.Clone(repoStore, workTree, &git.CloneOptions{
    URL: "https://github.com/stardustapp/app-suite",
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
  repoTree := &billyAdapter{ctx, "/n/aws-ns/app-suite.git"}
  workTree := &billyAdapter{ctx, "/n/aws-ns/app-suite"}

  repoStore, err := filesystem.NewStorage(repoTree)
  if err != nil {
    log.Println("[git] init storage problem:", err)
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
  repoTree := &billyAdapter{ctx, "/n/aws-ns/app-suite.git"}
  workTree := &billyAdapter{ctx, "/n/aws-ns/app-suite"}

  repoStore, err := filesystem.NewStorage(repoTree)
  if err != nil {
    log.Println("[git] open storage problem:", err)
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
	return []string{"status"}
}

func (e *gitApi) Fetch(name string) (entry base.Entry, ok bool) {
	switch name {

	case "status":
		return &gitStatusFunc{e.repo}, true

	default:
		return
	}
}

func (e *gitApi) Put(name string, entry base.Entry) (ok bool) {
	return false
}


type gitStatusFunc struct {
	repo *git.Repository
}

var _ base.Function = (*gitStatusFunc)(nil)

func (e *gitStatusFunc) Name() string {
	return "status"
}

func (e *gitStatusFunc) Invoke(ctx base.Context, input base.Entry) (output base.Entry) {
  // TODO
  return nil
}
