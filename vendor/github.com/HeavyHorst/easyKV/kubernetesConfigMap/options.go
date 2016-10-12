/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package kubernetesConfigMap

// Options contains all values that are needed to connect to kubernetes
type Options struct {
	Config string
}

// Option configures the kubernetes client
type Option func(*Options)

// WithConfig configures the kubernetes client with the given kubeconfig.
// For example ~/.kube/config
// This is only needed if remco is running outside of an kubernetes pod.
func WithConfig(c string) Option {
	return func(o *Options) {
		o.Config = c
	}
}
