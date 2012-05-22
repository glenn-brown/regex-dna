/*
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

    * Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.

    * Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.

    * Neither the name of "The Computer Language Benchmarks Game" nor the
    name of "The Computer Language Shootout Benchmarks" nor the names of
    its contributors may be used to endorse or promote products derived
    from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

/* The Computer Language Benchmarks Game
 * http://shootout.alioth.debian.org/
 *
 * contributed by The Go Authors.
 * modified by Glenn Brown to use libtcl and glib.
 */

package main

/*
#cgo CFLAGS: -I/usr/include/glib-2.0 -I/usr/lib/glib-2.0/include
#cgo LDFLAGS: -ltcl -lglib-2.0
#include <string.h>
#include <glib.h>		 // apt-get install libglib2.0-dev
#include <tcl8.4/tcl.h> 	 // apt-get install tcl8.4-dev
// cgo requires these trivial shims to use macros and array-indexed pointers:
static inline void ref  (Tcl_Obj *o) { Tcl_IncrRefCount (o); }
static inline void dref (Tcl_Obj *o) { Tcl_DecrRefCount (o); }
static inline int end (Tcl_RegExpInfo *i, int n) { return i->matches[n].end; }
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"os"
	"unsafe"
)

////////////////////////////////////////////////////////////////
// Regexp library shims, similar to regex-dna.c
////////////////////////////////////////////////////////////////

// Return a pointer to the array backing the slice
func data(bytes []byte) unsafe.Pointer {
	return *((*unsafe.Pointer)(unsafe.Pointer(&bytes)))
}

// Count the occurences of a pattern in input
func reCount(pattern []byte, input []byte) (count int) {
	// Compile regexp
	so := C.Tcl_NewStringObj((*C.char)(data(pattern)),
		(C.int)(len(pattern)))
	C.ref(so)
	defer C.dref(so)
	re := C.Tcl_GetRegExpFromObj(nil, so,
		C.TCL_REG_ADVANCED|C.TCL_REG_NOCASE|C.TCL_REG_NEWLINE)
	if re == nil {
		panic("Could not compile regexp \"" + string(pattern) + "\"")
	}
	
	// Count occurances.
	in := C.Tcl_NewStringObj((*C.char)(data(input)), (C.int)(len(input)))
	idx_max := len(input)
	for idx := 0; idx < idx_max; {
		rv := C.Tcl_RegExpExecObj(nil, re, in, C.int(idx), 1, 0)
		if rv == -1 {
			panic ("Tcl_RegExpExecObj")
		}
		if rv == 0 {
			break
		}
		var info C.Tcl_RegExpInfo
		C.Tcl_RegExpGetInfo(re, &info)
		idx += int(C.end(&info, 0))
		count++
	}
	return
}

// Return input with pattern replaced by sub.
func reSub(pattern []byte, sub []byte, input []byte) []byte {
	var err *C.GError = nil
	re := C.g_regex_new(
		(*C.gchar)(data(append(pattern, 0))),
		C.G_REGEX_CASELESS|
			C.G_REGEX_RAW|
			C.G_REGEX_NO_AUTO_CAPTURE|
			C.G_REGEX_OPTIMIZE,
		0,
		&err)
	if err != nil {
		panic("g_regex_new")
	}
	defer C.g_regex_unref(re)
	subd := C.g_regex_replace_literal(re, (*C.gchar)(data(input)),
		C.gssize(len(input)), 0, (*C.gchar)(data(sub)), 0, &err)
	if err != nil {
		panic("g_regex_replace_literal")
	}
	defer C.g_free (C.gpointer(subd));
	l := C.strlen((*C.char)(subd))
	rv := make([]byte, l)
	C.memcpy(data(rv), unsafe.Pointer (subd), l)
	return rv
}

////////////////////////////////////////////////////////////////
// Original Go Authors' program, tweaked to use library shims above.
////////////////////////////////////////////////////////////////

var variants = []string{
	"agggtaaa|tttaccct",
	"[cgt]gggtaaa|tttaccc[acg]",
	"a[act]ggtaaa|tttacc[agt]t",
	"ag[act]gtaaa|tttac[agt]ct",
	"agg[act]taaa|ttta[agt]cct",
	"aggg[acg]aaa|ttt[cgt]ccct",
	"agggt[cgt]aa|tt[acg]accct",
	"agggta[cgt]a|t[acg]taccct",
	"agggtaa[cgt]|[acg]ttaccct",
}

type Subst struct {
	pat, repl string
}

var substs = []Subst{
	Subst{"B", "(c|g|t)"},
	Subst{"D", "(a|g|t)"},
	Subst{"H", "(a|c|t)"},
	Subst{"K", "(g|t)"},
	Subst{"M", "(a|c)"},
	Subst{"N", "(a|c|g|t)"},
	Subst{"R", "(a|g)"},
	Subst{"S", "(c|g)"},
	Subst{"V", "(a|c|g)"},
	Subst{"W", "(a|t)"},
	Subst{"Y", "(c|t)"},
}

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read input: %s\n", err)
		os.Exit(2)
	}
	ilen := len(bytes)
	// Delete the comment lines and newlines
	bytes = reSub([]byte("(>[^\n]+)?\n"), []byte{}, bytes)
	clen := len(bytes)
	for _, s := range variants {
		fmt.Printf("%s %d\n", s, reCount([]byte(s), bytes))
	}
	for _, sub := range substs {
		bytes = reSub([]byte(sub.pat), []byte(sub.repl), bytes)
	}
	fmt.Printf("\n%d\n%d\n%d\n", ilen, clen, len(bytes))
}
