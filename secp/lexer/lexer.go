// Package jlexer contains a JSON lexer implementation.
//
// It is expected that it is mostly used with generated parser code, so the interface is tuned
// for a parser that knows what kind of data is expected.
package lexer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/josharian/intern"
)

// Kind determines type of a token.
type Kind byte

const (
	tokenUndef  Kind = iota // No token.
	tokenDelim              // Delimiter: one of '{', '}'.
	tokenString             // A string literal, e.g. "abc\u1234"
	tokenIdent              // Number literal, e.g. 1.5e5
)

// token describes a single token: type, position in the input and value.
type token struct {
	kind Kind // Type of a token.

	next            bool   // Value if a boolean literal token.
	byteValueCloned bool   // true if byteValue was allocated and does not refer to original json body
	byteValue       []byte // Raw value of a token.
	delimValue      byte
}

// Lexer is a JSON lexer: it iterates over JSON tokens in a byte slice.
type Lexer struct {
	Data []byte // Input data given to the lexer.

	start int   // Start of the current token.
	pos   int   // Current unscanned position in the input stream.
	token token // Last scanned token, if token.kind != tokenUndef.

	firstElement bool // Whether current element is the first in array or an object.
	wantSep      byte // A comma or a colon character, which need to occur before a token.

	UseMultipleErrors bool          // If we want to use multiple errors.
	fatalError        error         // Fatal error occurred during lexing. It is usually a syntax error.
	multipleErrors    []*LexerError // Semantic errors occurred during lexing. Marshalling will be continued after finding this errors.
}

// FetchToken scans the input for the next token.
func (r *Lexer) FetchToken() {
	r.token.kind = tokenUndef
	r.start = r.pos

	// Check if r.Data has r.pos element
	// If it doesn't, it mean corrupted input data
	if len(r.Data) < r.pos {
		r.errParse("Unexpected end of data")
		return
	}
	// Determine the type of a token by skipping whitespace and reading the
	// first character.
	for _, c := range r.Data[r.pos:] {
		switch c {
		case ':':
			if r.wantSep == c || r.wantSep == ';' {
				r.pos++
				r.start++
				r.wantSep = 0
			} else {
				r.errSyntax()
			}
		case ';':
			if r.wantSep == c {
				r.pos++
				r.start++
				r.wantSep = 0
			} else if r.wantSep == ':' {
				r.pos++
				r.start++
				r.token.next = true
				return
			} else {
				r.errSyntax()
			}
		case ' ', '\t', '\r', '\n':
			r.pos++
			r.start++

		case '"':
			if r.wantSep != 0 {
				r.errSyntax()
			}

			r.token.kind = tokenString
			r.fetchString()
			return

		case '{':
			if r.wantSep != 0 {
				r.errSyntax()
			}
			r.firstElement = true
			r.token.kind = tokenDelim
			r.token.delimValue = r.Data[r.pos]
			r.pos++
			return

		case '}':
			if !r.firstElement && (r.wantSep != ';' && r.wantSep != ':') {
				r.errSyntax()
			}
			r.wantSep = 0
			r.token.kind = tokenDelim
			r.token.delimValue = r.Data[r.pos]
			r.pos++
			return

		default:
			if r.wantSep != 0 {
				r.errSyntax()
				return
			}
			r.token.kind = tokenIdent
			r.fetchIdent()
			return
		}
	}
	r.fatalError = io.EOF
	return
}

// isTokenEnd returns true if the char can follow a non-delimiter token
func isTokenEnd(c byte) bool {
	return c == '{' || c == '}' || c == ';' || c == ':'
}

// fetchIdent scans a ident literal token.
func (r *Lexer) fetchIdent() {
	r.pos++
	for i, c := range r.Data[r.pos:] {
		if isTokenEnd(c) {
			r.token.byteValue = r.Data[r.start : r.start+i+1]
			return
		}
		r.pos++
		switch c {
		case '"', '\'', '`':
			r.errSyntax()
			return
		}
	}

	r.pos = len(r.Data)
	r.token.byteValue = r.Data[r.start:]
}

// findStringLen tries to scan into the string literal for ending quote char to determine required size.
// The size will be exact if no escapes are present and may be inexact if there are escaped chars.
func findStringLen(data []byte) (isValid bool, length int) {
	for {
		idx := bytes.IndexByte(data, '"')
		if idx == -1 {
			return false, len(data)
		}
		if idx == 0 || (idx > 0 && data[idx-1] != '\\') {
			return true, length + idx
		}

		// count \\\\\\\ sequences. even ident of slashes means quote is not really escaped
		cnt := 1
		for idx-cnt-1 >= 0 && data[idx-cnt-1] == '\\' {
			cnt++
		}
		if cnt%2 == 0 {
			return true, length + idx
		}

		length += idx + 1
		data = data[idx+1:]
	}
}

