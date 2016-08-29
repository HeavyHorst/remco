/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// values set with linker flags
// don't you dare modifying this values!
var version string
var buildDate string
var commit string

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("remco Version: " + version)
		fmt.Println("UTC Build Time: " + buildDate)
		fmt.Println("Git Commit Hash: " + commit)
		fmt.Println("Go Version: " + runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
