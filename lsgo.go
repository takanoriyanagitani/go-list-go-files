package lsgo

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"path/filepath"
	"regexp"

	gp "golang.org/x/tools/go/packages"
)

type Pattern struct{ *regexp.Regexp }

func (p Pattern) Skip(
	original iter.Seq2[string, error],
) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for fname, err := range original {
			if nil != err {
				yield(fname, err)
				return
			}

			var skip bool = p.Regexp.MatchString(fname)
			if skip {
				continue
			}

			if !yield(fname, nil) {
				return
			}
		}
	}
}

type Config gp.Config

func (c Config) NeedFiles() Config {
	c.Mode |= gp.NeedFiles
	return c
}

func (c Config) NeedName() Config {
	c.Mode |= gp.NeedName
	return c
}

func (c Config) NeedTests() Config {
	c.Tests = true
	return c
}

func (c Config) Load(patterns ...string) ([]*gp.Package, error) {
	var cfg gp.Config = gp.Config(c)
	return gp.Load(&cfg, patterns...)
}

type Package struct{ *gp.Package }

func (p Package) Files() []string { return p.Package.GoFiles }

func (p Package) FilesToWriter(wtr io.Writer) error {
	for _, file := range p.Files() {
		_, err := fmt.Fprintln(wtr, file)
		if nil != err {
			return err
		}
	}
	return nil
}

type Modifier func(original string) (modified string, err error)

func (m Modifier) ToModified(
	original iter.Seq2[string, error],
) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for fname, err := range original {
			if nil != err {
				yield(fname, err)
				return
			}

			modified, err := m(fname)
			if !yield(modified, err) {
				return
			}
		}
	}
}

func ModifierNop(s string) (string, error) { return s, nil }

//nolint:gochecknoglobals
var ModifierDefault Modifier = ModifierNop

func AbsToRelativeNew(base string) Modifier {
	return func(original string) (modified string, err error) {
		return filepath.Rel(base, original)
	}
}

type ListGo struct {
	Config
	io.Writer

	Patterns []string
}

func (l ListGo) ListGoFiles() error {
	pkgs, err := l.Config.Load(l.Patterns...)
	if nil != err {
		return err
	}

	var bwtr *bufio.Writer = bufio.NewWriter(l.Writer)

	for _, p := range pkgs {
		pkg := Package{Package: p}
		err = pkg.FilesToWriter(bwtr)
		if nil != err {
			return err
		}
	}

	return bwtr.Flush()
}

func (l ListGo) WithWriter(wtr io.Writer) ListGo {
	l.Writer = wtr
	return l
}

func (l ListGo) WithConfig(cfg Config) ListGo {
	l.Config = cfg
	return l
}

func (l ListGo) WithPatterns(patterns ...string) ListGo {
	l.Patterns = patterns
	return l
}

func (l ListGo) ToIter() iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pkgs, err := l.Config.Load(l.Patterns...)
		if nil != err {
			yield("", err)
			return
		}

		for _, p := range pkgs {
			pkg := Package{p}
			var files []string = pkg.Files()
			for _, file := range files {
				if !yield(file, nil) {
					return
				}
			}
		}
	}
}

func IterToSet(files iter.Seq2[string, error]) (map[string]struct{}, error) {
	ret := map[string]struct{}{}
	for fname, err := range files {
		if nil != err {
			return nil, err
		}

		ret[fname] = struct{}{}
	}
	return ret, nil
}

type Filter func(filename string) (keep bool)

func (f Filter) ToFiltered(original iter.Seq2[string, error]) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		for fname, err := range original {
			if nil != err {
				yield(fname, err)
				return
			}

			var keep bool = f(fname)
			var skip bool = !keep
			if skip {
				continue
			}

			if !yield(fname, nil) {
				return
			}
		}
	}
}

func FilterNop(_ string) (keep bool) { return true }

//nolint:gochecknoglobals
var FilterDefault Filter = FilterNop

//nolint:gochecknoglobals
var ListGoDefault ListGo = ListGo{}.
	WithWriter(io.Discard).
	WithPatterns("./...").
	WithConfig(
		Config{}.
			NeedName().
			NeedFiles().
			NeedTests(),
	)
