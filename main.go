package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var respectGitignore bool

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type SummaryStats struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	TotalLines       int
	TotalComments    int
	TotalCodeLines   int
}

var rootCmd = &cobra.Command{
	Use:   "TreeView",
	Short: "A fast directory information tool",
}

var createMdCmd = &cobra.Command{
	Use:   "create-md [directory]",
	Short: "Create a markdown file with directory details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createMarkdown(args[0])
	},
}

var printDirCmd = &cobra.Command{
	Use:   "print-dir [directory]",
	Short: "Print detailed directory information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		printDirectory(args[0])
	},
}

func init() {
	rootCmd.AddCommand(createMdCmd)
	rootCmd.AddCommand(printDirCmd)

	rootCmd.PersistentFlags().BoolVar(&respectGitignore, "respect-gitignore", false, "Respect .gitignore rules")
}

func createMarkdown(directory string) {
	err := createMarkdownFile(directory)
	if err != nil {
		fmt.Printf("Error creating markdown: %v\n", err)
	}
}

func printDirectory(directory string) {
	startTime := time.Now()

	tree, err := buildDirectoryTreeParallel(directory, respectGitignore)
	if err != nil {
		fmt.Printf("Error getting directory contents: %v\n", err)
		return
	}

	fmt.Println(generateMarkdownTree(tree, "", true))

	stats, err := calculateStats(directory)
	if err != nil {
		fmt.Printf("Error calculating statistics: %v\n", err)
	} else {
		printSummaryStats(stats)
	}

	elapsedTime := time.Since(startTime)
	roundedTime := roundDuration(elapsedTime)
	fmt.Printf("\nProcessing time: %v\n", roundedTime)
}

func calculateStats(rootPath string) (SummaryStats, error) {
	stats := SummaryStats{}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			stats.TotalDirectories++
		} else {
			stats.TotalFiles++
			stats.TotalSize += info.Size()

			if isTextFile(path) {
				lines, comments, codeLines := analyzeFile(path)
				stats.TotalLines += lines
				stats.TotalComments += comments
				stats.TotalCodeLines += codeLines
			}
		}

		return nil
	})

	return stats, err
}

func isTextFile(filePath string) bool {
	// Check file extensions (you can expand this list based on your needs)
	extensions := []string{".go", ".py", ".dart", ".js", ".java", ".c", ".cpp", ".sh", ".txt"}
	for _, ext := range extensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}

func analyzeFile(filePath string) (totalLines, totalComments, totalCodeLines int) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inBlockComment := false

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "#") {
			totalComments++
		} else if strings.HasPrefix(trimmedLine, "/*") {
			totalComments++
			inBlockComment = true
		} else if inBlockComment {
			totalComments++
			if strings.HasSuffix(trimmedLine, "*/") {
				inBlockComment = false
			}
		} else {
			totalCodeLines++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file %s: %v\n", filePath, err)
	}

	return
}

func printSummaryStats(stats SummaryStats) {
	fmt.Printf("\nSummary Statistics:\n")
	fmt.Printf("Total Files: %d\n", stats.TotalFiles)
	fmt.Printf("Total Directories: %d\n", stats.TotalDirectories)
	fmt.Printf("Total Size: %s\n", formatSize(stats.TotalSize))
	fmt.Printf("Total Lines: %d\n", stats.TotalLines)
	fmt.Printf("Total Comments: %d\n", stats.TotalComments)
	fmt.Printf("Total Code Lines: %d\n", stats.TotalCodeLines)
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func roundDuration(d time.Duration) time.Duration {
	return time.Duration(math.Round(float64(d)/float64(time.Millisecond))) * time.Millisecond
}
