package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

func createMarkdownFile(directory string, tree *FileNode, stats SummaryStats, extensionStats map[string]ExtensionStats) error {
	var sb strings.Builder

	// Write the header
	sb.WriteString("# Directory Structure Analysis\n\n")

	// Write the directory tree
	sb.WriteString("## Directory Tree\n\n")
	sb.WriteString("```\n")
	sb.WriteString(generateMarkdownTree(tree, "", true))
	sb.WriteString("```\n\n")

	// Write the summary statistics
	sb.WriteString("## Summary Statistics\n\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-20s %d\n", "Total Files:", stats.TotalFiles))
	sb.WriteString(fmt.Sprintf("%-20s %d\n", "Total Directories:", stats.TotalDirectories))
	sb.WriteString(fmt.Sprintf("%-20s %s\n", "Total Size:", formatSize(stats.TotalSize)))
	sb.WriteString(fmt.Sprintf("%-20s %d\n", "Total Lines:", stats.TotalLines))
	sb.WriteString(fmt.Sprintf("%-20s %d\n", "Total Comments:", stats.TotalComments))
	sb.WriteString(fmt.Sprintf("%-20s %d\n", "Total Code Lines:", stats.TotalCodeLines))
	sb.WriteString("```\n")

	// Write the file extension statistics
	sb.WriteString("\n## File Extension Statistics\n\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("%-10s %-12s %-12s %-12s %-12s\n", "Extension", "File Count", "Total Lines", "Comments", "Code Lines"))
	sb.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", 80)))

	// Create a sorted list of extensions
	var extensions []string
	for ext := range extensionStats {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	// Write the sorted extension statistics
	for _, ext := range extensions {
		extStat := extensionStats[ext]
		sb.WriteString(fmt.Sprintf("%-10s %-12d %-12d %-12d %-12d\n", ext, extStat.FileCount, extStat.TotalLines, extStat.TotalComments, extStat.TotalCodeLines))
	}
	sb.WriteString("```\n")

	// Write the content to the file
	outputFile := filepath.Join(directory, "directory_structure.md")
	err := os.WriteFile(outputFile, []byte(sb.String()), 0644)
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
