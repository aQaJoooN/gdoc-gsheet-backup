package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Storage interface {
	Save(path string, data []byte) error
	Commit(timestamp time.Time) error
}

func NewStorage(config BackupConfig) (Storage, error) {
	switch config.Type {
	case "local":
		return &LocalStorage{path: config.LocalPath}, nil
	case "git":
		return NewGitStorage(config)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}

type LocalStorage struct {
	path string
}

func (s *LocalStorage) Save(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *LocalStorage) Commit(timestamp time.Time) error {
	return nil
}

type GitStorage struct {
	config BackupConfig
	path   string
}

func NewGitStorage(config BackupConfig) (*GitStorage, error) {
	s := &GitStorage{
		config: config,
		path:   config.LocalPath,
	}

	if _, err := os.Stat(filepath.Join(s.path, ".git")); os.IsNotExist(err) {
		if err := s.initRepo(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *GitStorage) Save(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *GitStorage) Commit(timestamp time.Time) error {
	if err := s.runGit("add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	status, err := s.runGitOutput("status", "--porcelain")
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if strings.TrimSpace(status) == "" {
		fmt.Println("No changes to commit")
		return nil
	}

	message := strings.ReplaceAll(s.config.Git.CommitMessage, "{timestamp}", timestamp.Format("2006-01-02 15:04:05"))
	if err := s.runGit("commit", "-m", message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	if err := s.push(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	fmt.Println("✓ Changes committed and pushed")
	return nil
}

func (s *GitStorage) initRepo() error {
	if s.config.Git.Repository != "" {
		repoURL := s.getAuthenticatedURL()
		cmd := exec.Command("git", "clone", repoURL, s.path)
		if err := cmd.Run(); err != nil {
			return s.initNewRepo()
		}
		return nil
	}
	return s.initNewRepo()
}

func (s *GitStorage) initNewRepo() error {
	if err := os.MkdirAll(s.path, 0755); err != nil {
		return err
	}

	if err := s.runGit("init"); err != nil {
		return err
	}

	if s.config.Git.Repository != "" {
		if err := s.runGit("remote", "add", "origin", s.config.Git.Repository); err != nil {
			return err
		}
	}

	branch := s.config.Git.Branch
	if branch == "" {
		branch = "main"
	}
	
	if err := s.runGit("checkout", "-b", branch); err != nil {
		s.runGit("checkout", branch)
	}

	return nil
}

func (s *GitStorage) push() error {
	branch := s.config.Git.Branch
	if branch == "" {
		branch = "main"
	}

	repoURL := s.getAuthenticatedURL()
	return s.runGit("push", repoURL, branch)
}

func (s *GitStorage) getAuthenticatedURL() string {
	repo := s.config.Git.Repository
	creds := s.config.Git.Credentials

	if creds.Username != "" && creds.Token != "" {
		if strings.HasPrefix(repo, "https://") {
			return strings.Replace(repo, "https://", fmt.Sprintf("https://%s:%s@", creds.Username, creds.Token), 1)
		}
	}

	return repo
}

func (s *GitStorage) runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *GitStorage) runGitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	output, err := cmd.Output()
	return string(output), err
}
