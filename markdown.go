package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func createMarkdownFile(directory string) error {
	tree, err := buildDirectoryTreeParallel(directory, respectGitignore)
	if err != nil {
		return err
	}

	mdContent := generateMarkdownTree(tree, "", true)

	outputFile := filepath.Join(directory, "directory_structure.md")
	err = os.WriteFile(outputFile, []byte(mdContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing markdown file: %v", err)
	}

	fmt.Printf("Markdown file created: %s\n", outputFile)
	return nil
}

func generateMarkdownTree(node *FileNode, prefix string, isLast bool) string {
	var sb strings.Builder

	if node.Name != "" {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("|_ ")
			prefix += "   "
		} else {
			sb.WriteString("|_ ")
			prefix += "|  "
		}

		if node.IsDir {
			color.New(color.FgBlue, color.Bold).Fprintf(&sb, "%s", node.Name)
		} else {
			color.New(color.FgGreen).Fprintf(&sb, "%s", node.Name)
		}

		sb.WriteString("\n")
	}

	for i, child := range node.Children {
		sb.WriteString(generateMarkdownTree(child, prefix, i == len(node.Children)-1))
	}

	return sb.String()
}
