/*
 * This file is part of easyKV.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package easyKV

// WatchOptions represents options for watch operations
type WatchOptions struct {
	WaitIndex uint64
	Keys      []string
}

// WatchOption configures the WatchPrefix operation
type WatchOption func(*WatchOptions)

// WithKeys reduces the scope of keys that can trigger updates to keys (not an exact match)
func WithKeys(keys []string) WatchOption {
	return func(o *WatchOptions) {
		o.Keys = keys
	}
}

// WithWaitIndex sets the WaitIndex of the watcher
func WithWaitIndex(waitIndex uint64) WatchOption {
	return func(o *WatchOptions) {
		o.WaitIndex = waitIndex
	}
}

// A ReadWatcher - can get values and watch a prefix for changes
type ReadWatcher interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, stopChan chan bool, opts ...WatchOption) (uint64, error)
}

// TODO - create more interfaces - ReadWriter?
