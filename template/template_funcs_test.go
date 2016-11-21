/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"net"
	"os"

	. "gopkg.in/check.v1"
)

type FunctionTestSuite struct{}

var _ = Suite(&FunctionTestSuite{})

func (s *FunctionTestSuite) TestAddFuncs(t *C) {
	in := map[string]interface{}{
		"a": "hallo",
		"b": "hello",
	}
	out := make(map[string]interface{})

	addFuncs(out, in)

	t.Check(len(out), Equals, len(in))
}

func (s *FunctionTestSuite) TestLookupIP(t *C) {
	ips := lookupIP("localhost")
	if len(ips) > 0 {
		t.Check(ips[0], Equals, "127.0.0.1")
	} else {
		t.Error("lookupIP failed")
	}
}

func (s *FunctionTestSuite) TestLookupSRV(t *C) {
	expected := []*net.SRV{
		&net.SRV{
			Target:   "alt1.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		&net.SRV{
			Target:   "alt2.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		&net.SRV{
			Target:   "alt3.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		&net.SRV{
			Target:   "alt4.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		&net.SRV{
			Target:   "xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 5,
			Weight:   0,
		},
	}

	srv := lookupSRV("xmpp-server", "tcp", "google.com")
	t.Check(srv, DeepEquals, expected)
}

func (s *FunctionTestSuite) TestGetEnv(t *C) {
	key := "coolEnvVar"
	expected := "mmmh lecker saure Gurken!"
	err := os.Setenv(key, expected)
	if err != nil {
		t.Error(err)
	}

	t.Check(getenv(key), Equals, expected)
}

func (s *FunctionTestSuite) TestGetEnvDefault(t *C) {
	key := "ihopethisenvvardontexists"
	expected := "default"

	t.Check(getenv(key, "default"), Equals, expected)
}
