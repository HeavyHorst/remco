/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcd

import (
	"fmt"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestNew(t *C) {
	ba := BasicAuthOptions{
		Username: "",
		Password: "",
	}

	tls := TLSOptions{
		ClientCert:   "",
		ClientKey:    "",
		ClientCaKeys: "",
	}

	c, err := New([]string{"http://127.0.0.1:2379"}, WithVersion(2), WithBasicAuth(ba), WithTLSOptions(tls))
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	t.Check(fmt.Sprintf("%T", c), Equals, "*etcdv2.Client")

	c2, err := New([]string{"http://127.0.0.1:2379"}, WithVersion(3), WithBasicAuth(ba), WithTLSOptions(tls))
	if err != nil {
		t.Error(err)
	}
	defer c2.Close()

	t.Check(fmt.Sprintf("%T", c2), Equals, "*etcdv3.Client")
}
