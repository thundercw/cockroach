// Copyright 2014 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Andrew Bonventre (andybons@gmail.com)

package encoding

import (
	"bytes"
	"math"
	"testing"

	"github.com/cockroachdb/cockroach/util/randutil"
)

func TestFloatMandE(t *testing.T) {
	testCases := []struct {
		Value float64
		E     int
		M     []byte
	}{
		{1.0, 1, []byte{0x02}},
		{10.0, 1, []byte{0x14}},
		{99.0, 1, []byte{0xc6}},
		{99.01, 1, []byte{0xc7, 0x02}},
		{99.0001, 1, []byte{0xc7, 0x01, 0x02}},
		{100.0, 2, []byte{0x02}},
		{100.01, 2, []byte{0x03, 0x01, 0x02}},
		{100.1, 2, []byte{0x03, 0x01, 0x14}},
		{1234, 2, []byte{0x19, 0x44}},
		{9999, 2, []byte{0xc7, 0xc6}},
		{9999.000001, 2, []byte{0xc7, 0xc7, 0x01, 0x01, 0x02}},
		{9999.000009, 2, []byte{0xc7, 0xc7, 0x01, 0x01, 0x12}},
		{9999.00001, 2, []byte{0xc7, 0xc7, 0x01, 0x01, 0x14}},
		{9999.00009, 2, []byte{0xc7, 0xc7, 0x01, 0x01, 0xb4}},
		{9999.000099, 2, []byte{0xc7, 0xc7, 0x01, 0x01, 0xc6}},
		{9999.0001, 2, []byte{0xc7, 0xc7, 0x01, 0x02}},
		{9999.001, 2, []byte{0xc7, 0xc7, 0x01, 0x14}},
		{9999.01, 2, []byte{0xc7, 0xc7, 0x02}},
		{9999.1, 2, []byte{0xc7, 0xc7, 0x14}},
		{10000, 3, []byte{0x02}},
		{10001, 3, []byte{0x03, 0x01, 0x02}},
		{12345, 3, []byte{0x03, 0x2f, 0x5a}},
		{123450, 3, []byte{0x19, 0x45, 0x64}},
		{1234.5, 2, []byte{0x19, 0x45, 0x64}},
		{12.345, 1, []byte{0x19, 0x45, 0x64}},
		{0.123, 0, []byte{0x19, 0x3c}},
		{0.0123, 0, []byte{0x03, 0x2e}},
		{0.00123, -1, []byte{0x19, 0x3c}},
		{1e-307, -153, []byte{0x14}},
		{1e308, 155, []byte{0x2}},
		// The following value cannot be precisely represented as a float.
		// {9223372036854775807, 10, []byte{0x13, 0x2d, 0x43, 0x91, 0x07, 0x89, 0x6d, 0x9b, 0x75, 0x0e}},
	}
	for _, c := range testCases {
		if e, m := floatMandE(nil, c.Value); e != c.E || !bytes.Equal(m, c.M) {
			t.Errorf("unexpected mismatch in E/M for %v. expected E=%v | M=[% x], got E=%v | M=[% x]",
				c.Value, c.E, c.M, e, m)
		}
	}
}

