package adapter

import "testing"

func TestParseTextAction(t *testing.T) {
	action, err := ParseTextAction("Action: click\nAction Input: {\"x\": 1}")
	if err != nil {
		t.Fatal(err)
	}
	if action.Name != "click" {
		t.Errorf("name = %q, want click", action.Name)
	}
	if action.Arguments["input"] != `{"x": 1}` {
		t.Errorf("arguments = %v", action.Arguments)
	}
}

func TestParseTextActionRejectsAResponseWithoutAnAction(t *testing.T) {
	if _, err := ParseTextAction("no action here"); err == nil {
		t.Error("expected a parse failure")
	}
}

func TestParseTextActionRejectsAnEmptyResponse(t *testing.T) {
	if _, err := ParseTextAction("  "); err == nil {
		t.Error("expected an empty response to be a parse failure")
	}
}
