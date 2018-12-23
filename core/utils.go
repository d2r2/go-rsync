package core

import (
	"math"
	"regexp"
	"strings"

	shell "github.com/d2r2/go-shell"
)

// Round returns the nearest integer, rounding ties away from zero.
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

// func IsKillPending(ctx context.Context) bool {
// 	select {
// 	case <-ctx.Done():
// 		return true
// 	default:
// 		return false
// 	}
// }

func RunExecutableWithExtraVars(path string, env []string, args ...string) (error, int) {
	app := shell.NewApp(path, args...)
	app.AddEnvironments(env)
	ec := app.Run(nil, nil)
	return ec.Error, ec.ExitCode
}

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