func TestEncodeFloat(t *testing.T) {
	testCases := []struct {
		Value    float64
		Encoding []byte
	}{
		{math.NaN(), []byte{0x013}},
		{math.Inf(-1), []byte{0x14}},
		{-math.MaxFloat64, []byte{0x15, 0x64, 0xfc, 0x60, 0x66, 0x44, 0xe4, 0x9e, 0x82, 0xc0, 0x8d, 0x0}},
		{-1e308, []byte{0x15, 0x64, 0xfd, 0x0}},
		{-10000.0, []byte{0x1d, 0xfd, 0x0}},
		{-9999.0, []byte{0x1e, 0x38, 0x39, 0x00}},
		{-100.0, []byte{0x1e, 0xfd, 0x00}},
		{-99.0, []byte{0x1f, 0x39, 0x00}},
		{-1.0, []byte{0x1f, 0xfd, 0x0}},
		{-0.00123, []byte{0x21, 0x1, 0xe6, 0xc3, 0x0}},
		{-1e-307, []byte{0x21, 0x99, 0xeb, 0x0}},
		{-math.SmallestNonzeroFloat64, []byte{0x21, 0xa1, 0xf5, 0x0}},
		{0, []byte{0x22}},
		{math.SmallestNonzeroFloat64, []byte{0x23, 0x5e, 0xa, 0x0}},
		{1e-307, []byte{0x23, 0x66, 0x14, 0x0}},
		{0.00123, []byte{0x23, 0xfe, 0x19, 0x3c, 0x0}},
		{0.0123, []byte{0x24, 0x03, 0x2e, 0x0}},
		{0.123, []byte{0x24, 0x19, 0x3c, 0x0}},
		{1.0, []byte{0x25, 0x02, 0x0}},
		{10.0, []byte{0x25, 0x14, 0x0}},
		{12.345, []byte{0x25, 0x19, 0x45, 0x64, 0x0}},
		{99.0, []byte{0x25, 0xc6, 0x0}},
		{99.0001, []byte{0x25, 0xc7, 0x01, 0x02, 0x0}},
		{99.01, []byte{0x25, 0xc7, 0x02, 0x0}},
		{100.0, []byte{0x26, 0x02, 0x0}},
		{100.01, []byte{0x26, 0x03, 0x01, 0x02, 0x0}},
		{100.1, []byte{0x26, 0x03, 0x01, 0x14, 0x0}},
		{1234, []byte{0x26, 0x19, 0x44, 0x0}},
		{1234.5, []byte{0x26, 0x19, 0x45, 0x64, 0x0}},
		{9999, []byte{0x26, 0xc7, 0xc6, 0x0}},
		{9999.000001, []byte{0x26, 0xc7, 0xc7, 0x01, 0x01, 0x02, 0x0}},
		{9999.000009, []byte{0x26, 0xc7, 0xc7, 0x01, 0x01, 0x12, 0x0}},
		{9999.00001, []byte{0x26, 0xc7, 0xc7, 0x01, 0x01, 0x14, 0x0}},
		{9999.00009, []byte{0x26, 0xc7, 0xc7, 0x01, 0x01, 0xb4, 0x0}},
		{9999.000099, []byte{0x26, 0xc7, 0xc7, 0x01, 0x01, 0xc6, 0x0}},
		{9999.0001, []byte{0x26, 0xc7, 0xc7, 0x01, 0x02, 0x0}},
		{9999.001, []byte{0x26, 0xc7, 0xc7, 0x01, 0x14, 0x0}},
		{9999.01, []byte{0x26, 0xc7, 0xc7, 0x02, 0x0}},
		{9999.1, []byte{0x26, 0xc7, 0xc7, 0x14, 0x0}},
		{10000, []byte{0x27, 0x02, 0x0}},
		{10001, []byte{0x27, 0x03, 0x01, 0x02, 0x0}},
		{12345, []byte{0x27, 0x03, 0x2f, 0x5a, 0x0}},
		{123450, []byte{0x27, 0x19, 0x45, 0x64, 0x0}},
		{1e308, []byte{0x2f, 0x9b, 0x2, 0x0}},
		{math.MaxFloat64, []byte{0x2f, 0x9b, 0x3, 0x9f, 0x99, 0xbb, 0x1b, 0x61, 0x7d, 0x3f, 0x72, 0x0}},
		{math.Inf(1), []byte{0x30}},
	}

	for i, c := range testCases {
		enc := EncodeFloat(nil, c.Value)
		if !bytes.Equal(enc, c.Encoding) {
			t.Errorf("unexpected mismatch for %v. expected [% x], got [% x]",
				c.Value, c.Encoding, enc)
		}
		if i > 0 {
			if bytes.Compare(testCases[i-1].Encoding, enc) >= 0 {
				t.Errorf("%v: expected [% x] to be less than [% x]",
					c.Value, testCases[i-1].Encoding, enc)
			}
		}
		_, dec, err := DecodeFloat(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if math.IsNaN(c.Value) {
			if !math.IsNaN(dec) {
				t.Errorf("unexpected mismatch for %v. got %v", c.Value, dec)
			}
		} else if dec != c.Value {
			t.Errorf("unexpected mismatch for %v. got %v", c.Value, dec)
		}
	}

	// Test that appending the float to an existing buffer works.
	enc := EncodeFloat([]byte("hello"), 1.23)
	if _, dec, _ := DecodeFloat(enc[5:], nil); dec != 1.23 {
		t.Errorf("unexpected mismatch for %v. got %v", 1.23, dec)
	}
}

func BenchmarkEncodeFloat(b *testing.B) {
	rng, _ := randutil.NewPseudoRand()

	vals := make([]float64, 10000)
	for i := range vals {
		vals[i] = rng.Float64()
	}

	buf := make([]byte, 0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeFloat(buf, vals[i%len(vals)])
	}
}

func BenchmarkDecodeFloat(b *testing.B) {
	rng, _ := randutil.NewPseudoRand()

	vals := make([][]byte, 10000)
	for i := range vals {
		vals[i] = EncodeFloat(nil, rng.Float64())
	}

	buf := make([]byte, 0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DecodeFloat(vals[i%len(vals)], buf)
	}
}
