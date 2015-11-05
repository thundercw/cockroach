// Copyright 2015 The Cockroach Authors.
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
// Author: Ben Darnell

package pgwire

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"

	"github.com/cockroachdb/cockroach/util"
)

const maxMessageSize = 1 << 24

// bufferedReader is implemented by bufio.Reader and bytes.Buffer.
type bufferedReader interface {
	io.Reader
	ReadString(delim byte) (string, error)
	ReadByte() (byte, error)
}

type readBuffer struct {
	msg []byte
	tmp [4]byte
}

// readMsg reads a length-prefixed message. It is only used directly
// during the authentication phase of the protocol; readTypedMsg is
// used at all other times.
func (b *readBuffer) readUntypedMsg(rd io.Reader) error {
	if _, err := io.ReadFull(rd, b.tmp[:]); err != nil {
		return err
	}
	size := int32(binary.BigEndian.Uint32(b.tmp[:]))
	// size includes itself.
	size -= 4
	if size > maxMessageSize || size < 0 {
		return util.Errorf("message size %d out of bounds (0..%d)",
			size, maxMessageSize)
	}
	b.msg = make([]byte, size)
	_, err := io.ReadFull(rd, b.msg)
	return err
}

// readTypedMsg reads a message, returning its type code and body.
func (b *readBuffer) readTypedMsg(rd bufferedReader) (messageType, error) {
	typ, err := rd.ReadByte()
	if err != nil {
		return 0, err
	}
	return messageType(typ), b.readUntypedMsg(rd)
}

// getString reads a null-terminated string.
func (b *readBuffer) getString() (string, error) {
	pos := bytes.IndexByte(b.msg, 0)
	if pos == -1 {
		return "", util.Errorf("NUL terminator not found")
	}
	// Note: this is a conversion from a byte slice to a string which avoids
	// allocation and copying. It is safe because we never reuse the bytes in our
	// read buffer. It is effectively the same as: "s := string(b.msg[0:pos])"
	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(unsafe.Pointer(&b.msg[0]))
	hdr.Len = pos
	b.msg = b.msg[pos+1:]
	return s, nil
}

func (b *readBuffer) getInt16() (int16, error) {
	if len(b.msg) < 2 {
		return 0, util.Errorf("insufficient data: %d", len(b.msg))
	}
	v := int16(binary.BigEndian.Uint16(b.msg[:2]))
	b.msg = b.msg[2:]
	return v, nil
}

func (b *readBuffer) getInt32() (int32, error) {
	if len(b.msg) < 4 {
		return 0, util.Errorf("insufficient data: %d", len(b.msg))
	}
	v := int32(binary.BigEndian.Uint32(b.msg[:4]))
	b.msg = b.msg[4:]
	return v, nil
}

type writeBuffer struct {
	bytes.Buffer
	putbuf [64]byte
}

// writeString writes a null-terminated string.
func (b *writeBuffer) writeString(s string) error {
	if _, err := b.WriteString(s); err != nil {
		return err
	}
	return b.WriteByte(0)
}

func (b *writeBuffer) putInt16(v int16) {
	binary.BigEndian.PutUint16(b.putbuf[:], uint16(v))
	b.Write(b.putbuf[:2])
}

func (b *writeBuffer) putInt32(v int32) {
	binary.BigEndian.PutUint32(b.putbuf[:], uint32(v))
	b.Write(b.putbuf[:4])
}

func (b *writeBuffer) putInt64(v int64) {
	binary.BigEndian.PutUint64(b.putbuf[:], uint64(v))
	b.Write(b.putbuf[:8])
}

func (b *writeBuffer) initMsg(typ messageType) {
	b.Reset()
	b.Write(b.putbuf[:5]) // message type + message length
	b.Bytes()[0] = byte(typ)
}

func (b *writeBuffer) finishMsg(w io.Writer) error {
	bytes := b.Bytes()
	binary.BigEndian.PutUint32(bytes[1:5], uint32(b.Len()-1))
	_, err := w.Write(bytes)
	b.Reset()
	return err
}