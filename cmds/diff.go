package cmds

import (
	"fmt"
	"os"
)

func Diff() {
	config, err := LoadBitConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dependencyfile = []string{}
	currentLanguages := iterateToFindLanguage()
	currentDeps := TrackDependentFile()

	fmt.Println("Comparing current project against .bitconfig:")
	fmt.Println()

	hasChanges := false

	if !stringSlicesEqual(config.Languages, currentLanguages) {
		hasChanges = true
		fmt.Println("Language changes:")
		printSliceDiff("  added", subtractStrings(currentLanguages, config.Languages))
		printSliceDiff("  removed", subtractStrings(config.Languages, currentLanguages))
		fmt.Println()
	}

	depChanges := diffDependencies(config.Dependencies, currentDeps)
	if len(depChanges) > 0 {
		hasChanges = true
		fmt.Println("Dependency file changes:")
		for _, change := range depChanges {
			fmt.Println(" -", change)
		}
		fmt.Println()
	}

	if !hasChanges {
		fmt.Println("No differences found. Your project matches .bitconfig.")
		fmt.Println("Run 'bitconfig update' only when you add or change dependencies.")
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
		if seen[s] < 0 {
			return false
		}
	}
	return true
}

func subtractStrings(from, remove []string) []string {
	removeSet := map[string]bool{}
	for _, s := range remove {
		removeSet[s] = true
	}
	var result []string
	for _, s := range from {
		if !removeSet[s] {
			result = append(result, s)
		}
	}
	return result
}

func printSliceDiff(label string, items []string) {
	if len(items) == 0 {
		return
	}
	for _, item := range items {
		fmt.Printf("  %s: %s\n", label, item)
	}
}

func diffDependencies(stored, current map[string]int) []string {
	var changes []string

	for name, currentLines := range current {
		storedLines, exists := stored[name]
		if !exists {
			changes = append(changes, fmt.Sprintf("%s (new file, %d lines)", name, currentLines))
			continue
		}
		if storedLines != currentLines {
			changes = append(changes, fmt.Sprintf("%s (lines: %d → %d)", name, storedLines, currentLines))
		}
	}

	for name := range stored {
		if _, exists := current[name]; !exists {
			changes = append(changes, fmt.Sprintf("%s (removed)", name))
		}
	}

	return changes
}
