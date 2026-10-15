package intervention

import "testing"

func TestLinksResolveOldAndSecondStampedIDs(t *testing.T) {
	root := t.TempDir()
	add(t, root, "first one", "run", "halt", "B1", t0)
	add(t, root, "second one", "run", "halt", "B1", t0)
	newID := "I20260102-103001-second-one"
	oldID := "I20260102-1030-old-form"
	v := Intervention{Context: "see " + newID + " and " + oldID}
	links := Links(root, v)
	if len(links) != 2 || links[0].Kind != "intervention" || links[1].Kind != "intervention" {
		t.Fatalf("links %+v", links)
	}
	if links[0].Title != "second one" {
		t.Fatalf("title %q", links[0].Title)
	}
}
