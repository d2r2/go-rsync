package core

import (
	"math"
	"regexp"
	"strings"

	shell "github.com/d2r2/go-shell"
)

// Round returns the nearest integer, rounding ties away from zero.
// This functionality is for "before Go 1.10" period, because
// math.Round() was added only since Go 1.10.
func Round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func SplitByEOL(text string) []string {
	return strings.Split(strings.Replace(text, "\r\n", "\n", 0), "\n")
}

// RunExecutableWithExtraVars execute external process returning exit code either any
// error which might happens during start up or execution phases.
func RunExecutableWithExtraVars(pathToApp string, env []string, args ...string) (error, int) {
	app := shell.NewApp(pathToApp, args...)
	app.AddEnvironments(env)
	ec := app.Run(nil, nil)
	return ec.Error, ec.ExitCode
}

// FindStringSubmatchIndexes simplify named Regexp subexpressions extraction via map interface.
// Each entry return 2-byte array with start/end indexes of occurrence.
func FindStringSubmatchIndexes(re *regexp.Regexp, s string) map[string][2]int {
	captures := make(map[string][2]int)
	ind := re.FindStringSubmatchIndex(s)
	names := re.SubexpNames()
	for i, name := range names {
		if name != "" && i < len(ind)/2 {
			if ind[i*2] != -1 && ind[i*2+1] != -1 {
				captures[name] = [2]int{ind[i*2], ind[i*2+1]}
			}
		}
	}
	return captures
}
