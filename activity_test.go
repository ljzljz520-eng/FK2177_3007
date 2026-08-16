package clubscore

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func fixtureActivity(id string, score float64) Activity {
	return Activity{ID: id, Name: "Activity " + id, Leader: "Leader " + id, Score: score, Participants: int(score) + 10}
}

func TestActivityBusinessChain(t *testing.T) {
	list := NewList()
	if err := list.Append(fixtureActivity("C-03", 72)); err != nil {
		t.Fatal(err)
	}
	if err := list.Append(fixtureActivity("C-01", 88)); err != nil {
		t.Fatal(err)
	}
	if err := list.Append(fixtureActivity("C-02", 95)); err != nil {
		t.Fatal(err)
	}
	if got, ok := list.Query("C-01"); !ok || got.Score != 88 {
		t.Fatalf("query returned %#v, %v", got, ok)
	}
	updated := fixtureActivity("C-04", 91)
	if err := list.Modify("C-01", updated); err != nil {
		t.Fatal(err)
	}
	if _, ok := list.Query("C-01"); ok {
		t.Fatal("old id is still queryable")
	}
	if got, ok := list.Query("C-04"); !ok || got.Participants != updated.Participants {
		t.Fatalf("updated activity missing: %#v, %v", got, ok)
	}
	list.SortByID()
	ids := []string{list.ToSlice()[0].ID, list.ToSlice()[1].ID, list.ToSlice()[2].ID}
	if want := []string{"C-02", "C-03", "C-04"}; !equalStrings(ids, want) {
		t.Fatalf("id order %v, want %v", ids, want)
	}
	list.SortByScore()
	ordered := list.ToSlice()
	if ordered[0].ID != "C-02" || ordered[2].ID != "C-03" {
		t.Fatalf("score order %#v", ordered)
	}
	var saved bytes.Buffer
	if err := list.Save(&saved); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(&saved)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.ToSlice(); !equalActivities(got, ordered) {
		t.Fatalf("loaded %#v, want %#v", got, ordered)
	}
	if err := loaded.Delete("C-03"); err != nil {
		t.Fatal(err)
	}
	if loaded.Len() != 2 {
		t.Fatalf("length %d", loaded.Len())
	}
	path := filepath.Join(t.TempDir(), "activities.yaml")
	if err := loaded.SaveFile(path); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	fileLoaded, err := Load(file)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !equalActivities(fileLoaded.ToSlice(), loaded.ToSlice()) {
		t.Fatalf("file round trip %#v, want %#v", fileLoaded.ToSlice(), loaded.ToSlice())
	}
}

func TestValidationAndMissingRecords(t *testing.T) {
	list := NewList()
	if err := list.Append(Activity{ID: "", Name: "x", Leader: "y"}); err == nil {
		t.Fatal("empty id accepted")
	}
	if err := list.Append(fixtureActivity("C-01", 50)); err != nil {
		t.Fatal(err)
	}
	if err := list.Append(fixtureActivity("C-01", 51)); !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("duplicate error %v", err)
	}
	if err := list.Delete("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing delete error %v", err)
	}
	if _, err := Load(bytes.NewBufferString("activities: [bad")); err == nil {
		t.Fatal("invalid archive accepted")
	}
}

func TestConcurrentAppendsAndQueries(t *testing.T) {
	list := NewList()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, activity := range []Activity{fixtureActivity("C-10", 60), fixtureActivity("C-11", 61)} {
		activity := activity
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- list.Append(activity)
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	if list.Len() != 2 {
		t.Fatalf("length %d", list.Len())
	}
	if _, ok := list.Query("C-10"); !ok {
		t.Fatal("C-10 missing")
	}
	if _, ok := list.Query("C-11"); !ok {
		t.Fatal("C-11 missing")
	}
}

func TestLoadEmptyArchiveThenAppend(t *testing.T) {
	list, err := Load(bytes.NewBufferString("activities: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	activity := fixtureActivity("C-EMPTY", 84)
	if err := list.Append(activity); err != nil {
		t.Fatal(err)
	}
	if got, ok := list.Query(activity.ID); !ok || got != activity {
		t.Fatalf("query returned %#v, %v", got, ok)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func equalActivities(left, right []Activity) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
