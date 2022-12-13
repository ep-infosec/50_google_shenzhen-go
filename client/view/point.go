// Copyright 2018 Google Inc.
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

package view

// Pointer is anything that has a position on the canvas.
type Pointer interface {
	Pt() Point
}

// Point is a basic implementation of Point.
type Point complex128

// Pt implements Pointer.
func (p Point) Pt() Point { return p }

// C translates any Pointer into a complex128.
func C(p Pointer) complex128 { return complex128(p.Pt()) }

// Pt translates any x and y into a Point.
func Pt(x, y float64) Point { return Point(complex(x, y)) }
