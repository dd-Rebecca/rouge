package rouge

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Rouge struct {
	returnLengths bool
	rawResults    bool
	exclusive     bool
	metrics       []string
	stats         []string
}

var DefaultMetrics = []string{"rouge-1", "rouge-2", "rouge-l"}
var DefaultStats = []string{"r", "p", "f"}
var AvailableMetrics = map[string]func([]string, []string, bool) map[string]float64{
	"rouge-1": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 1, false, exclusive)
	},
	"rouge-2": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 2, false, exclusive)
	},
	"rouge-l": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeLSummaryLevel(hyp, ref, false, exclusive)
	},
}

func NewRouge(metrics, stats []string, returnLengths, rawResults, exclusive bool) (*Rouge, error) {
	r := &Rouge{
		returnLengths: returnLengths,
		rawResults:    rawResults,
		exclusive:     exclusive,
	}

	if metrics != nil {
		r.metrics = metrics
		for _, m := range r.metrics {
			if _, ok := AvailableMetrics[m]; !ok {
				return nil, errors.New(fmt.Sprintf("Unknown metric '%s'", m))
			}
		}
	} else {
		r.metrics = DefaultMetrics
	}

	if rawResults {
		r.stats = []string{"hyp", "ref", "overlap"}
	} else {
		if stats != nil {
			r.stats = stats
			for _, s := range r.stats {
				if !contains(DefaultStats, s) {
					return nil, errors.New(fmt.Sprintf("Unknown stat '%s'", s))
				}
			}
		} else {
			r.stats = DefaultStats
		}
	}

	return r, nil
}

func contains(slice []string, item string) bool {
	for _, each := range slice {
		if each == item {
			return true
		}
	}
	return false
}

func (r *Rouge) GetScores(hyps, refs []string, avg, ignoreEmpty bool) (interface{}, error) {
	if reflect.TypeOf(hyps) != reflect.TypeOf(refs) {
		return nil, errors.New("hypotheses and references are not of the same type")
	}

	if ignoreEmpty {
		filteredHyps := []string{}
		filteredRefs := []string{}
		for i := 0; i < len(hyps); i++ {
			if len(hyps[i]) > 0 && len(refs[i]) > 0 {
				filteredHyps = append(filteredHyps, hyps[i])
				filteredRefs = append(filteredRefs, refs[i])
			}
		}
		hyps = filteredHyps
		refs = filteredRefs
	}

	if len(hyps) != len(refs) {
		return nil, errors.New("the number of hypotheses and references must be equal")
	}

	if !avg {
		scores, err := r.getScores(hyps, refs)
		if err != nil {
			return nil, err
		}
		return scores, nil
	}
	return r.getAvgScores(hyps, refs)
}

func (r *Rouge) getScores(hyps, refs []string) ([]map[string]map[string]float64, error) {
	scores := []map[string]map[string]float64{}

	for i := 0; i < len(hyps); i++ {
		senScore := map[string]map[string]float64{}
		hyp := splitSentences(hyps[i])
		ref := splitSentences(refs[i])

		for _, m := range r.metrics {
			fn := AvailableMetrics[m]
			sc := fn(hyp, ref, r.exclusive)
			senScore[m] = make(map[string]float64)
			for _, s := range r.stats {
				senScore[m][s] = sc[s]
			}
		}

		if r.returnLengths {
			lengths := map[string]float64{
				"hyp": float64(len(strings.Fields(strings.Join(hyp, " ")))),
				"ref": float64(len(strings.Fields(strings.Join(ref, " ")))),
			}
			senScore["lengths"] = lengths
		}

		scores = append(scores, senScore)
	}

	return scores, nil
}

func (r *Rouge) getAvgScores(hyps, refs []string) (map[string]map[string]float64, error) {
	scores := make(map[string]map[string]float64)
	for _, m := range r.metrics {
		scores[m] = make(map[string]float64)
		for _, s := range r.stats {
			scores[m][s] = 0
		}
	}

	count := 0
	for i := 0; i < len(hyps); i++ {
		hyp := splitSentences(hyps[i])
		ref := splitSentences(refs[i])

		for _, m := range r.metrics {
			fn := AvailableMetrics[m]
			sc := fn(hyp, ref, r.exclusive)
			for _, s := range r.stats {
				scores[m][s] += sc[s]
			}
		}

		count++
	}

	avgScores := make(map[string]map[string]float64)
	for m := range scores {
		avgScores[m] = make(map[string]float64)
		for s := range scores[m] {
			avgScores[m][s] = scores[m][s] / float64(count)
		}
	}

	return avgScores, nil
}

func splitSentences(text string) []string {
	sentences := strings.Split(text, ".")
	processedSentences := []string{}
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 0 {
			processedSentences = append(processedSentences, trimmed)
		}
	}
	return processedSentences
}
