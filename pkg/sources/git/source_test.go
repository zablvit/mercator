package git_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	git2 "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zablvit/mercator/pkg/sources/git"
)

const knownHostsVar = "SSH_KNOWN_HOSTS"

type Source interface {
	Clone(repoUrl string, branch string, project string, options git.CloneOptions) error
}

func TestShouldCloneSourceRepositoryForTheFirstTime(t *testing.T) {
	tests := map[string]struct {
		givenProjectRoot  string
		givenRepoUrl      string
		givenBranch       string
		givenCloneOptions git.CloneOptions
		expectedError     error
		expectedFiles     []string
	}{
		"should fetch zablvit/zero repo from github": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://github.com/zablvit/zero",
			givenBranch:      "master",
			expectedError:    nil,
			expectedFiles: []string{
				".gitignore",
				"Dockerfile",
				"LICENSE",
				"README.md",
				"zero.asm",
			},
		},
		"should fetch zablvit/mercator-test main branch from github": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://github.com/zablvit/mercator-test",
			givenBranch:      "main",
			expectedError:    nil,
			expectedFiles: []string{
				"LICENSE",
				"README.md",
			},
		},
		"should fetch zablvit/mercator-test test_folder_structure branch from github": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://github.com/zablvit/mercator-test",
			givenBranch:      "test_folder_structure",
			expectedError:    nil,
			expectedFiles: []string{
				"LICENSE",
				"README.md",
				"dir1",
				filepath.Join("dir1", "file1"),
				filepath.Join("dir1", "file2"),
				filepath.Join("dir1", "file3"),
				"dir2",
				filepath.Join("dir2", "dirA"),
				filepath.Join("dir2", "dirA", "dirAA"),
				filepath.Join("dir2", "dirA", "dirAA", "file4"),
			},
		},
		"should fail on non existing branch": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://github.com/zablvit/mercator-test",
			givenBranch:      "non-existing-branch",
			expectedError:    fmt.Errorf("reference not found"),
			expectedFiles:    []string{},
		},
		"should fail on non existing github repository": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://github.com/zablvit/non-existing-repo",
			givenBranch:      "",
			expectedError:    fmt.Errorf("authentication required"),
			expectedFiles:    []string{},
		},
		"should fail on non existing gitlab repository": {
			givenProjectRoot: filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:     "https://gitlab.com/zablvit/non-existing-repo",
			givenBranch:      "",
			expectedError:    fmt.Errorf("authentication required"),
			expectedFiles:    []string{},
		},
		"should fetch zablvit/mercator-test-private main branch from github using rsa key": {
			givenProjectRoot:  filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:      "git@github.com:zablvit/mercator-test-private.git",
			givenBranch:       "main",
			givenCloneOptions: git.CloneOptions{PemBytes: []byte(testPrivateRSAKey)},
			expectedError:     nil,
			expectedFiles: []string{
				"LICENSE",
				"README.md",
				"test-file",
			},
		},
		"should fetch zablvit/mercator-test-private main branch from github using ed25519 key": {
			givenProjectRoot:  filepath.Join(os.TempDir(), "mercator", "projects", "proj1"),
			givenRepoUrl:      "git@github.com:zablvit/mercator-test-private.git",
			givenBranch:       "main",
			givenCloneOptions: git.CloneOptions{PemBytes: []byte(testPrivateED25519Key)},
			expectedError:     nil,
			expectedFiles: []string{
				"LICENSE",
				"README.md",
				"test-file",
			},
		},
	}
	var source Source

	for name, tt := range tests {
		tt := tt
		name := name
		t.Run(name, func(t *testing.T) {
			hostsFile, _ := filepath.Abs("../../../etc/known_hosts")
			fmt.Println(hostsFile)
			_ = os.Setenv(knownHostsVar, hostsFile)

			defer func(path string) {
				_ = os.RemoveAll(path)
				_ = os.Unsetenv(knownHostsVar)
			}(filepath.Join(os.TempDir(), "mercator"))

			source = git.New()
			actualErr := source.Clone(tt.givenRepoUrl, tt.givenBranch, tt.givenProjectRoot, tt.givenCloneOptions)
			assert.Equal(t, tt.expectedError, actualErr)

			files := make([]string, 0)

			require.NoError(t, filepath.WalkDir(tt.givenProjectRoot,
				func(path string, d fs.DirEntry, err error) error {
					rel, err := filepath.Rel(tt.givenProjectRoot, path)
					if err != nil {
						return err
					}
					if strings.Split(rel, string(filepath.Separator))[0] != ".git" && rel != "." {
						files = append(files, rel)
					}
					return nil
				},
			))

			assert.Equal(t, tt.expectedFiles, files)
		})
	}
}

func TestShouldFailCloneOnAlreadyExistingRepository(t *testing.T) {
	tempTestDir := filepath.Join(os.TempDir(), "mercatorTestClone")
	defer func() {
		_ = os.RemoveAll(tempTestDir)
	}()

	_, err := git2.PlainInit(tempTestDir, false)
	require.NoError(t, err, "Should init empty repository in empty dir - setup is not clean")

	source := git.New()
	err = source.Clone("", "", tempTestDir, git.CloneOptions{})
	require.NotNil(t, err, "Expected non nil error")
	assert.EqualError(t, err, "repository already exists")
}

const (
	testPrivateRSAKey = `-----BEGIN OPENSSH PRIVATE KEY-----
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
-----END OPENSSH PRIVATE KEY-----
`
	testPrivateED25519Key = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACAFuquIP6KTx0SmK9WBPwdwv8s/OywBw7VpZ9jJk+GVOwAAAJhzH4dQcx+H
UAAAAAtzc2gtZWQyNTUxOQAAACAFuquIP6KTx0SmK9WBPwdwv8s/OywBw7VpZ9jJk+GVOw
AAAEAhj3jG3tNPfS0U0MaKzVzmNecv4RJOGZDiP8bUis10VAW6q4g/opPHRKYr1YE/B3C/
yz87LAHDtWln2MmT4ZU7AAAAEWVtYWlsQGV4YW1wbGUuY29tAQIDBA==
-----END OPENSSH PRIVATE KEY-----
`
)
