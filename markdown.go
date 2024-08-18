package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func createMarkdownFile(directory string, tree *FileNode, stats SummaryStats, extensionStats map[string]ExtensionStats) error {
	var sb strings.Builder

	sb.WriteString("# Directory Structure Analysis\n\n")

	sb.WriteString("## Directory Tree\n\n")
	sb.WriteString("```\n")
	sb.WriteString(generateMarkdownTree(tree, "", true))
	sb.WriteString("```\n\n")

	sb.WriteString("## Summary Statistics\n\n")
	sb.WriteString("| Statistic         | Value                |\n")
	sb.WriteString("|-------------------|----------------------|\n")
	sb.WriteString(fmt.Sprintf("| Total Files        | %-20d |\n", stats.TotalFiles))
	sb.WriteString(fmt.Sprintf("| Total Directories  | %-20d |\n", stats.TotalDirectories))
	sb.WriteString(fmt.Sprintf("| Total Size         | %-20s |\n", formatSize(stats.TotalSize)))
	sb.WriteString(fmt.Sprintf("| Total Lines        | %-20d |\n", stats.TotalLines))
	sb.WriteString(fmt.Sprintf("| Total Comments     | %-20d |\n", stats.TotalComments))
	sb.WriteString(fmt.Sprintf("| Total Code Lines   | %-20d |\n", stats.TotalCodeLines))
	sb.WriteString("\n")

	sb.WriteString("## File Extension Statistics\n\n")

	maxExtWidth := len("Extension")
	maxFileCountWidth := len("File Count")
	maxTotalLinesWidth := len("Total Lines")
	maxCommentsWidth := len("Comments")
	maxCodeLinesWidth := len("Code Lines")

	for ext, stat := range extensionStats {
		extName := ext
		if extName == "" {
			extName = "(no extension)"
		}
		maxExtWidth = max(maxExtWidth, len(extName))
		maxFileCountWidth = max(maxFileCountWidth, len(fmt.Sprintf("%d", stat.FileCount)))
		maxTotalLinesWidth = max(maxTotalLinesWidth, len(fmt.Sprintf("%d", stat.TotalLines)))
		maxCommentsWidth = max(maxCommentsWidth, len(fmt.Sprintf("%d", stat.TotalComments)))
		maxCodeLinesWidth = max(maxCodeLinesWidth, len(fmt.Sprintf("%d", stat.TotalCodeLines)))
	}

	headerFormat := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n",
		maxExtWidth, maxFileCountWidth, maxTotalLinesWidth, maxCommentsWidth, maxCodeLinesWidth)
	sb.WriteString(fmt.Sprintf(headerFormat, "Extension", "File Count", "Total Lines", "Comments", "Code Lines"))

	separatorFormat := fmt.Sprintf("|%%-%ds|%%-%ds|%%-%ds|%%-%ds|%%-%ds|\n",
		maxExtWidth+2, maxFileCountWidth+2, maxTotalLinesWidth+2, maxCommentsWidth+2, maxCodeLinesWidth+2)
	sb.WriteString(fmt.Sprintf(separatorFormat,
		strings.Repeat("-", maxExtWidth+2),
		strings.Repeat("-", maxFileCountWidth+2),
		strings.Repeat("-", maxTotalLinesWidth+2),
		strings.Repeat("-", maxCommentsWidth+2),
		strings.Repeat("-", maxCodeLinesWidth+2)))

	var extensions []string
	for ext := range extensionStats {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	rowFormat := fmt.Sprintf("| %%-%ds | %%%dd | %%%dd | %%%dd | %%%dd |\n",
		maxExtWidth, maxFileCountWidth, maxTotalLinesWidth, maxCommentsWidth, maxCodeLinesWidth)
	for _, ext := range extensions {
		extStat := extensionStats[ext]
		extName := ext
		if extName == "" {
			extName = "(no extension)"
		}
		sb.WriteString(fmt.Sprintf(rowFormat,
			extName, extStat.FileCount, extStat.TotalLines, extStat.TotalComments, extStat.TotalCodeLines))
	}
	sb.WriteString("\n")

	outputFile := filepath.Join(directory, "directory_structure.md")
	err := os.WriteFile(outputFile, []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("error writing markdown file: %v", err)
	}

	fmt.Printf("Markdown file created: %s\n", outputFile)
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generateMarkdownTree(node *FileNode, prefix string, isLast bool) string {
	var sb strings.Builder

	if node.Name != "" {
		sb.WriteString(prefix)
		if isLast {
			sb.WriteString("└── ")
			prefix += "    "
		} else {
			sb.WriteString("├── ")
			prefix += "│   "
		}

		if node.IsDir {
			sb.WriteString(node.Name + "/\n")
		} else {
			sb.WriteString(node.Name + "\n")
		}
	}

	for i, child := range node.Children {
		sb.WriteString(generateMarkdownTree(child, prefix, i == len(node.Children)-1))
	}

	return sb.String()
}
