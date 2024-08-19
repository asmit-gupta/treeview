package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/spf13/cobra"
)

var respectGitignore bool

// SummaryStats holds the aggregated statistics for a directory.
// This struct summarizes key metrics such as the total number of files,
// directories, total size in bytes, total lines, lines of comments, and lines of code.
type SummaryStats struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	TotalLines       int
	TotalComments    int
	TotalCodeLines   int
}

// ExtensionStats holds statistics for a specific file extension.
// It tracks the number of files with the extension, as well as the total lines,
// lines of comments, and lines of code.
type ExtensionStats struct {
	FileCount      int
	TotalLines     int
	TotalComments  int
	TotalCodeLines int
}

var rootCmd = &cobra.Command{
	Use:   "TreeView",
	Short: "A fast directory information tool",
}

var createMdCmd = &cobra.Command{
	Use:   "create-md [directory]",
	Short: "Create a markdown file with directory details",
	Args:  cobra.ExactArgs(1),
	Run:   createMarkdown,
}

var printDirCmd = &cobra.Command{
	Use:   "print-dir [directory]",
	Short: "Print detailed directory information",
	Args:  cobra.ExactArgs(1),
	Run:   printDirectory,
}

func init() {
	rootCmd.AddCommand(createMdCmd)
	rootCmd.AddCommand(printDirCmd)
	rootCmd.PersistentFlags().BoolVar(&respectGitignore, "respect-gitignore", false, "Respect .gitignore rules")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createMarkdown(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	tree, err := buildDirectoryTreeParallel(args[0], respectGitignore)
	if err != nil {
		fmt.Printf("Error building directory tree: %v\n", err)
		return
	}

	stats, extensionStats, err := calculateStatsParallel(args[0], respectGitignore)
	if err != nil {
		fmt.Printf("Error calculating statistics: %v\n", err)
		return
	}

	err = createMarkdownFile(args[0], tree, stats, extensionStats)
	if err != nil {
		fmt.Printf("Error creating markdown: %v\n", err)
		return
	}

	elapsedTime := time.Since(startTime)
	roundedTime := roundDuration(elapsedTime)
	fmt.Printf("Markdown file created. Processing time: %v\n", roundedTime)
}

func calculateStatsParallel(rootPath string, respectGitignore bool) (SummaryStats, map[string]ExtensionStats, error) {
	stats := SummaryStats{}
	extensionStats := make(map[string]ExtensionStats)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	var matcher gitignore.Matcher
	var err error
	if respectGitignore {
		matcher, err = buildGitignoreMatcher(rootPath)
		if err != nil {
			return stats, extensionStats, fmt.Errorf("error building gitignore matcher: %v", err)
		}
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootPath, path)
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

		if info.IsDir() {
			stats.TotalDirectories++
		} else {
			stats.TotalFiles++
			stats.TotalSize += info.Size()

			ext := strings.ToLower(filepath.Ext(path))
			wg.Add(1)
			go func() {
				defer wg.Done()
				lines, comments, codeLines := analyzeFile(path)
				mutex.Lock()
				stats.TotalLines += lines
				stats.TotalComments += comments
				stats.TotalCodeLines += codeLines
				extStat := extensionStats[ext]
				extStat.FileCount++
				extStat.TotalLines += lines
				extStat.TotalComments += comments
				extStat.TotalCodeLines += codeLines
				extensionStats[ext] = extStat
				mutex.Unlock()
			}()
		}

		return nil
	})

	wg.Wait()

	return stats, extensionStats, err
}

