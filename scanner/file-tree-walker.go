package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

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

// readFileLines safely reads a file into memory line by line.
func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// auditSingleFile runs all charter rules against a single file.
func auditSingleFile(rootDir string, path string, info os.FileInfo) ([]rules.Violation, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	if _, supported := supportedCodeExtensions[ext]; !supported {
		return nil, false
	}

	relPath, relErr := filepath.Rel(rootDir, path)
	if relErr != nil {
		relPath = path
	}

	lines, readErr := readFileLines(path)
	if readErr != nil {
		return nil, false
	}

	ctx := rules.FileContext{
		RelativePath: filepath.ToSlash(relPath),
		FileName:     info.Name(),
		Lines:        lines,
	}

	var fileViolations []rules.Violation
	fileViolations = append(fileViolations, rules.CheckFileSizeCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckNamingCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckNestingDepthCompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckSignatureSLACompliance(ctx)...)
	fileViolations = append(fileViolations, rules.CheckCodeHygieneCompliance(ctx)...)

	return fileViolations, true
}

// isDirectoryExcluded checks if a directory should be skipped.
func isDirectoryExcluded(info os.FileInfo) bool {
	if !info.IsDir() {
		return false
	}
	_, excluded := defaultExcludedDirs[info.Name()]
	return excluded
}

// ScanWorkspace collects and audits all eligible source code files in rootDir.
func ScanWorkspace(rootDir string) ([]rules.Violation, int, error) {
	var allViolations []rules.Violation
	scannedFilesCount := 0

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if isDirectoryExcluded(info) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		fileViolations, ok := auditSingleFile(rootDir, path, info)
		if !ok {
			return nil
		}

		scannedFilesCount++
		allViolations = append(allViolations, fileViolations...)
		return nil
	})

	return allViolations, scannedFilesCount, err
}
