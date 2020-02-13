package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/becgabri/enigma"
	"github.com/mkideal/cli"
)

// CLIOpts sets the parameter format for Enigma CLI. It also includes a "help"
// flag and a "condensed" flag telling the program to output plain result.
// Also, this CLI module abuses tags so much it hurts. Oh well. ¯\_(ツ)_/¯
type CLIOpts struct {
	Help      bool `cli:"!h,help" usage:"Show help."`
	Condensed bool `cli:"c,condensed" name:"false" usage:"Output the result without additional information."`

	Rotors    []string `cli:"rotors" name:"\"I II III\"" usage:"Rotor configuration. Supported: I, II, III, IV, V, VI, VII, VIII, Beta, Gamma."`
	Rings     []int    `cli:"rings" name:"\"1 1 1\"" usage:"Rotor rings offset: from 1 (default) to 26 for each rotor."`
	Position  []string `cli:"position" name:"\"A A A\"" usage:"Starting position of the rotors: from A (default) to Z for each."`
	Plugboard []string `cli:"plugboard" name:"\"AB CD\"" usage:"Optional plugboard pairs to scramble the message further."`

	Reflector string `cli:"reflector" name:"C" usage:"Reflector. Supported: A, B, C, B-Thin, C-Thin."`
}

// CLIDefaults is used to populate default values in case
// one or more of the parameters aren't set. It is assumed
// that rotor rings and positions will be the same for all
// rotors if not set explicitly, so only one value is stored.
var CLIDefaults = struct {
	Reflector string
	Ring      []int
	Position  []string
	Rotors    []string
}{
	Reflector: "C-thin",
	Ring:      []int{1,1,1,16},
	Position:  []string{"B","C","B","Q"},
	Rotors:    []string{"I","II","IV","III"},
}

// SetDefaults sets values for all Enigma parameters that
// were not set explicitly.
// Plugboard is the only parameter that does not require a
// default, since it may not be set, and in some Enigma versions
// there was no plugboard at all.
func SetDefaults(argv *CLIOpts) {
	if argv.Reflector == "" {
		argv.Reflector = CLIDefaults.Reflector
	}
	if len(argv.Rotors) == 0 {
		argv.Rotors = CLIDefaults.Rotors
	}
	loadRings := (len(argv.Rings) == 0)
	loadPosition := (len(argv.Position) == 0)
	if loadRings {
        argv.Rings = CLIDefaults.Ring
    }
    if loadPosition {
		argv.Position = CLIDefaults.Position

	}
}
func IoCCal(mes string) float64 {
	var stat [27] int
	var i, asc, l int
	_=i
	var IoC float64
	l = len(mes)
	IoC = 0
	for i := 0; i < l; i++ {
		asc = int(mes[i]) - 65
		stat[asc]++
	}
	for i := 0; i <= 26; i++ {
		IoC = IoC + (float64(stat[i])*float64(stat[i]-1)/float64(l)/float64(l-1))
	}
	return IoC
}
func main() {

	cli.SetUsageStyle(cli.DenseManualStyle)
	cli.Run(new(CLIOpts), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*CLIOpts)
		originalPlaintext := strings.Join(ctx.Args(), " ")
		plaintext := enigma.SanitizePlaintext(originalPlaintext)

		if argv.Help || len(plaintext) == 0 {
			com := ctx.Command()
			com.Text = DescriptionTemplate
			ctx.String(com.Usage(ctx))
			return nil
		}
		var rotors = []string{}
		rotors = []string{"Gamma", "VI", "IV", "III"}

		var plugboard = []string{}
		var position = []string{}
		position = []string{"D", "A", "B", "Q"}
		var hillclimb float64
		var letter = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
		var plug string
		hillclimb = 0.0
		argv.Rotors = rotors
		argv.Position = position
		for p := 0; p < 10; p++ {
			for m := 0; m < len(letter); m++ {
				for n := 0; n < len(letter); n++ {
					argv.Plugboard = plugboard
					config := make([]enigma.RotorConfig, 4)
					for index, rotor := range argv.Rotors {
						ring := argv.Rings[index]
						v := argv.Position[index][0]
						config[index] = enigma.RotorConfig{ID: rotor, Start: v, Ring: ring}
					}
					e := enigma.NewEnigma(config, argv.Reflector, argv.Plugboard)
					encode := e.EncodeString(plaintext)
					if argv.Condensed {
						fmt.Print(encode)
						return nil
					}
					if IoCCal(encode) > hillclimb {
						hillclimb = IoCCal(encode)
						plug = letter[m] + letter[n]
						fmt.Print(hillclimb)
						if hillclimb > 0.05 {
							tmpl, _ := template.New("cli").Parse(OutputTemplate)
							err := tmpl.Execute(os.Stdout, struct {
								Original, Plain, Encoded string
								Args                     *CLIOpts
								Ctx                      *cli.Context
							}{originalPlaintext, plaintext, encode, argv, ctx})
							fmt.Print(err)
						}
					}
				}
				for m := 0; m < len(letter); m++ {
					if (strings.Contains(plug, letter[m])) {
						letter = append(letter[:m], letter[m+1:]...)
						m--
					}
				}
				plugboard = append(plugboard, plug)
			}
		}
		return nil
	})
}