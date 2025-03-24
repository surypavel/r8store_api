package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// GetFileByVersion retrieves a specific file version from Git
func GetFileByVersion(store, extension, version, username, password string) (interface{}, error) {
	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	// Initialize an in-memory repository
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:        store,
		Tags:       git.NoTags,
		Auth:       auth,
		NoCheckout: true,
	})
	if err != nil {
		return "", err
	}

	// Fetch only the specific tag
	tagPrefix := fmt.Sprintf("ext/%s/v", extension)
	tagRef := plumbing.NewTagReferenceName(fmt.Sprintf("%s%s", tagPrefix, version))
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:%s", tagRef, tagRef))},
	})

	if err != nil {
		return "", err
	}

	// Resolve the tag
	ref, err := repo.Reference(tagRef, true)
	if err != nil {
		return "", err
	}

	// Get commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", err
	}

	// Get the tree
	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	meta, err := getFileFromTree(repo, tree, fmt.Sprintf("%s/meta.json", extension))

	if err != nil {
		return "", err
	}

	var dat interface{}
	if err := json.Unmarshal([]byte(meta), &dat); err != nil {
		return "", err
	}

	codeSource := dat.(map[string]interface{})["config"].(map[string]interface{})["code_source"]

	code, err := getFileFromTree(repo, tree, fmt.Sprintf("%s/%s", extension, codeSource))

	if err != nil {
		return "", err
	}

	dat.(map[string]interface{})["config"].(map[string]interface{})["code"] = code

	return dat, nil
}

func getFileFromTree(repo *git.Repository, tree *object.Tree, path string) (string, error) {
	// Locate the file in the tree
	file := fmt.Sprintf("dist/%s", path)

	entry, err := tree.FindEntry(file)
	if err != nil {
		return "", err
	}

	// Get the file blob
	blob, err := repo.BlobObject(entry.Hash)
	if err != nil {
		return "", err
	}

	// Read file contents
	reader, err := blob.Reader()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GetVersions(store, extension string) ([]string, error) {
	// Run git ls-remote to fetch only matching tags
	tagPrefix := fmt.Sprintf("ext/%s/v", extension)
	cmd := exec.Command("git", "ls-remote", "--tags", "--refs", store, tagPrefix+"*")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return make([]string, 0), err
	}

	var versions []*semver.Version
	tagRegex := regexp.MustCompile(`refs/tags/` + regexp.QuoteMeta(tagPrefix) + `(\d+\.\d+\.\d+(?:[-+.\w]*)?)$`)

	// Process output lines
	for _, line := range strings.Split(out.String(), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		tagRef := parts[1] // refs/tags/vX.Y.Z

		matches := tagRegex.FindStringSubmatch(tagRef)
		if matches != nil {
			ver, err := semver.NewVersion(matches[1])
			if err == nil {
				versions = append(versions, ver)
			}
		}
	}

	// If no valid tags found
	if len(versions) == 0 {
		return make([]string, 0), fmt.Errorf("no matching tags found with prefix '%s'", tagPrefix)
	}

	// Sort versions in descending order
	sort.Sort(sort.Reverse(semver.Collection(versions)))

	versionStrings := make([]string, 0)
	for _, version := range versions {
		versionStrings = append(versionStrings, version.String())
	}
	return versionStrings, nil
}

func GetStoreHandler(store string, username, password string) (map[string]interface{}, error) {
	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	// Clone the repository into memory with sparse checkout
	storer := memory.NewStorage()
	fs := memfs.New()
	repo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:           store,
		Auth:          auth,
		Depth:         1, // Shallow clone
		SingleBranch:  true,
		ReferenceName: plumbing.ReferenceName("refs/heads/main"),
		NoCheckout:    true, // No full checkout
	})
	if err != nil {
		log.Fatal(err)
	}

	// Configure sparse checkout to fetch only "dist/*/meta.json"
	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch:                    "refs/heads/main",
		Force:                     true,
		SparseCheckoutDirectories: []string{"dist"},
	})
	if err != nil {
		log.Fatal(err)
	}

	return findAndReadMetaFiles(worktree)
}

func findAndReadMetaFiles(worktree *git.Worktree) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	basePath := "dist"
	files, err := worktree.Filesystem.ReadDir(basePath)
	if err != nil {
		log.Fatal("Failed to read dist directory:", err)
	}

	for _, dir := range files {
		if !dir.IsDir() {
			continue
		}
		metaFilePath := filepath.Join(basePath, dir.Name(), "meta.json")

		// Try to open meta.json if it exists
		f, err := worktree.Filesystem.Open(metaFilePath)
		if err != nil {
			return make(map[string]interface{}), err
		}
		defer f.Close()

		// Read file contents into memory
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(f)
		if err != nil {
			return make(map[string]interface{}), err
		}

		s := buf.Bytes()
		var dat map[string]interface{}

		if err := json.Unmarshal(s, &dat); err != nil {
			return make(map[string]interface{}), err
		}

		m[dir.Name()] = dat
	}

	return m, nil
}
