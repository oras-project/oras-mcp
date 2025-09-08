/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package output

import (
	"fmt"
	"io"
	"sync"
)

// Printer prints for status handlers.
type Printer struct {
	lock sync.Mutex
	out  io.Writer
	err  io.Writer
}

// NewPrinter creates a new Printer.
func NewPrinter(out io.Writer, err io.Writer) *Printer {
	return &Printer{out: out, err: err}
}

// Println prints objects concurrent-safely with newline.
func (p *Printer) Println(a ...any) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	_, err := fmt.Fprintln(p.out, a...)
	if err != nil {
		err = fmt.Errorf("display output error: %w", err)
		_, _ = fmt.Fprint(p.err, err)
		return err
	}
	return nil
}