// unescapeStringToken performs unescaping of string token.
// if no escaping is needed, original string is returned, otherwise - a new one allocated
func (r *Lexer) unescapeStringToken() (err error) {
	data := r.token.byteValue
	var unescapedData []byte

	for {
		i := bytes.IndexByte(data, '\\')
		if i == -1 {
			break
		}

		escapedRune, escapedBytes, err := decodeEscape(data[i:])
		if err != nil {
			r.errParse(err.Error())
			return err
		}

		if unescapedData == nil {
			unescapedData = make([]byte, 0, len(r.token.byteValue))
		}

		var d [4]byte
		s := utf8.EncodeRune(d[:], escapedRune)
		unescapedData = append(unescapedData, data[:i]...)
		unescapedData = append(unescapedData, d[:s]...)

		data = data[i+escapedBytes:]
	}

	if unescapedData != nil {
		r.token.byteValue = append(unescapedData, data...)
		r.token.byteValueCloned = true
	}
	return
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var val rune
	for i := 2; i < len(s) && i < 6; i++ {
		var v byte
		c := s[i]
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			v = c - '0'
		case 'a', 'b', 'c', 'd', 'e', 'f':
			v = c - 'a' + 10
		case 'A', 'B', 'C', 'D', 'E', 'F':
			v = c - 'A' + 10
		default:
			return -1
		}

		val <<= 4
		val |= rune(v)
	}
	return val
}

// decodeEscape processes a single escape sequence and returns ident of bytes processed.
func decodeEscape(data []byte) (decoded rune, bytesProcessed int, err error) {
	if len(data) < 2 {
		return 0, 0, errors.New("incorrect escape symbol \\ at the end of token")
	}

	c := data[1]
	switch c {
	case '"', '/', '\\':
		return rune(c), 2, nil
	case 'b':
		return '\b', 2, nil
	case 'f':
		return '\f', 2, nil
	case 'n':
		return '\n', 2, nil
	case 'r':
		return '\r', 2, nil
	case 't':
		return '\t', 2, nil
	case 'u':
		rr := getu4(data)
		if rr < 0 {
			return 0, 0, errors.New("incorrectly escaped \\uXXXX sequence")
		}

		read := 6
		if utf16.IsSurrogate(rr) {
			rr1 := getu4(data[read:])
			if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
				read += 6
				rr = dec
			} else {
				rr = unicode.ReplacementChar
			}
		}
		return rr, read, nil
	}

	return 0, 0, errors.New("incorrectly escaped bytes")
}

// fetchString scans a string literal token.
func (r *Lexer) fetchString() {
	r.pos++
	data := r.Data[r.pos:]

	isValid, length := findStringLen(data)
	if !isValid {
		r.pos += length
		r.errParse("unterminated string literal")
		return
	}
	r.token.byteValue = data[:length]
	r.pos += length + 1 // skip closing '"' as well
}

// scanToken scans the next token if no token is currently available in the lexer.
func (r *Lexer) scanToken() {
	if r.token.kind != tokenUndef || r.fatalError != nil {
		return
	}

	r.FetchToken()
}

// consume resets the current token to allow scanning the next one.
func (r *Lexer) consume() {
	r.token.kind = tokenUndef
	r.token.byteValueCloned = false
	r.token.delimValue = 0
}

// Ok returns true if no error (including io.EOF) was encountered during scanning.
func (r *Lexer) Ok() bool {
	return r.fatalError == nil
}

const maxErrorContextLen = 13

func (r *Lexer) errParse(what string) {
	if r.fatalError == nil {
		var str string
		if len(r.Data)-r.pos <= maxErrorContextLen {
			str = string(r.Data)
		} else {
			str = string(r.Data[r.pos:r.pos+maxErrorContextLen-3]) + "..."
		}
		r.fatalError = &LexerError{
			Reason: what,
			Offset: r.pos,
			Data:   str,
		}
	}
}

func (r *Lexer) errSyntax() {
	r.errParse("syntax error")
}

