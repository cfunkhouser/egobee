package egobee

import "testing"

func TestAssembleSelectURL(t *testing.T) {
	testAPIURL := "http://heylookathing"
	testSelection := &Selection{
		SelectionType:  SelectionTypeRegistered,
		SelectionMatch: "awwyiss",
	}

	want := "http://heylookathing?json=%7B%22selection%22%3A%7B%22selectionType%22%3A%22registered%22%2C%22selectionMatch%22%3A%22awwyiss%22%7D%7D"

	got, err := assembleSelectURL(testAPIURL, testSelection)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got != want {
		t.Fatalf("got: %+v, want: %v", got, want)
	}
}
