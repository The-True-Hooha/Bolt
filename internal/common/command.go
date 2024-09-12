package common

import (
	"github.com/spf13/pflag"
)


type Command struct {
	Name string
	Description string
	Flags *pflag.FlagSet
	Execute func(args []string) error
}