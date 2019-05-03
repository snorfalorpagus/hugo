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
	"strings"
	"time"

	"github.com/spf13/afero"
)

func restrictToBase(fs afero.Fs, base string) afero.Fs {
	return &baseRestrictorFs{
		Fs:   fs,
		base: base,
	}
}

type baseRestrictorFs struct {
	afero.Fs
	base string
}

func (fs *baseRestrictorFs) belongs(name string) bool {
	return strings.HasPrefix(name, fs.base)
}

func (fs *baseRestrictorFs) Create(name string) (afero.File, error) {
	if !fs.belongs(name) {
		return nil, os.ErrNotExist
	}
	return fs.Fs.Create(name)
}

func (fs *baseRestrictorFs) Mkdir(name string, perm os.FileMode) error {
	if !fs.belongs(name) {
		return os.ErrNotExist
	}
	return fs.Fs.Mkdir(name, perm)
}

func (fs *baseRestrictorFs) MkdirAll(path string, perm os.FileMode) error {
	if !fs.belongs(path) {
		return os.ErrNotExist
	}
	return fs.Fs.MkdirAll(path, perm)
}

func (fs *baseRestrictorFs) Open(name string) (afero.File, error) {
	if !fs.belongs(name) {
		return nil, os.ErrNotExist
	}
	return fs.Fs.Open(name)
}

func (fs *baseRestrictorFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if !fs.belongs(name) {
		return nil, os.ErrNotExist
	}
	return fs.Fs.OpenFile(name, flag, perm)
}

func (fs *baseRestrictorFs) Remove(name string) error {
	if !fs.belongs(name) {
		return os.ErrNotExist
	}
	return fs.Fs.Remove(name)
}

func (fs *baseRestrictorFs) RemoveAll(path string) error {
	if !fs.belongs(path) {
		return os.ErrNotExist
	}
	return fs.Fs.RemoveAll(path)
}

func (fs *baseRestrictorFs) Rename(oldname string, newname string) error {
	if !fs.belongs(oldname) {
		return os.ErrNotExist
	}
	return fs.Fs.Rename(oldname, newname)
}

func (fs *baseRestrictorFs) Stat(name string) (os.FileInfo, error) {
	if !fs.belongs(name) {
		return nil, os.ErrNotExist
	}
	return fs.Fs.Stat(name)
}

func (fs *baseRestrictorFs) Name() string {
	return "baseRestrictorFs"
}

func (fs *baseRestrictorFs) Chmod(name string, mode os.FileMode) error {
	if !fs.belongs(name) {
		return os.ErrNotExist
	}
	return fs.Fs.Chmod(name, mode)
}

func (fs *baseRestrictorFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if !fs.belongs(name) {
		return os.ErrNotExist
	}
	return fs.Fs.Chtimes(name, atime, mtime)
}
