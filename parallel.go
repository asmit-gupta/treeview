package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

func buildDirectoryTreeParallel(root string, respectGitignore bool) (*FileNode, error) {
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

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	var processDirParallel func(string, *FileNode)
	processDirParallel = func(dir string, node *FileNode) {
		defer wg.Done()

		entries, err := os.ReadDir(dir)
		if err != nil {
			errChan <- fmt.Errorf("error reading directory %s: %v", dir, err)
			return
		}

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				errChan <- fmt.Errorf("error getting relative path: %v", err)
				continue
			}

			if respectGitignore && matcher != nil {
				if matcher.Match(strings.Split(relPath, string(os.PathSeparator)), entry.IsDir()) {
					continue
				}
			}

			childNode := &FileNode{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
			}
			node.Children = append(node.Children, childNode)

			if entry.IsDir() {
				wg.Add(1)
				go processDirParallel(path, childNode)
			}
		}
	}

	wg.Add(1)
	go processDirParallel(root, rootNode)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return rootNode, nil
}
