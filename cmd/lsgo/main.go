package main

import (
	"bufio"
	"flag"
	"fmt"
	"iter"
	"log"
	"maps"
	"os"
	"regexp"
	"strings"

	lg "github.com/takanoriyanagitani/go-list-go-files"
)

//nolint:gochecknoglobals
var noCache lg.Filter = func(fname string) (keep bool) {
	var ignore bool = strings.Contains(fname, "Library/Caches")
	return !ignore
}

func sub() error {
	var useRel bool
	var includeTests bool
	var excludeCache bool
	var skipPattern string
	var keepPattern string

	flag.BoolVar(&useRel, "use-relative-path", false, "use relative path")
	flag.BoolVar(&includeTests, "include-tests", false, "include tests")
	flag.BoolVar(&excludeCache, "exclude-cache", false, "exclude cache")
	flag.StringVar(&skipPattern, "skip-pattern", "", "skip pattern(regex)")
	flag.StringVar(&keepPattern, "keep-pattern", "", "keep pattern(regex)")

	flag.Parse()

	var patterns []string = flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	var cfg lg.Config = lg.Config{}.
		NeedName().
		NeedFiles()

	if includeTests {
		cfg = cfg.NeedTests()
	}

	var lgo lg.ListGo = lg.ListGoDefault.
		WithWriter(os.Stdout).
		WithConfig(cfg).
		WithPatterns(patterns...)

	var fnames iter.Seq2[string, error] = lgo.ToIter()

	if excludeCache {
		fnames = noCache.ToFiltered(fnames)
	}

	if skipPattern != "" {
		re, err := regexp.Compile(skipPattern)
		if nil != err {
			return err
		}
		var skipPat lg.Pattern = lg.Pattern{Regexp: re}
		fnames = skipPat.Skip(fnames)
	}

	if keepPattern != "" {
		re, err := regexp.Compile(keepPattern)
		if nil != err {
			return err
		}
		var keepF lg.Filter = func(fname string) (keep bool) {
			return re.MatchString(fname)
		}
		fnames = keepF.ToFiltered(fnames)
	}

	var mod lg.Modifier = lg.ModifierNop
	if useRel {
		pwd, err := os.Getwd()
		if nil != err {
			return err
		}
		mod = lg.AbsToRelativeNew(pwd)
	}

	var modified iter.Seq2[string, error] = mod.ToModified(fnames)

	nodup, err := lg.IterToSet(modified)
	if nil != err {
		return err
	}

	var keys iter.Seq[string] = maps.Keys(nodup)

	var bwtr *bufio.Writer = bufio.NewWriter(os.Stdout)

	for key := range keys {
		_, err = fmt.Fprintln(bwtr, key)
		if nil != err {
			return err
		}
	}

	return bwtr.Flush()
}

func main() {
	err := sub()
	if nil != err {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
}
