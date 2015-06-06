package email

import (
	"encoding/json"
	"testing"
)

const (
	test_email_config = `
	{
		"type": "email",
		"source_email": "foo@foo.com",
		"dest_emails": [
			"bar@bar.com",
			"baz@baz.com"
		],
		"host": "test.foo.com",
		"port": 25,
		"user": "foo",
		"password": "bar"
	}`
)

func TestParse(t *testing.T) {
	e := &EmailConfig{}
	err := json.Unmarshal([]byte(test_email_config), e)

	if err != nil {
		t.Error(err)
	}
	if e.Password != "bar" {
		t.Logf(e.Password)
		t.Error("Email config not properly parsed")
	}
}
