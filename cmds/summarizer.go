package cmds

import (
	"regexp"
	"sort"
	"strings"
)

var stopWords = map[string]bool{
	"a": true, "about": true, "above": true, "after": true, "again": true, "against": true,
	"all": true, "am": true, "an": true, "and": true, "any": true, "are": true, "as": true,
	"at": true, "be": true, "because": true, "been": true, "before": true, "being": true,
	"below": true, "between": true, "both": true, "but": true, "by": true, "could": true,
	"did": true, "do": true, "does": true, "doing": true, "down": true, "during": true,
	"each": true, "few": true, "for": true, "from": true, "further": true, "had": true,
	"has": true, "have": true, "having": true, "he": true, "her": true, "here": true,
	"hers": true, "herself": true, "him": true, "himself": true, "his": true, "how": true,
	"i": true, "if": true, "in": true, "into": true, "is": true, "it": true, "its": true,
	"itself": true, "just": true, "me": true, "more": true, "most": true, "my": true,
	"myself": true, "no": true, "nor": true, "not": true, "now": true, "of": true, "off": true,
	"on": true, "once": true, "only": true, "or": true, "other": true, "our": true, "ours": true,
	"ourselves": true, "out": true, "over": true, "own": true, "same": true, "she": true,
	"should": true, "so": true, "some": true, "such": true, "than": true, "that": true,
	"the": true, "their": true, "theirs": true, "them": true, "themselves": true, "then": true,
	"there": true, "these": true, "they": true, "this": true, "those": true, "through": true,
	"to": true, "too": true, "under": true, "until": true, "up": true, "very": true, "was": true,
	"we": true, "were": true, "what": true, "when": true, "where": true, "which": true,
	"while": true, "who": true, "whom": true, "why": true, "with": true, "would": true,
	"you": true, "your": true, "yours": true, "yourself": true, "yourselves": true,
}

type sentenceScore struct {
	index    int
	sentence string
	score    float64
}

// SummarizeText extracts key sentences by word frequency in pure Go
func SummarizeText(text string, maxSentences int) string {
	if maxSentences <= 0 {
		maxSentences = 10
	}

	// Clean markdown formatting
	reCode := regexp.MustCompile("(?s)```.*?```")
	cleaned := reCode.ReplaceAllString(text, "")
	reBadges := regexp.MustCompile(`\[!\[.*?\]\(.*?\)\]\(.*?\)`)
	cleaned = reBadges.ReplaceAllString(cleaned, "")
	reLinks := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	cleaned = reLinks.ReplaceAllString(cleaned, "$1")

	rawSentences := splitSentences(cleaned)
	if len(rawSentences) == 0 {
		return ""
	}
	if len(rawSentences) <= maxSentences {
		return strings.Join(rawSentences, ".\n") + "."
	}

	// Calculate word frequencies
	wordFreq := map[string]int{}
	wordRegex := regexp.MustCompile(`\b[a-zA-Z0-9_-]{2,}\b`)
	for _, s := range rawSentences {
		words := wordRegex.FindAllString(strings.ToLower(s), -1)
		for _, w := range words {
			if !stopWords[w] {
				wordFreq[w]++
			}
		}
	}

	// Score sentences
	scores := make([]sentenceScore, 0, len(rawSentences))
	for idx, s := range rawSentences {
		words := wordRegex.FindAllString(strings.ToLower(s), -1)
		if len(words) < 3 {
			continue
		}
		var totalScore float64
		for _, w := range words {
			totalScore += float64(wordFreq[w])
		}
		score := totalScore / float64(len(words))
		scores = append(scores, sentenceScore{index: idx, sentence: strings.TrimSpace(s), score: score})
	}

	if len(scores) == 0 {
		limit := maxSentences
		if len(rawSentences) < limit {
			limit = len(rawSentences)
		}
		return strings.Join(rawSentences[:limit], ".\n") + "."
	}

	// Sort by score descending to find top sentences
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	top := scores
	if len(top) > maxSentences {
		top = top[:maxSentences]
	}

	// Re-sort by original appearance order
	sort.Slice(top, func(i, j int) bool {
		return top[i].index < top[j].index
	})

	var result []string
	for _, sc := range top {
		s := sc.sentence
		if !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "!") && !strings.HasSuffix(s, "?") {
			s += "."
		}
		result = append(result, s)
	}

	return strings.Join(result, "\n")
}

func splitSentences(text string) []string {
	var sentences []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := regexp.MustCompile(`[.!?]+\s+`).Split(line, -1)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if len(p) > 10 {
				sentences = append(sentences, p)
			}
		}
	}
	return sentences
}
