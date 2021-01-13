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
	ips, err := lookupIP("localhost")
	if err != nil {
		t.Error(err)
	}
	if len(ips) > 0 {
		t.Check(ips[0], Equals, "127.0.0.1")
	} else {
		t.Error("lookupIP failed")
	}
}

func (s *FunctionTestSuite) TestLookupSRV(t *C) {
	expected := []*net.SRV{
		{
			Target:   "alt1.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		{
			Target:   "alt2.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		{
			Target:   "alt3.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		{
			Target:   "alt4.xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 20,
			Weight:   0,
		},
		{
			Target:   "xmpp-server.l.google.com.",
			Port:     5269,
			Priority: 5,
			Weight:   0,
		},
	}

	srv, err := lookupSRV("xmpp-server", "tcp", "google.com")
	if err != nil {
		t.Error(err)
	}
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

func (s *FunctionTestSuite) TestInterfaceSet(t *C) {
	set := createSet()
	set.Append("Hallo")
	set.Append("Hallo")
	set.Append(1)
	set.Append(true)
	set.Append(false)

	t.Check(len(set), Equals, 4)
	t.Check(set.Contains("Hallo"), Equals, true)
	set.Remove("Hallo")
	t.Check(len(set), Equals, 3)
	t.Check(set.Contains("Hallo"), Equals, false)
	t.Check(set.Contains(false), Equals, true)
}

func (s *FunctionTestSuite) TestTemplateMap(t *C) {
	m := createMap()
	m.Set("Hallo", "OneOneOne")
	m.Set("Test", "Snickers")
	m.Set("One", 1)

	t.Check(m.Get("Hallo"), DeepEquals, "OneOneOne")
	t.Check(m.Get("One"), DeepEquals, 1)

	m.Remove("One")
	t.Check(m.Get("One"), DeepEquals, nil)
}
