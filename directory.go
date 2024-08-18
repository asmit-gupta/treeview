package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type FileNode struct {
	Name     string
	Children []*FileNode
	IsDir    bool
}

func buildDirectoryTree(root string, respectGitignore bool) (*FileNode, error) {
	rootInfo, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("error accessing root directory: %v", err)
	}

	rootNode := &FileNode{
		Name:  rootInfo.Name(),
		IsDir: rootInfo.IsDir(),
	}

	var matcher gitignore.Matcher
	if respectGitignore {
		matcher, err = buildGitignoreMatcher(root)
		if err != nil {
			return nil, fmt.Errorf("error building gitignore matcher: %v", err)
		}
	}

	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if respectGitignore && matcher != nil {
			if matcher.Match(strings.Split(relPath, string(os.PathSeparator)), info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		addNode(rootNode, strings.Split(relPath, string(os.PathSeparator)), info.IsDir())
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return rootNode, nil
}

func buildGitignoreMatcher(root string) (gitignore.Matcher, error) {
	var patterns []gitignore.Pattern

	gitignorePath := filepath.Join(root, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		content, err := os.ReadFile(gitignorePath)
		if err != nil {
			return nil, fmt.Errorf("error reading .gitignore: %v", err)
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, gitignore.ParsePattern(line, nil))
			}
		}
	}

	return gitignore.NewMatcher(patterns), nil
}

func addNode(parent *FileNode, pathParts []string, isDir bool) {
	if len(pathParts) == 0 {
		return
	}

	name := pathParts[0]
	for _, child := range parent.Children {
		if child.Name == name {
			if len(pathParts) > 1 {
				addNode(child, pathParts[1:], isDir)
			}
			return
		}
	}

	newNode := &FileNode{
		Name:  name,
		IsDir: isDir,
	}
	parent.Children = append(parent.Children, newNode)

	if len(pathParts) > 1 {
		addNode(newNode, pathParts[1:], isDir)
	}
}
