package rouge

import (
	"fmt"
	"strings"
)

type Ngrams struct {
	ngrams    map[string]int
	exclusive bool
}

func NewNgrams(exclusive bool) *Ngrams {
	return &Ngrams{ngrams: make(map[string]int), exclusive: exclusive}
}

func (n *Ngrams) Add(o string) {
	if n.exclusive {
		n.ngrams[o] = 1
	} else {
		n.ngrams[o]++
	}
}

func (n *Ngrams) Len() int {
	return len(n.ngrams)
}

func (n *Ngrams) Intersection(o *Ngrams) *Ngrams {
	intersection := NewNgrams(n.exclusive)
	for k := range n.ngrams {
		if _, ok := o.ngrams[k]; ok {
			intersection.Add(k)
		}
	}
	return intersection
}

func (n *Ngrams) BatchAdd(o []string) {
	for _, v := range o {
		n.Add(v)
	}
}

func (n *Ngrams) Union(others ...*Ngrams) *Ngrams {
	union := NewNgrams(n.exclusive)
	for k := range n.ngrams {
		union.Add(k)
	}
	for _, other := range others {
		for k := range other.ngrams {
			union.Add(k)
		}
	}
	return union
}

func getNgrams(n int, text []string, exclusive bool) *Ngrams {
	ngramSet := NewNgrams(exclusive)
	for i := 0; i <= len(text)-n; i++ {
		ngramSet.Add(strings.Join(text[i:i+n], " "))
	}
	return ngramSet
}

func splitIntoWords(sentences []string) []string {
	words := []string{}
	for _, sentence := range sentences {
		words = append(words, strings.Split(sentence, " ")...)
	}
	return words
}

func getWordNgrams(n int, sentences []string, exclusive bool) *Ngrams {
	words := splitIntoWords(sentences)
	return getNgrams(n, words, exclusive)
}

func lenLcs(x, y []string) int {
	table := lcs(x, y)
	return table[len(x)][len(y)]
}

func lcs(x, y []string) [][]int {
	n, m := len(x), len(y)
	table := make([][]int, n+1)
	for i := range table {
		table[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if x[i-1] == y[j-1] {
				table[i][j] = table[i-1][j-1] + 1
			} else {
				table[i][j] = max(table[i-1][j], table[i][j-1])
			}
		}
	}
	return table
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func reconLcs(x, y []string, exclusive bool) *Ngrams {
	i, j := len(x), len(y)
	table := lcs(x, y)

	var _recon func(int, int) []string
	_recon = func(i, j int) []string {
		if i == 0 || j == 0 {
			return []string{}
		} else if x[i-1] == y[j-1] {
			return append(_recon(i-1, j-1), x[i-1])
		} else if table[i-1][j] > table[i][j-1] {
			return _recon(i-1, j)
		} else {
			return _recon(i, j-1)
		}
	}

	reconList := _recon(i, j)
	ngramList := NewNgrams(exclusive)
	for _, word := range reconList {
		ngramList.Add(word)
	}
	return ngramList
}

func rougeN(evaluatedSentences, referenceSentences []string, n int, rawResults, exclusive bool) map[string]float64 {
	if len(evaluatedSentences) == 0 {
		panic("Hypothesis is empty.")
	}
	if len(referenceSentences) == 0 {
		panic("Reference is empty.")
	}

	evaluatedNgrams := getWordNgrams(n, evaluatedSentences, exclusive)
	referenceNgrams := getWordNgrams(n, referenceSentences, exclusive)
	referenceCount := referenceNgrams.Len()
	evaluatedCount := evaluatedNgrams.Len()

	overlappingNgrams := evaluatedNgrams.Intersection(referenceNgrams)
	overlappingCount := overlappingNgrams.Len()

	results := make(map[string]float64)
	if rawResults {
		results["hyp"] = float64(evaluatedCount)
		results["ref"] = float64(referenceCount)
		results["overlap"] = float64(overlappingCount)
		return results
	} else {
		return calculateRougeN(evaluatedCount, referenceCount, overlappingCount)
	}
}

func calculateRougeN(evaluatedCount, referenceCount, overlappingCount int) map[string]float64 {
	results := make(map[string]float64)
	if evaluatedCount == 0 {
		results["p"] = 0.0
	} else {
		results["p"] = float64(overlappingCount) / float64(evaluatedCount)
	}

	if referenceCount == 0 {
		results["r"] = 0.0
	} else {
		results["r"] = float64(overlappingCount) / float64(referenceCount)
	}

	results["f"] = 2.0 * ((results["p"] * results["r"]) / (results["p"] + results["r"] + 1e-8))
	return results
}

func unionLcs(evaluatedSentences []string, referenceSentence string, prevUnion *Ngrams, exclusive bool) (int, *Ngrams) {
	if prevUnion == nil {
		prevUnion = NewNgrams(exclusive)
	}

	if len(evaluatedSentences) == 0 {
		panic("Collections must contain at least 1 sentence.")
	}

	lcsUnion := prevUnion
	prevCount := len(prevUnion.ngrams)
	referenceWords := splitIntoWords([]string{referenceSentence})

	combinedLcsLength := 0
	for _, evalS := range evaluatedSentences {
		evaluatedWords := splitIntoWords([]string{evalS})
		lcs := reconLcs(referenceWords, evaluatedWords, exclusive)
		combinedLcsLength += lcs.Len()
		lcsUnion = lcsUnion.Union(lcs)
	}

	newLcsCount := lcsUnion.Len() - prevCount
	return newLcsCount, lcsUnion
}

func rougeLSummaryLevel(
	evaluatedSentences, referenceSentences []string, rawResults, exclusive bool) map[string]float64 {
	if len(evaluatedSentences) == 0 || len(referenceSentences) == 0 {
		panic("Collections must contain at least 1 sentence.")
	}

	referenceNgrams := NewNgrams(exclusive)
	referenceNgrams.BatchAdd(splitIntoWords(referenceSentences))
	m := referenceNgrams.Len()

	evaluatedNgrams := NewNgrams(exclusive)
	evaluatedNgrams.BatchAdd(splitIntoWords(evaluatedSentences))
	n := evaluatedNgrams.Len()

	unionLcsSumAcrossAllReferences := 0
	union := NewNgrams(exclusive)
	for _, refS := range referenceSentences {
		lcsCount, newUnion := unionLcs(evaluatedSentences, refS, union, exclusive)
		union = newUnion
		unionLcsSumAcrossAllReferences += lcsCount
	}

	llcs := unionLcsSumAcrossAllReferences
	rLcs := float64(llcs) / float64(m)
	pLcs := float64(llcs) / float64(n)

	fLcs := 2.0 * ((pLcs * rLcs) / (pLcs + rLcs + 1e-8))

	results := make(map[string]float64)
	if rawResults {
		results["hyp"] = float64(n)
		results["ref"] = float64(m)
		results["overlap"] = float64(llcs)
		return results
	} else {
		results["f"] = fLcs
		results["p"] = pLcs
		results["r"] = rLcs
		return results
	}
}

func main() {
	// 示例
	rouge, err := NewRouge(nil, nil, false, false, true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	hyps := []string{"the transcript is a written version of each day 's cnn student news programK"}
	refs := []string{"this page includes the show transcript use the transcript to help students with reading comprehension and vocabulary at the bottom of the page"}

	scores, err := rouge.GetScores(refs, hyps, false, false)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Scores:", scores)
}
