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

func (ts *TestSuite) TestTimeZones(c *C) {
    repo := VL.NewRepo("posts")
    posts, _ := repo.FindLatest(repo.Len())
    for _, post := range posts {
        _, offset := post.PublishedOn.Zone()
        c.Assert(offset, Not(Equals), 0)
    }
}
