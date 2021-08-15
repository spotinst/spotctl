package ocean

import (
	"strings"

	oceanv1alpha1 "github.com/spotinst/ocean-operator/api/v1alpha1"
	"github.com/spotinst/spotctl/internal/log"
)

// ComponentsFlag implements pflag.Value interface to store component names,
// allowing transforming and validating them while being set.
type ComponentsFlag struct {
	list map[oceanv1alpha1.OceanComponentName]struct{}
	log  log.Logger
}

// NewEmptyComponentsFlag returns a new ComponentsFlag with an empty list of components.
func NewEmptyComponentsFlag(log log.Logger) *ComponentsFlag {
	return &ComponentsFlag{
		list: make(map[oceanv1alpha1.OceanComponentName]struct{}),
		log:  log,
	}
}

// NewDefaultComponentsFlag returns a new ComponentsFlag with a default list of components.
func NewDefaultComponentsFlag(log log.Logger) *ComponentsFlag {
	f := NewEmptyComponentsFlag(log)
	f.list[oceanv1alpha1.OceanControllerComponentName] = struct{}{}
	f.list[oceanv1alpha1.MetricsServerComponentName] = struct{}{}
	return f
}

func (c *ComponentsFlag) Type() string {
	return "strings"
}

func (c *ComponentsFlag) Set(arg string) error {
	c.list = make(map[oceanv1alpha1.OceanComponentName]struct{})
	v := strings.Split(arg, ",")
	for _, val := range v {
		name := oceanv1alpha1.OceanComponentName(val)
		switch name {
		case oceanv1alpha1.OceanControllerComponentName, oceanv1alpha1.MetricsServerComponentName:
			c.list[name] = struct{}{}
		default:
			if name != "" && c.log != nil {
				c.log.Infof("unknown component name input, ignoring: %q", name)
			}
		}
	}
	return nil
}

func (c *ComponentsFlag) String() string {
	s := make([]string, 0, len(c.list))
	for n := range c.list {
		s = append(s, string(n))
	}
	return strings.Join(s, ",")
}

func (c *ComponentsFlag) StringSlice() []string {
	return strings.Split(c.String(), ",")
}

func (c *ComponentsFlag) List() []oceanv1alpha1.OceanComponentName {
	s := make([]oceanv1alpha1.OceanComponentName, 0, len(c.list))
	for n := range c.list {
		s = append(s, n)
	}
	return s
}
