/*
 * This file is part of remco.
 * Based on confd.
 * https://github.com/kelseyhightower/confd/blob/abba746a0cb7c8cb5fe135fa2d884ea3c4a5f666/resource/template/util.go
 * © 2013 Kelsey Hightower
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"path"
)

func appendPrefix(prefix string, keys []string) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = path.Join(prefix, k)
	}
	return s
}
