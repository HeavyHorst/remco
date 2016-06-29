package template

import (
	"os"

	. "gopkg.in/check.v1"
)

type FuncSuite struct{}

var _ = Suite(&FuncSuite{})

func (s *FuncSuite) Testgetenv(t *C) {
	// existing env var - no default value
	os.Setenv("remco_test", "remco_test")
	res := getenv("remco_test")
	t.Check(res, Equals, "remco_test")

	//existing env var - with default
	res = getenv("remco_test", "test")
	t.Check(res, Equals, "remco_test")

	//non existing env var - no default
	res = getenv("non_existing_123")
	t.Check(res, Equals, "")

	//non existing env var - with default
	res = getenv("non_existing_123", "default")
	t.Check(res, Equals, "default")
}
