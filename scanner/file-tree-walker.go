package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/elm-as/elmasguard/rules"
)

var defaultExcludedDirs = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
	"bin":          {},
	"obj":          {},
	"dist":         {},
	"build":        {},
	".vscode":      {},
	".idea":        {},
	"coverage":     {},
}

var supportedCodeExtensions = map[string]struct{}{
	".go":    {},
	".ts":    {},
	".tsx":   {},
	".js":    {},
	".jsx":   {},
	".py":    {},
	".rs":    {},
	".c":     {},
	".h":     {},
	".cpp":   {},
	".hpp":   {},
	".cs":    {},
	".fs":    {},
	".dart":  {},
	".php":   {},
	".rb":    {},
	".java":  {},
	".kt":    {},
	".swift": {},
}

type fileCandidate struct {
	fullPath string
	relPath  string
	fileName string
}

// loadGitignorePatterns reads basic directory/file ignore patterns from .gitignore if present.
func loadGitignorePatterns(rootDir string) []string {
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	lines, err := readFileLines(gitignorePath)
	if err != nil {
		return nil
	}

	var patterns []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "#") {
			continue
		}
		patterns = append(patterns, strings.TrimSuffix(trimmed, "/"))
	}
	return patterns
}

// isIgnoredByGitignore checks whether a relative path matches a .gitignore entry.
func isIgnoredByGitignore(relPath string, baseName string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, baseName); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}

// readFileLines safely reads a file into memory line by line.
func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	bufScanner := bufio.NewScanner(file)
	for bufScanner.Scan() {
		lines = append(lines, bufScanner.Text())
	}
	return lines, bufScanner.Err()
}

// auditSingleFile runs all charter rules against a single file candidate.
func auditSingleFile(candidate fileCandidate) ([]rules.Violation, bool) {
	lines, readErr := readFileLines(candidate.fullPath)
	if readErr != nil {
		return nil, false
	}

	ctx := rules.FileContext{
		RelativePath: candidate.relPath,
		FileName:     candidate.fileName,
		Lines:        lines,
	}

	var fileViolations []rules.Violation
	fileViolations = append(fileViolations, rules.CheckFileSizeCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckNamingCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckCodeHygieneCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckCouplingAndSecurityCompliance(ctx)...)

	if astViolations, handledByAst := rules.CheckGoAstCompliance(ctx); handledByAst {
		fileViolations = append(fileViolations, astViolations...)
		return fileViolations, true
	}

	fileViolations = append(fileViolations, rules.CheckNestingDepthCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckSignatureSLACompliance(ctx)...)
	return fileViolations, true
}

// collectEligibleFiles walks rootDir and returns all eligible source files.
func collectEligibleFiles(rootDir string) ([]fileCandidate, error) {
	ignorePatterns := loadGitignorePatterns(rootDir)
	var candidates []fileCandidate

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, relErr := filepath.Rel(rootDir, path)
		if relErr != nil {
			relPath = path
		}
		slashRel := filepath.ToSlash(relPath)

		if info.IsDir() {
			if _, excluded := defaultExcludedDirs[info.Name()]; excluded {
				return filepath.SkipDir
			}
			if isIgnoredByGitignore(slashRel, info.Name(), ignorePatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if _, supported := supportedCodeExtensions[ext]; !supported {
			return nil
		}
		if isIgnoredByGitignore(slashRel, info.Name(), ignorePatterns) {
			return nil
		}

		candidates = append(candidates, fileCandidate{
			fullPath: path,
			relPath:  slashRel,
			fileName: info.Name(),
		})
		return nil
	})

	return candidates, err
}

// sortViolationsDeterministically sorts violations by file path and line number.
func sortViolationsDeterministically(violations []rules.Violation) {
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].FilePath != violations[j].FilePath {
			return violations[i].FilePath < violations[j].FilePath
		}
		return violations[i].LineNumber < violations[j].LineNumber
	})
}

// ScanWorkspace audits all eligible source files concurrently using a worker pool.
func ScanWorkspace(rootDir string) ([]rules.Violation, int, error) {
	candidates, err := collectEligibleFiles(rootDir)
	if err != nil {
		return nil, 0, err
	}

	workerCount := runtime.NumCPU()
	if workerCount < 2 {
		workerCount = 2
	}

	jobs := make(chan fileCandidate, len(candidates))
	results := make(chan []rules.Violation, len(candidates))

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				if v, ok := auditSingleFile(candidate); ok {
					results <- v
				}
			}
		}()
	}

	for _, c := range candidates {
		jobs <- c
	}
	close(jobs)

	wg.Wait()
	close(results)

	var allViolations []rules.Violation
	scannedCount := 0
	for v := range results {
		scannedCount++
		allViolations = append(allViolations, v...)
	}

	sortViolationsDeterministically(allViolations)
	return allViolations, scannedCount, nil
}
