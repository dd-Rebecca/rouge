package rouge

import (
	"fmt"
	"testing"
)

func Test_score(t *testing.T) {
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
