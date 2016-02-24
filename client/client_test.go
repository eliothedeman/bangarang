package client

import (
	"fmt"
	"testing"

	c "github.com/glycerine/goconvey/convey"
)

func TestFmtHost(t *testing.T) {
	host := "10.0.0.1"
	port := 8080
	c.Convey("Given a host "+host+" and port "+fmt.Sprint(port), t, func() {
		c.ShouldEqual(fmtHost(host, port), "10.0.0.1:8080")
	})
}
