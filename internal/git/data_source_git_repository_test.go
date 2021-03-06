package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceGitRepository_new(t *testing.T) {
	dir, err := ioutil.TempDir("", "acctest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fs := osfs.New(dir)
	dot, _ := fs.Chroot("storage")
	storage := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	repo, err := git.Init(storage, fs)
	if err != nil {
		t.Fatal(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello world!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = worktree.Add("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	commit, err := worktree.Commit("test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Terraform User",
			Email: "terraform@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	obj, err := repo.CommitObject(commit)
	if err != nil {
		t.Fatal(err)
	}
	hash := obj.Hash.String()

	path := filepath.ToSlash(dir)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitRepositoryPathConfig(path),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.git_repository.test", "url", ""),
					resource.TestCheckResourceAttr("data.git_repository.test", "branch", "master"),
					resource.TestCheckResourceAttr("data.git_repository.test", "commit_sha", hash),
				),
			},
		},
	})
}

func TestAccDataSourceGitRepository_path(t *testing.T) {
	url := "https://github.com/innovationnorway/terraform-git-module-acctest.git"
	dir, err := ioutil.TempDir("", "acctest-*")
	if err != nil {
		t.Fatal(err)
	}
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.ToSlash(dir)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitRepositoryPathConfig(path),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.git_repository.test", "url", "https://github.com/innovationnorway/terraform-git-module-acctest.git"),
					resource.TestCheckResourceAttr("data.git_repository.test", "branch", "master"),
					resource.TestCheckResourceAttr("data.git_repository.test", "commit_sha", "750e948880ebe167eba524dded790e8b9a79d01d"),
				),
			},
		},
	})
}

func TestAccDataSourceGitRepository_branch(t *testing.T) {
	url := "https://github.com/innovationnorway/terraform-git-module-acctest.git"
	branch := "master"

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitRepositoryBranchConfig(url, branch),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.git_repository.test", "url", "https://github.com/innovationnorway/terraform-git-module-acctest.git"),
					resource.TestCheckResourceAttr("data.git_repository.test", "branch", "master"),
					resource.TestCheckResourceAttr("data.git_repository.test", "commit_sha", "750e948880ebe167eba524dded790e8b9a79d01d"),
				),
			},
		},
	})
}

func TestAccDataSourceGitRepository_tag(t *testing.T) {
	url := "https://github.com/innovationnorway/terraform-git-module-acctest.git"
	tag := "v0.1.0"

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitRepositoryTagConfig(url, tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.git_repository.test", "url", "https://github.com/innovationnorway/terraform-git-module-acctest.git"),
					resource.TestCheckResourceAttr("data.git_repository.test", "tag", "v0.1.0"),
					resource.TestCheckResourceAttr("data.git_repository.test", "commit_sha", "750e948880ebe167eba524dded790e8b9a79d01d"),
				),
			},
		},
	})
}

func testAccDataSourceGitRepositoryPathConfig(path string) string {
	return fmt.Sprintf(`
data git_repository "test" {
  path = "%s"
}
`, path)
}

func testAccDataSourceGitRepositoryBranchConfig(url string, branch string) string {
	return fmt.Sprintf(`
data git_repository "test" {
  url   = "%s"
  branch = "%s" 
}
`, url, branch)
}

func testAccDataSourceGitRepositoryTagConfig(url string, tag string) string {
	return fmt.Sprintf(`
data git_repository "test" {
  url = "%s"
  tag  = "%s" 
}
`, url, tag)
}