func (r *Lexer) errInvalidToken(expected string) {
	if r.fatalError != nil {
		return
	}
	if r.UseMultipleErrors {
		r.pos = r.start
		r.consume()
		r.SkipRecursive()
		switch expected {
		case "[":
			r.token.delimValue = ']'
			r.token.kind = tokenDelim
		case "{":
			r.token.delimValue = '}'
			r.token.kind = tokenDelim
		}
		r.addNonfatalError(&LexerError{
			Reason: fmt.Sprintf("expected %s", expected),
			Offset: r.start,
			Data:   string(r.Data[r.start:r.pos]),
		})
		return
	}

	var str string
	if len(r.token.byteValue) <= maxErrorContextLen {
		str = string(r.token.byteValue)
	} else {
		str = string(r.token.byteValue[:maxErrorContextLen-3]) + "..."
	}
	r.fatalError = &LexerError{
		Reason: fmt.Sprintf("expected %s", expected),
		Offset: r.pos,
		Data:   str,
	}
}

func (r *Lexer) GetPos() int {
	return r.pos
}

// Delim consumes a token and verifies that it is the given delimiter.
func (r *Lexer) Delim(c byte) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() || r.token.delimValue != c {
		r.consume() // errInvalidToken can change token if UseMultipleErrors is enabled.
		r.errInvalidToken(string([]byte{c}))
	} else {
		r.consume()
	}
}

// IsDelim returns true if there was no scanning error and next token is the given delimiter.
func (r *Lexer) IsDelim(c byte) bool {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	fmt.Println(fmt.Sprintf("%q", string(r.token.delimValue)))
	return !r.Ok() || r.token.delimValue == c
}

// Skip skips a single token.
func (r *Lexer) Skip() {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	r.consume()
}

// SkipRecursive skips next array or object completely, or just skips a single token if not
// an array/object.
//
// Note: no syntax validation is performed on the skipped data.
func (r *Lexer) SkipRecursive() {
	r.scanToken()
	var start, end byte
	startPos := r.start

	switch r.token.delimValue {
	case '{':
		start, end = '{', '}'
	default:
		r.consume()
		return
	}

	r.consume()

	level := 1
	inQuotes := false
	wasEscape := false

	for i, c := range r.Data[r.pos:] {
		switch {
		case c == start && !inQuotes:
			level++
		case c == end && !inQuotes:
			level--
			if level == 0 {
				r.pos += i + 1
				if !json.Valid(r.Data[startPos:r.pos]) {
					r.pos = len(r.Data)
					r.fatalError = &LexerError{
						Reason: "skipped array/object json value is invalid",
						Offset: r.pos,
						Data:   string(r.Data[r.pos:]),
					}
				}
				return
			}
		case c == '\\' && inQuotes:
			wasEscape = !wasEscape
			continue
		case c == '"' && inQuotes:
			inQuotes = wasEscape
		case c == '"':
			inQuotes = true
		}
		wasEscape = false
	}
	r.pos = len(r.Data)
	r.fatalError = &LexerError{
		Reason: "EOF reached while skipping array/object or token",
		Offset: r.pos,
		Data:   string(r.Data[r.pos:]),
	}
}

// Raw fetches the next item recursively as a data slice
func (r *Lexer) Raw() []byte {
	r.SkipRecursive()
	if !r.Ok() {
		return nil
	}
	return r.Data[r.start:r.pos]
}

// IsStart returns whether the lexer is positioned at the start
// of an input string.
func (r *Lexer) IsStart() bool {
	return r.pos == 0
}

// Consumed reads all remaining bytes from the input, publishing an error if
// there is anything but whitespace remaining.
func (r *Lexer) Consumed() {
	if r.pos > len(r.Data) || !r.Ok() {
		return
	}

	for _, c := range r.Data[r.pos:] {
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			r.AddError(&LexerError{
				Reason: "invalid character '" + string(c) + "' after top-level value",
				Offset: r.pos,
				Data:   string(r.Data[r.pos:]),
			})
			return
		}

		r.pos++
		r.start++
	}
}

func (r *Lexer) unsafeString(skipUnescape bool) (string, []byte) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")
		return "", nil
	}
	if !skipUnescape {
		if err := r.unescapeStringToken(); err != nil {
			r.errInvalidToken("string")
			return "", nil
		}
	}

	bytes := r.token.byteValue
	ret := bytesToStr(r.token.byteValue)
	r.consume()
	return ret, bytes
}

// UnsafeString returns the string value if the token is a string literal.
//
// Warning: returned string may point to the input buffer, so the string should not outlive
// the input buffer. Intended pattern of usage is as an argument to a switch statement.
func (r *Lexer) UnsafeString() string {
	ret, _ := r.unsafeString(false)
	return ret
}

// UnsafeBytes returns the byte slice if the token is a string literal.
func (r *Lexer) UnsafeBytes() []byte {
	_, ret := r.unsafeString(false)
	return ret
}

