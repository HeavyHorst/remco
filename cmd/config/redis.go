/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"github.com/HeavyHorst/remco/backends/redis"
	"github.com/spf13/cobra"
)

var redisConfig = &redis.Config{}

// RedisCmd represents the redis command
var RedisCmd = &cobra.Command{
	Use:   "redis",
	Short: "load a config file from redis",
	Run:   defaultConfigRunCMD(redisConfig),
}

func init() {
	RedisCmd.Flags().StringSliceVar(&redisConfig.Nodes, "nodes", []string{"127.0.0.1:6379"}, "The redis nodes")
	RedisCmd.Flags().StringVar(&redisConfig.Password, "password", "", "The redis password")
	RedisCmd.Flags().StringP("config", "c", "", "The path in redis where the config is stored")

	CfgCmd.AddCommand(RedisCmd)
}
