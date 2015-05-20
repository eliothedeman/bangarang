package alarm

import (
	"encoding/json"
	"testing"

	"github.com/eliothedeman/bangarang/event"
)

const (
	test_policy_with_regex = `
	{
		"match": {
			"host": "test\\.hello"
		}
	}
`
)

func TestPolicyRegexParsing(t *testing.T) {
	p := &Policy{}

	err := json.Unmarshal([]byte(test_policy_with_regex), p)
	if err != nil {
		t.Error(err)
	}

	if p.Match["host"] != `test\.hello` {
		t.Error("regex not properly parsed")
	}

}

func TestMatchTags(t *testing.T) {
	p := &Policy{}
	e := &event.Event{}
	e.Tags = map[string]string{
		"test_tag": "0",
	}

	p.Match = map[string]string{
		"test_tag": "[0-9]+",
	}
	p.Compile()

	if !p.CheckMatch(e) {
		t.Fail()
	}
}

func TestMatchTagsExtra(t *testing.T) {
	p := &Policy{}
	e := &event.Event{}
	e.Tags = map[string]string{
		"test_tag":  "0",
		"extra_tag": "w234",
	}

	p.Match = map[string]string{
		"test_tag": "[0-9]+",
	}
	p.Compile()

	if !p.CheckMatch(e) {
		t.Fail()
	}
}

func TestMatchStructFiled(t *testing.T) {
	p := &Policy{}
	e := &event.Event{}
	e.Host = "my_host"

	p.Match = map[string]string{
		"host": "my.*",
	}
	p.Compile()

	if !p.CheckMatch(e) {
		t.Fail()
	}

	e.Host = ""
	if p.CheckMatch(e) {
		t.Fail()
	}
}