// UnsafeFieldName returns current member name string token
func (r *Lexer) UnsafeFieldName(skipUnescape bool) string {
	ret, _ := r.unsafeString(skipUnescape)
	return ret
}

// String reads a string literal.
func (r *Lexer) String() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	if !r.Ok() || (r.token.kind != tokenString && r.token.kind != tokenIdent) {
		r.errInvalidToken("string|ident")
		return ""
	}
	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")
		return ""
	}
	var ret string
	if r.token.byteValueCloned {
		ret = bytesToStr(r.token.byteValue)
	} else {
		ret = string(r.token.byteValue)
	}
	r.consume()
	return ret
}

// StringIntern reads a string literal, and performs string interning on it.
func (r *Lexer) StringIntern() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")
		return ""
	}
	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")
		return ""
	}
	ret := intern.Bytes(r.token.byteValue)
	r.consume()
	return ret
}

// Bytes reads a string literal and base64 decodes it into a byte slice.
func (r *Lexer) Bytes() []byte {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	if !r.Ok() || r.token.kind != tokenString {
		r.errInvalidToken("string")
		return nil
	}
	if err := r.unescapeStringToken(); err != nil {
		r.errInvalidToken("string")
		return nil
	}
	ret := make([]byte, base64.StdEncoding.DecodedLen(len(r.token.byteValue)))
	n, err := base64.StdEncoding.Decode(ret, r.token.byteValue)
	if err != nil {
		r.fatalError = &LexerError{
			Reason: err.Error(),
		}
		return nil
	}

	r.consume()
	return ret[:n]
}

func (r *Lexer) ident() string {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}
	if !r.Ok() || r.token.kind != tokenIdent {
		r.errInvalidToken("ident")
		return ""
	}
	ret := bytesToStr(r.token.byteValue)
	r.consume()
	return ret
}

func (r *Lexer) Error() error {
	return r.fatalError
}

func (r *Lexer) AddError(e error) {
	if r.fatalError == nil {
		r.fatalError = e
	}
}

func (r *Lexer) AddNonFatalError(e error) {
	r.addNonfatalError(&LexerError{
		Offset: r.start,
		Data:   string(r.Data[r.start:r.pos]),
		Reason: e.Error(),
	})
}

func (r *Lexer) addNonfatalError(err *LexerError) {
	if r.UseMultipleErrors {
		// We don't want to add errors with the same offset.
		if len(r.multipleErrors) != 0 && r.multipleErrors[len(r.multipleErrors)-1].Offset == err.Offset {
			return
		}
		r.multipleErrors = append(r.multipleErrors, err)
		return
	}
	r.fatalError = err
}

func (r *Lexer) GetNonFatalErrors() []*LexerError {
	return r.multipleErrors
}

// Interface fetches an interface{} analogous to the 'encoding/json' package.
func (r *Lexer) Interface() interface{} {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() {
		return nil
	}
	switch r.token.kind {
	case tokenString:
		return r.String()
	case tokenIdent:
		return r.String()
	}

	if r.token.delimValue == '{' {
		return r.Statements()
	} else if r.token.delimValue == '}' {
		return nil
	}
	if r.token.next {
		r.token.next = false
		return nil
	}
	r.errSyntax()
	return nil
}

// Interface fetches an interface{} analogous to the 'encoding/json' package.
func (r *Lexer) Statements() (ret []*Stmt) {
	if r.token.kind == tokenUndef && r.Ok() {
		r.FetchToken()
	}

	if !r.Ok() {
		return nil
	}
	if r.token.kind != tokenDelim && r.token.delimValue != '{' {
		r.errInvalidToken("{")
		return nil
	}

	r.consume()
	for !r.IsDelim('}') {
		key := r.Interface()
		if !r.Ok() {
			return nil
		}
		r.WantColon()
		val := r.Interface()
		if val == nil {
			ret = append(ret, &Stmt{Value: key})
			if r.token.kind == tokenDelim && r.token.delimValue == '}' {
				break
			}
			r.token.kind = tokenUndef
			r.wantSep = 0
		} else {
			ret = append(ret, &Stmt{key, val})
			r.WantComma()
		}
	}
	r.Delim('}')

	if r.Ok() {
		return
	}
	return nil
}

// WantComma requires a comma to be present before fetching next token.
func (r *Lexer) WantComma() {
	r.wantSep = ';'
	r.firstElement = false
}

// WantColon requires a colon to be present before fetching next token.
func (r *Lexer) WantColon() {
	r.wantSep = ':'
	r.firstElement = false
}
