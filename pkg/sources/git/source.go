package git

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Source struct {
}

type CloneOptions struct {
	PemBytes    []byte
	PemPassword string
}

func (g Source) Clone(repoUrl string, branch string, projectRoot string, options CloneOptions) error {
	projectRootCleanPath := filepath.Clean(projectRoot)
	if _, err := os.Stat(projectRootCleanPath); os.IsNotExist(err) {
		err := os.MkdirAll(projectRootCleanPath, os.ModeDir)
		if err != nil {
			return err
		}
	}

	var auth ssh.AuthMethod = nil

	if len(options.PemBytes) > 0 {
		var err error
		auth, err = ssh.NewPublicKeys("git", options.PemBytes, options.PemPassword)
		if err != nil {
			return err
		}
	}

	_, err := git.PlainClone(projectRoot, false, &git.CloneOptions{
		URL:           repoUrl,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Auth:          auth,
	})
	if err != nil {
		return err
	}

	return nil
}

func New() Source {
	return Source{}
}
