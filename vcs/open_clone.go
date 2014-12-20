package vcs

import "fmt"

type initFn func(dir string, isBare bool) (Repository, error)

var initers = map[string]initFn{}

func RegisterIniter(vcs string, f initFn) {
	panicParamsMissing(vcs, f)
	initers[vcs] = f
}

// Init initializes a Repository at the given directory.
func Init(vcs, dir string, isBare bool) (Repository, error) {
	init, ok := initers[vcs]
	if !ok {
		return nil, &UnsupportedVCSError{vcs, "Init"}
	}
	return init(dir, isBare)
}

// An Opener is a function that opens a repository rooted at dir in
// the filesystem.
type Opener func(dir string) (Repository, error)

// openers maps from a VCS type ("git", "hg", etc.) to its opener
// func.
var openers = map[string]Opener{}

// RegisterOpener registers a func to open VCS repositories of a
// specific type.
//
// Library users should import the VCS implementation packages (using
// underscore-import if necessary) to register their openers.
// Afterwards, they may call Open to open repositories of that type.
//
// An implementation of a VCS should call RegisterOpener in an init
// func.
//
// It is not safe in general to call RegisterOpener concurrently or at
// any time after init. Its internal storage is not protected by a
// mutex.
//
// If an opener for the VCS is already registered, RegisterOpener
// overwrites it with f. If vcs is empty or f is nil, it also panics.
func RegisterOpener(vcs string, f Opener) {
	panicParamsMissing(vcs, f)
	openers[vcs] = f
}

// Open opens a repository rooted at dir. An opener for its VCS must
// be registered (typically by importing a subpackage of go-vcs that
// calls RegisterOpener, using underscore-import if necessary).
func Open(vcs, dir string) (Repository, error) {
	opener, present := openers[vcs]
	if !present {
		return nil, &UnsupportedVCSError{vcs, "Open"}
	}
	return opener(dir)
}

// A cloner is a function that clones a repository from a URL to dir
// in the filesystem.
type Cloner func(url, dir string, opt CloneOpt) (Repository, error)

// cloners maps from a VCS type ("git", "hg", etc.) to its cloner
// func.
var cloners = map[string]Cloner{}

// RegisterCloner registers a func to clone VCS repositories of a
// specific type.
//
// Library users should import the VCS implementation packages (using
// underscore-import if necessary) to register their cloners.
// Afterwards, they may call Clone to clone repositories of that type.
//
// An implementation of a VCS should call RegisterCloner in an init
// func.
//
// It is not safe in general to call RegisterCloner concurrently or at
// any time after init. Its internal storage is not protected by a
// mutex.
//
// If a cloner for the VCS is already registered, RegisterCloner
// overwrites it with f. If vcs is empty or f is nil, it also panics.
func RegisterCloner(vcs string, f Cloner) {
	panicParamsMissing(vcs, f)
	cloners[vcs] = f
}

func panicParamsMissing(vcs string, t interface{}) {
	if vcs == "" {
		panic("empty VCS type")
	}
	if t == nil {
		panic("Cloner func for '" + vcs + "' is nil")
	}
}

// CloneOpt configures a clone operation.
type CloneOpt struct {
	Bare   bool // create a bare repo
	Mirror bool // create a mirror repo (`git clone --mirror`)

	RemoteOpts // configures communication with the remote repository

	// TODO(sqs): these options are fairly
	// VCS-implementation-specific. What's a better way of doing this?
}

// Clone clones a repository rooted at dir. A cloner for its VCS must
// be registered (typically by importing a subpackage of go-vcs that
// calls RegisterCloner, using underscore-import if necessary).
func Clone(vcs, url, dir string, opt CloneOpt) (Repository, error) {
	cloner, present := cloners[vcs]
	if !present {
		return nil, &UnsupportedVCSError{vcs, "Clone"}
	}
	return cloner(url, dir, opt)
}

// UnsupportedVCSError is when Open is called to open a repository of
// a VCS type that doesn't have an Opener registered.
type UnsupportedVCSError struct {
	VCS string // the VCS type that was attempted to be used
	op  string // the operation (Open, etc.) that was attempted
}

func (e *UnsupportedVCSError) Error() string {
	return fmt.Sprintf("unsupported VCS for %s: %s", e.op, e.VCS)
}
