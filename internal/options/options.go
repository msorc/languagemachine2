package options

import (
	"errors"
	"fmt"
)

type OptArgs struct {
	options map[string]OptionArgument
	entries map[uint]OptionArgument
	count   uint
}

type OptionArgument interface {
	OptionAction(string, string) (bool, error)
	ConnectOptions(OptArgs, string)
	ExplainOption(int)
	OptionCount() uint
}

func NewOptArgs() *OptArgs {
	return &OptArgs{
		options: make(map[string]OptionArgument),
		entries: make(map[uint]OptionArgument),
	}
}

func (o *OptArgs) Add(k string, x OptionArgument) {
	o.options[k] = x
	x.ConnectOptions(*o, k)
	o.entries[o.count] = x
	o.count++
}

func (o OptArgs) Usage(detail int) {
	for _, x := range o.entries {
		x.ExplainOption(detail)
	}
}

func (o *OptArgs) Arguments(args []string) error {
	for i := 1; i < len(args); i++ {
		k := args[i]
		if k[0] == '-' {
			a, found := o.options[k]
			if found {
				if a.OptionCount() > 0 {
					if i++; i < len(args) {
						_, e := a.OptionAction(k, args[i])
						if e != nil {
							return e
						}
					} else {
						return errors.New("Bad arguments: missing argument for " + k)
					}
				} else {
					_, e := a.OptionAction(k, "")
					if e != nil {
						return e
					}
				}
			} else {
				return errors.New("Bad arguments: unknown option " + k)
			}
		} else if _, found := o.options["..."]; found {
			_, e := o.options["..."].OptionAction("...", k)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

type OptArg struct {
	Options OptArgs
	N       uint   // arguments required
	S       string // short option
	L       string // long option
	X       string // argument help
	U       string // usage help
}

func NewOptArg(count uint, longopt string, arghelp string, help string) *OptArg {
	return &OptArg{N: count, L: longopt, X: arghelp, U: help}
}

func (o OptArg) OptionCount() uint {
	return o.N
}

func (o *OptArg) ConnectOptions(opts OptArgs, k string) {
	o.Options = opts
	opts.options[o.L] = o
	o.S = k
}

func (o OptArg) OptionAction(a string, x string) (noAction bool, err error) {
	fmt.Println("not implemented")
	return
}

func (o OptArg) ExplainOption(detail int) {
	fmt.Printf("%3s %16s %16s %s\n", o.S, o.L, o.X, o.U)
}
