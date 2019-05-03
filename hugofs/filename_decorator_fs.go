// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugofs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// TODO(bep) mod unexport the New* not used on the outside. Also possibly *FilenameDecoratorFs.
func decorateDirs(fs afero.Fs, meta FileMeta) *FilenameDecoratorFs {
	ffs := &FilenameDecoratorFs{Fs: fs}

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		if !fi.IsDir() {
			// Leave regular files as they are.
			return fi, nil
		}

		inOpener := getFileOpener(fi)

		opener := func() (afero.File, error) {
			return ffs.open(inOpener, name)
		}

		return decorateFileInfo("dird", fs, opener, fi, "", "", meta), nil
	}

	ffs.decorate = decorator

	return ffs

}

func NewPathDecorator(fs afero.Fs, base string) *FilenameDecoratorFs {
	if base == "" || base == "." {
		panic("invalid base")
	}

	ffs := &FilenameDecoratorFs{Fs: fs}

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		path := strings.TrimPrefix(name, base)
		path = strings.TrimLeft(path, string(os.PathSeparator))
		inOpener := getFileOpener(fi)
		opener := func() (afero.File, error) {
			return ffs.open(inOpener, name)
		}
		return decorateFileInfo("pathd", fs, opener, fi, "", path, nil), nil
	}

	ffs.decorate = decorator

	return ffs

}

// NewBasePathPathDecorator returns a new Fs that adds path information
// to the os.FileInfo provided by base.
func NewBasePathPathDecorator(base *afero.BasePathFs) *FilenameDecoratorFs {
	basePath, _ := base.RealPath("")
	if !strings.HasSuffix(basePath, filepathSeparator) {
		basePath += filepathSeparator
	}

	ffs := &FilenameDecoratorFs{Fs: base}

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		path := strings.TrimPrefix(name, basePath)
		inOpener := getFileOpener(fi)
		opener := func() (afero.File, error) {
			return ffs.open(inOpener, name)
		}

		return decorateFileInfo("bpd", base, opener, fi, "", path, nil), nil
	}

	ffs.decorate = decorator

	return ffs
}

// NewFilenameDecorator decorates the given Fs to provide the real filename
// and an Opener func.
func NewFilenameDecorator(fs afero.Fs) *FilenameDecoratorFs {

	ffs := &FilenameDecoratorFs{Fs: fs}

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		inOpener := getFileOpener(fi)
		opener := func() (afero.File, error) {
			return ffs.open(inOpener, name)
		}

		return decorateFileInfo("fnd", fs, opener, fi, name, "", nil), nil
	}

	ffs.decorate = decorator
	return ffs
}

// NewCompositeDirDecorator decorates the given filesystem to make sure
// that directories is always opened by that filesystem.
// TODO(bep) mod check if needed
func NewCompositeDirDecorator(fs afero.Fs) *FilenameDecoratorFs {

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		if !fi.IsDir() {
			return fi, nil
		}
		opener := func() (afero.File, error) {
			return fs.Open(name)
		}
		return decorateFileInfo("compd", fs, opener, fi, "", "", nil), nil
	}

	return &FilenameDecoratorFs{Fs: fs, decorate: decorator}
}

// BasePathRealFilenameFs is a thin wrapper around afero.BasePathFs that
// provides the real filename in Stat and LstatIfPossible.
type FilenameDecoratorFs struct {
	afero.Fs
	decorate func(fi os.FileInfo, filename string) (os.FileInfo, error)
}

// Stat adds path information to os.FileInfo returned from the containing
// BasePathFs.
func (fs *FilenameDecoratorFs) Stat(name string) (os.FileInfo, error) {
	fi, err := fs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}

	return fs.decorate(fi, name)

}

// LstatIfPossible returns the os.FileInfo structure describing a given file.
// It attempts to use Lstat if supported or defers to the os.  In addition to
// the FileInfo, a boolean is returned telling whether Lstat was called.
func (b *FilenameDecoratorFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	var (
		fi  os.FileInfo
		err error
		ok  bool
	)

	if lstater, isLstater := b.Fs.(afero.Lstater); isLstater {
		fi, ok, err = lstater.LstatIfPossible(name)
	} else {
		fi, err = b.Fs.Stat(name)
	}

	if err != nil {
		return nil, false, err
	}

	fi, err = b.decorate(fi, name)

	return fi, ok, err
}

// Open opens the named file for reading.
func (fs *FilenameDecoratorFs) Open(name string) (afero.File, error) {
	return fs.open(nil, name)
}

var counter int

func (fs *FilenameDecoratorFs) open(opener func() (afero.File, error), name string) (afero.File, error) {
	counter++
	// TODO(bep) mod
	if counter > 1000000 {
		panic("bail")
	}
	var (
		f   afero.File
		err error
	)

	if opener != nil {
		f, err = opener()
	} else {
		f, err = fs.Fs.Open(name)
	}

	if err != nil {
		return nil, err
	}

	return &filenameDecoratorFile{File: f, fs: fs}, nil
}

type filenameDecoratorFile struct {
	afero.File
	fs *FilenameDecoratorFs
}

// Readdir adds path information to the ps.FileInfo slice returned from
func (l *filenameDecoratorFile) Readdir(c int) (ofi []os.FileInfo, err error) {
	fis, err := l.File.Readdir(c)
	if err != nil {
		return nil, err
	}

	fisp := make([]os.FileInfo, len(fis))

	for i, fi := range fis {
		filename := filepath.Join(l.Name(), fi.Name())
		fi, err = l.fs.decorate(fi, filename)
		if err != nil {
			return nil, err
		}
		fisp[i] = fi
	}

	return fisp, err
}
