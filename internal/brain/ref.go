package brain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Errors resolving an engram argument.
var (
	ErrNoEngram    = errors.New("no such engram")
	ErrAmbiguous   = errors.New("in more than one brain")
	ErrCrossBrain  = errors.New("in different brains")
	ErrNotInBrain  = errors.New("not in a brain")
	ErrNotAnEngram = errors.New("not a top-level .md file")
)

// Ref is one engram: a top-level .md file of a brain.
type Ref struct {
	Brain Brain
	File  string
}

// Path returns the engram's absolute path.
func (r Ref) Path() string {
	return filepath.Join(r.Brain.Location, r.File)
}

// Resolve reads an engram argument: a path when it holds a "/", else a name
// looked up in brainName, or in every brain when brainName is "".
func (r *Registry) Resolve(arg, brainName string) (Ref, error) {
	if isPath(arg) {
		return r.resolvePath(arg, brainName)
	}
	file := fileName(arg)
	if brainName != "" {
		b, err := r.Get(brainName)
		if err != nil {
			return Ref{}, err
		}
		if !holds(b, file) {
			return Ref{}, fmt.Errorf("(%s) %s: %w", b.Name, file, ErrNoEngram)
		}
		return Ref{Brain: b, File: file}, nil
	}
	b, err := one(file, r.holders(file))
	return Ref{Brain: b, File: file}, err
}

// ResolvePair reads two engram arguments that must share a brain.
func (r *Registry) ResolvePair(a, b, brainName string) (Ref, Ref, error) {
	if brainName == "" {
		var err error
		if brainName, err = r.pairBrain(a, b); err != nil {
			return Ref{}, Ref{}, err
		}
	}
	ra, err := r.Resolve(a, brainName)
	if err != nil {
		return Ref{}, Ref{}, err
	}
	rb, err := r.Resolve(b, brainName)
	if err != nil {
		return Ref{}, Ref{}, err
	}
	return ra, rb, nil
}

func (r *Registry) pairBrain(a, b string) (string, error) {
	switch {
	case isPath(a) && isPath(b):
		ra, err := r.Resolve(a, "")
		if err != nil {
			return "", err
		}
		rb, err := r.Resolve(b, "")
		if err != nil {
			return "", err
		}
		if ra.Brain.Name != rb.Brain.Name {
			return "", fmt.Errorf("%s and %s: %w: %s, %s", a, b, ErrCrossBrain, ra.Brain.Name, rb.Brain.Name)
		}
		return ra.Brain.Name, nil
	case isPath(a):
		ra, err := r.Resolve(a, "")
		return ra.Brain.Name, err
	case isPath(b):
		rb, err := r.Resolve(b, "")
		return rb.Brain.Name, err
	}
	fa, fb := fileName(a), fileName(b)
	both := slices.DeleteFunc(r.holders(fa), func(x Brain) bool { return !holds(x, fb) })
	shared, err := one(fa+" and "+fb, both)
	return shared.Name, err
}

// In nested brains, the one the file is top-level in wins.
func (r *Registry) resolvePath(arg, brainName string) (Ref, error) {
	abs, err := filepath.Abs(arg)
	if err != nil {
		return Ref{}, err
	}
	brains := r.List()
	if brainName != "" {
		b, err := r.Get(brainName)
		if err != nil {
			return Ref{}, err
		}
		brains = []Brain{b}
	}
	inside := false
	for _, b := range brains {
		rel, ok := Within(b.Location, abs)
		if ok && isEngramFile(rel) {
			return Ref{Brain: b, File: rel}, nil
		}
		inside = inside || ok
	}
	if inside {
		return Ref{}, fmt.Errorf("%s: %w", abs, ErrNotAnEngram)
	}
	if brainName != "" {
		return Ref{}, fmt.Errorf("(%s) %s: %w named by --brain", brainName, abs, ErrNotInBrain)
	}
	return Ref{}, fmt.Errorf("%s: %w; register-brain its folder", abs, ErrNotInBrain)
}

func one(engrams string, brains []Brain) (Brain, error) {
	switch len(brains) {
	case 0:
		return Brain{}, fmt.Errorf("%s: %w", engrams, ErrNoEngram)
	case 1:
		return brains[0], nil
	default:
		return Brain{}, fmt.Errorf("%s: %w: %s", engrams, ErrAmbiguous, names(brains))
	}
}

func (r *Registry) holders(file string) []Brain {
	var out []Brain
	for _, b := range r.List() {
		if holds(b, file) {
			out = append(out, b)
		}
	}
	return out
}

func holds(b Brain, file string) bool {
	info, err := os.Stat(filepath.Join(b.Location, file))
	return err == nil && info.Mode().IsRegular()
}

func names(brains []Brain) string {
	out := make([]string, len(brains))
	for i, b := range brains {
		out[i] = b.Name
	}
	return strings.Join(out, ", ")
}

func isPath(arg string) bool {
	return strings.Contains(arg, "/")
}

func fileName(name string) string {
	return strings.TrimSuffix(name, ".md") + ".md"
}

func isEngramFile(rel string) bool {
	return !strings.ContainsRune(rel, filepath.Separator) && filepath.Ext(rel) == ".md"
}
