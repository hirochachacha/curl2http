package curl2http

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/markbates/pkger"
)

type Flag struct {
	ShortName   string
	LongName    string
	Description string
	Parameter   string

	IsSet  bool
	Values []string
}

func (flg *Flag) Value(i int) string {
	if i < 0 {
		i += len(flg.Values)
	}
	if i < 0 || i >= len(flg.Values) {
		return ""
	}
	return flg.Values[i]
}

type FlagSet struct {
	sflags map[string]*Flag
	lflags map[string]*Flag
	args   []string
	done   bool
}

func NewFlagSet() *FlagSet {
	f, err := pkger.Open("/assets/help.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sflags := make(map[string]*Flag)
	lflags := make(map[string]*Flag)

	sc := bufio.NewScanner(f)

	if sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "Usage:") {
			panic("unexpected help format")
		}

		for sc.Scan() {
			line = strings.TrimSpace(sc.Text())
			if !strings.HasPrefix(line, "-") {
				panic("unexpected help format")
			}

			flg := new(Flag)

			for strings.HasPrefix(line, "-") {
				if strings.HasPrefix(line, "--") {
					i := 2
					j := strings.IndexAny(line, ", ")
					if j == -1 {
						panic("unexpected help format")
					}

					flg.LongName = line[i:j]

					line = strings.TrimSpace(line[j+1:])
				} else {
					i := 1
					j := strings.IndexAny(line, ", ")
					if j == -1 {
						panic("unexpected help format")
					}

					flg.ShortName = line[i:j]

					line = strings.TrimSpace(line[j+1:])
				}
			}

			var params []string

			for strings.HasPrefix(line, "<") {
				i := 1
				j := strings.IndexRune(line, '>')

				params = append(params, line[i:j])

				line = strings.TrimSpace(line[j+1:])
			}

			switch len(params) {
			case 0:
			case 1:
				flg.Parameter = params[0]
			default:
				panic("unsupported help format")
			}

			flg.Description = line

			if flg.ShortName != "" {
				sflags[flg.ShortName] = flg
			}
			if flg.LongName != "" {
				lflags[flg.LongName] = flg
			}
		}
	}

	if err := sc.Err(); err != nil {
		panic(err)
	}

	return &FlagSet{
		sflags: sflags,
		lflags: lflags,
	}
}

func (fs *FlagSet) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			var flg *Flag
			var ok bool
			if strings.HasPrefix(arg, "--") {
				flg, ok = fs.lflags[arg[2:]]
			} else {
				flg, ok = fs.sflags[arg[1:]]
			}
			if !ok {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			flg.IsSet = true
			if flg.Parameter != "" {
				if len(args) < i+2 {
					return fmt.Errorf("insufficient argument: %s", arg)
				}
				val := args[i+1]
				if strings.HasPrefix(val, "-") {
					return fmt.Errorf("insufficient argument: %s", arg)
				}
				flg.Values = append(flg.Values, val)
				i++
			}
		} else {
			fs.args = append(fs.args, arg)
		}
	}

	fs.done = true

	return nil
}

func (fs *FlagSet) Parsed() bool {
	return fs.done
}

func (fs *FlagSet) ShortFlag(name string) *Flag {
	return fs.sflags[name]
}

func (fs *FlagSet) LongFlag(name string) *Flag {
	return fs.lflags[name]
}

func (fs *FlagSet) Args() []string {
	return fs.args
}

func (fs *FlagSet) Arg(i int) string {
	if i < 0 {
		i += len(fs.args)
	}
	if i < 0 || i >= len(fs.args) {
		return ""
	}
	return fs.args[i]
}

func (fs *FlagSet) VisitAll(fn func(flg *Flag)) {
	for _, flg := range fs.lflags {
		fn(flg)
	}
}

func (fs *FlagSet) Visit(fn func(flg *Flag)) {
	for _, flg := range fs.lflags {
		if flg.IsSet {
			fn(flg)
		}
	}
}