func printDirectory(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	tree, err := buildDirectoryTreeParallel(args[0], respectGitignore)
	if err != nil {
		fmt.Printf("Error getting directory contents: %v\n", err)
		return
	}

	fmt.Println(generateMarkdownTree(tree, "", true))

	stats, extensionStats, err := calculateStatsParallel(args[0], respectGitignore)
	if err != nil {
		fmt.Printf("Error calculating statistics: %v\n", err)
	} else {
		printPrettySummaryStats(stats)
		printPrettyExtensionStats(extensionStats)
	}

	elapsedTime := time.Since(startTime)
	roundedTime := roundDuration(elapsedTime)
	fmt.Printf("\nProcessing time: %v\n", roundedTime)
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
	ext := strings.ToLower(filepath.Ext(filePath))

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		totalLines++

		if trimmedLine == "" {
			continue
		}

		switch ext {
		case ".go", ".java", ".js", ".ts", ".c", ".cpp", ".cs", ".h", ".cc", ".swift", ".kt", ".scala":
			if inBlockComment {
				if strings.Contains(trimmedLine, "*/") {
					inBlockComment = false
				}
			} else if strings.HasPrefix(trimmedLine, "//") {
				totalComments++
				continue
			} else if strings.HasPrefix(trimmedLine, "/*") {
				totalComments++
				inBlockComment = true
				if strings.Contains(trimmedLine, "*/") {
					inBlockComment = false
				}
				continue
			}

			if !inBlockComment {
				totalCodeLines++
			}
		case ".py":
			if strings.HasPrefix(trimmedLine, "#") {
				totalComments++
			} else {
				totalCodeLines++
			}
		case ".html", ".xml", ".svg":
			if strings.HasPrefix(trimmedLine, "<!--") {
				totalComments++
				if !strings.HasSuffix(trimmedLine, "-->") {
					inBlockComment = true
				}
			} else if inBlockComment {
				if strings.HasSuffix(trimmedLine, "-->") {
					inBlockComment = false
				}
			} else {
				totalCodeLines++
			}
		case ".dart":
			if inBlockComment {
				if strings.Contains(trimmedLine, "*/") {
					inBlockComment = false
				}
			} else if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "///") {
				totalComments++
				continue
			} else if strings.HasPrefix(trimmedLine, "/*") {
				totalComments++
				inBlockComment = true
				if strings.Contains(trimmedLine, "*/") {
					inBlockComment = false
				}
				continue
			}

			if !inBlockComment {
				totalCodeLines++
			}
		case ".md", ".txt", ".json", ".yaml", ".yml":
			totalCodeLines++
		default:
			// For other file types, not counting comments or code lines; because I don't know what else to do right now
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error scanning file %s: %v\n", filePath, err)
	}

	return
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

func printPrettySummaryStats(stats SummaryStats) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 40))
	fmt.Printf("%-20s %s\n", "Summary Statistics", strings.Repeat("=", 20))
	fmt.Printf("%s\n", strings.Repeat("=", 40))
	fmt.Printf("%-20s %d\n", "Total Files:", stats.TotalFiles)
	fmt.Printf("%-20s %d\n", "Total Directories:", stats.TotalDirectories)
	fmt.Printf("%-20s %s\n", "Total Size:", formatSize(stats.TotalSize))
	fmt.Printf("%-20s %d\n", "Total Lines:", stats.TotalLines)
	fmt.Printf("%-20s %d\n", "Total Comments:", stats.TotalComments)
	fmt.Printf("%-20s %d\n", "Total Code Lines:", stats.TotalCodeLines)
	fmt.Printf("%s\n", strings.Repeat("=", 40))
}

func printPrettyExtensionStats(extensionStats map[string]ExtensionStats) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("%-20s %s\n", "File Extension Statistics", strings.Repeat("=", 60))
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Printf("%-10s %-12s %-12s %-12s %-12s\n", "Extension", "File Count", "Total Lines", "Comments", "Code Lines")
	fmt.Printf("%s\n", strings.Repeat("-", 80))

	var extensions []string
	for ext := range extensionStats {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	for _, ext := range extensions {
		stats := extensionStats[ext]
		fmt.Printf("%-10s %-12d %-12d %-12d %-12d\n", ext, stats.FileCount, stats.TotalLines, stats.TotalComments, stats.TotalCodeLines)
	}
	fmt.Printf("%s\n", strings.Repeat("=", 80))
}
