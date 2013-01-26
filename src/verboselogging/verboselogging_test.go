package verboselogging_test

import (
    . "launchpad.net/gocheck"
    "testing"
    VL "verboselogging"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func (ts *TestSuite) TestPostsLoad(c *C) {
    VL.NewRepo("posts")
    c.Succeed()
}

func (ts *TestSuite) TestPagesLoad(c *C) {
    VL.NewRepo("pages")
    c.Succeed()
}
