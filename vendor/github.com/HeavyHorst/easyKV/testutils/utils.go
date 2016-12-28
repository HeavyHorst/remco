/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package testutils

import (
	"context"

	"github.com/HeavyHorst/easyKV"

	. "gopkg.in/check.v1"
)

var expected = map[string]string{
	"/premtest/database/url":              "www.google.de",
	"/premtest/database/user":             "Boris",
	"/remtest/database/hosts/192.168.0.1": "test1",
	"/remtest/database/hosts/192.168.0.2": "test2",
}

var expectedPrefix = map[string]string{
	"/premtest/database/url":  "www.google.de",
	"/premtest/database/user": "Boris",
}

func GetExpected() map[string]string {
	return expected
}

// GetValues is a util function to test the easyKV.ReadWatcher.GetValues Method
func GetValues(t *C, c easyKV.ReadWatcher) error {
	m, err := c.GetValues([]string{"/remtest", "/premtest"})
	if err != nil {
		return err
	}
	t.Check(m, DeepEquals, expected)

	m2, err := c.GetValues([]string{"/premtest"})
	if err != nil {
		return err
	}
	t.Check(m2, DeepEquals, expectedPrefix)
	return nil
}

func WatchPrefix(t *C, c easyKV.ReadWatcher, ctx context.Context, prefix string, keys []string) uint64 {
	n, err := c.WatchPrefix(ctx, prefix, easyKV.WithWaitIndex(0), easyKV.WithKeys(keys))
	if err != nil {
		if err != easyKV.ErrWatchCanceled {
			t.Error(err)
		}
	}
	return n
}

func WatchPrefixError(t *C, c easyKV.ReadWatcher) {
	num, err := c.WatchPrefix(context.Background(), "")
	t.Check(num, Equals, uint64(0))
	t.Check(err, Equals, easyKV.ErrWatchNotSupported)
}
