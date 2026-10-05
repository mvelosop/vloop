package main

import "testing"

// A review session judges the work; what it changes outside .vloop/tmp/ is
// reverted and fails the verdict.

func TestRunReviewChangesReverted(t *testing.T) {
	t.Parallel()
	r := newRunRepo(t)
	r.commitFile("src.txt", "orig\n")
	r.scripted(oneTask(t), defaultScript+`if [ "$PHASE" = review ] && [ "$ATTEMPT" = 1 ]; then echo tampered >> src.txt; echo new > review-new.txt; fi
if [ "$PHASE" = review ]; then echo scratch > .vloop/tmp/review-notes.txt; fi
`)
	wantExit(t, r.runWith(nil, runBrief), 0)
	r.wantIterations("T1:rejected", "T1:done")
	log := r.runLog()
	wantIn(t, "run log", log, "vloop: the review session changed src.txt — reverted")
	wantIn(t, "run log", log, "vloop: the review session changed review-new.txt — reverted")
	if got := r.read("src.txt"); got != "orig\n" {
		t.Errorf("src.txt not reverted: %q", got)
	}
	if out := r.git("ls-files", "review-new.txt"); out != "" {
		t.Errorf("the file the review created was committed: %q", out)
	}
}
