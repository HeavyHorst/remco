/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package easyKV

import "errors"

// ErrWatchNotSupported is returned if the backend has no watch support and the WatchPrefix method is called
var ErrWatchNotSupported = errors.New("this backend doesn't support watches - use polling instead")

// ErrWatchCanceled is returned is the watcher is canceled
var ErrWatchCanceled = errors.New("watcher error: watcher canceled")
