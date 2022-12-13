// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The embed script wraps file contents with some Go to access it.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	pkg     = flag.String("pkg", "", "Name of package, required")
	vn      = flag.String("var", "", "Name of map variable, required")
	outf    = flag.String("out", "", "Name of output file, required")
	base    = flag.String("base", ".", "Base directory of input files; similar to tar -C mode")
	verbose = flag.Bool("verbose", false, "If set, prints additional log messages")
	gzipit  = flag.Bool("gzip", false, "If set, passes data through gzip compression; checks for existing gzip header first so as not to double-compress")
)

func vlog(format string, v ...interface{}) {
	if !*verbose {
		return
	}
	log.Printf(format, v...)
}

// gzipContent returns a gzipped version of the input, unless the input is already gzipped, in which case
// it returns the input unmodified.
func gzipContent(in []byte, name string, mtime time.Time) []byte {
	// Already gzipped?
	if bytes.Equal(in[:2], []byte{0x1f, 0x8b}) {
		return in
	}
	buf := new(bytes.Buffer)
	// Ignoring error here, because gzip.NewWriterLevel only returns err != nil if the level is invalid,
	// and this level is coming straight from the gzip package.
	gw, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	gw.Header.Name = name
	gw.Header.ModTime = mtime
	// Ignoring errors: Writes to buffers don't error.
	gw.Write(in)
	gw.Close()
	return buf.Bytes()
}

func main() {
	flag.Parse()
	if *pkg == "" || *vn == "" || *outf == "" || len(flag.Args()) == 0 {
		flag.Usage()
		return
	}
	*base = filepath.FromSlash(*base)

	vlog("Creating %s", *outf)
	o, err := os.Create(*outf)
	if err != nil {
		log.Fatalf("Cannot create output file: %v", err)
	}
	defer o.Close()

	w := bufio.NewWriter(o)
	fmt.Fprintf(w, "// This file was generated by embed.go. DO NOT EDIT.\n\npackage %s\n\n", *pkg)
	fmt.Fprintf(w, "var %s = map[string][]byte{\n", *vn)

	for _, ifs := range flag.Args() {
		vlog("Processing arg %s", ifs)
		ms, err := filepath.Glob(filepath.Join(*base, filepath.FromSlash(ifs)))
		if err != nil {
			log.Fatalf("Input pattern invalid: %v", err)
		}
		for _, m := range ms {
			vlog("Processing file %s", m)
			fi, err := os.Stat(m)
			if err != nil {
				vlog("Cannot stat input file, skipping: %v", err)
				continue
			}
			if fi.IsDir() {
				// Skip directories caught in glob.
				continue
			}
			i, err := ioutil.ReadFile(m)
			if err != nil {
				vlog("Cannot read input file, skipping: %v", err)
				continue
			}
			r, err := filepath.Rel(*base, m)
			if err != nil {
				vlog("Cannot compute relative path, skipping: %v", err)
				continue
			}
			if *gzipit {
				name := filepath.Base(m)
				vlog("Applying gzip compression to content of %s", name)
				i = gzipContent(i, name, fi.ModTime())
			}
			fmt.Fprintf(w, "\t%q: []byte(%q),\n", filepath.ToSlash(r), string(i))
		}
	}
	fmt.Fprintln(w, "}")

	vlog("Flushing and closing %s", *outf)
	if err := w.Flush(); err != nil {
		log.Fatalf("Cannot write output file: %v", err)
	}
	if err := o.Close(); err != nil {
		log.Fatalf("Cannot write output file: %v", err)
	}
}
