package lexer

import (
	"bytes"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: `"simple string"`, want: "simple string"},
		{toParse: " \r\r\n\t  " + `"test"`, want: "test"},
		{toParse: `"\n\t\"\/\\\f\r"`, want: "\n\t\"/\\\f\r"},
		{toParse: `"\u0020"`, want: " "},
		{toParse: `"\u0020-\t"`, want: " -\t"},
		{toParse: `"\ufffd\uFFFD"`, want: "\ufffd\ufffd"},
		{toParse: `"\ud83d\ude00"`, want: "😀"},
		{toParse: `"\ud83d\ude08"`, want: "😈"},
		{toParse: `"\ud8"`, wantError: true},

		{toParse: `"test"junk`, want: "test"},

		{toParse: `5`, wantError: true},    // not a string
		{toParse: `"\x"`, wantError: true}, // invalid escape
		{toParse: `"\ud800"`, want: "�"},   // invalid utf-8 char; return replacement char
	} {
		{
			l := Lexer{Data: []byte(test.toParse)}

			got := l.String()
			if got != test.want {
				t.Errorf("[%d, %q] String() = %v; want %v", i, test.toParse, got, test.want)
			}
			err := l.Error()
			if err != nil && !test.wantError {
				t.Errorf("[%d, %q] String() error: %v", i, test.toParse, err)
			} else if err == nil && test.wantError {
				t.Errorf("[%d, %q] String() ok; want error", i, test.toParse)
			}
		}
		{
			l := Lexer{Data: []byte(test.toParse)}

			got := l.StringIntern()
			if got != test.want {
				t.Errorf("[%d, %q] String() = %v; want %v", i, test.toParse, got, test.want)
			}
			err := l.Error()
			if err != nil && !test.wantError {
				t.Errorf("[%d, %q] String() error: %v", i, test.toParse, err)
			} else if err == nil && test.wantError {
				t.Errorf("[%d, %q] String() ok; want error", i, test.toParse)
			}
		}
	}
}

func TestStringIntern(t *testing.T) {
	data := []byte(`"string interning test"`)
	var l Lexer

	allocsPerRun := testing.AllocsPerRun(1000, func() {
		l = Lexer{Data: data}
		_ = l.StringIntern()
	})
	if allocsPerRun != 0 {
		t.Fatalf("expected 0 allocs, got %f", allocsPerRun)
	}

	allocsPerRun = testing.AllocsPerRun(1000, func() {
		l = Lexer{Data: data}
		_ = l.String()
	})
	if allocsPerRun != 1 {
		t.Fatalf("expected 1 allocs, got %f", allocsPerRun)
	}
}

func TestBytes(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: `"c2ltcGxlIHN0cmluZw=="`, want: "simple string"},
		{toParse: " \r\r\n\t  " + `"dGVzdA=="`, want: "test"},
		{toParse: `"c3ViamVjdHM\/X2Q9MQ=="`, want: "subjects?_d=1"}, // base64 with forward slash escaped

		{toParse: `5`, wantError: true},                     // not a JSON string
		{toParse: `"foobar"`, wantError: true},              // not base64 encoded
		{toParse: `"c2ltcGxlIHN0cmluZw="`, wantError: true}, // invalid base64 padding
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Bytes()
		if bytes.Compare(got, []byte(test.want)) != 0 {
			t.Errorf("[%d, %q] Bytes() = %v; want: %v", i, test.toParse, got, []byte(test.want))
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Bytes() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Bytes() ok; want error", i, test.toParse)
		}
	}
}

func TestIdent(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      string
		wantError bool
	}{
		{toParse: "123", want: "123"},
		{toParse: "\r\n12 35", want: "12 35"},
		{toParse: "12.35e+1", want: "12.35e+1"},
		{toParse: "12.35e-15", want: "12.35e-15"},
		{toParse: "12.35E-15", want: "12.35E-15"},
		{toParse: "12.35E15", want: "12.35E15"},
		{toParse: "123junk", want: "123junk"},

		{toParse: `"a"`, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.ident()
		if got != test.want {
			t.Errorf("[%d, %q] ident() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] ident() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] ident() ok; want error", i, test.toParse)
		}
	}
}

func TestSkipRecursive(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		left      string
		wantError bool
	}{
		{toParse: "{5, 6}, 4", left: ", 4"},
		{toParse: "{5, {7,8}}: 4", left: ": 4"},

		{toParse: `{"a":1}, 4`, left: ", 4"},
		{toParse: `{"a":1, "b":{"c": 5}, "e":{12,15}}, 4`, left: ", 4"},

		// array start/end chars in a string
		{toParse: `{5, "]"}, 4`, left: ", 4"},
		{toParse: `{5, "\"}"}, 4`, left: ", 4"},
		{toParse: `{5, "{"}, 4`, left: ", 4"},
		{toParse: `{5, "\"{"}, 4`, left: ", 4"},

		// object start/end chars in a string
		{toParse: `{"a}":1}, 4`, left: ", 4"},
		{toParse: `{"a\"}":1}, 4`, left: ", 4"},
		{toParse: `{"a{":1}, 4`, left: ", 4"},
		{toParse: `{"a\"{":1}, 4`, left: ", 4"},

		// object with double slashes at the end of string
		{toParse: `{"a":"hey\\"}, 4`, left: ", 4"},

		// make sure skipping an invalid json results in an error
		{toParse: `{"a": { ##invalid json## }}, 4`, wantError: true},
		{toParse: `{"a": { {1}, { ##invalid json## }}}, 4`, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		l.SkipRecursive()

		got := string(l.Data[l.pos:])
		if got != test.left {
			t.Errorf("[%d, %q] SkipRecursive() left = %v; want %v", i, test.toParse, got, test.left)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] SkipRecursive() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] SkipRecursive() ok; want error", i, test.toParse)
		}
	}
}

func TestInterface(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		want      interface{}
		wantError bool
	}{
		//		{toParse: "true", want: "true"},
		//		{toParse: `"a"`, want: "a"},
		//		{toParse: "5", want: "5"},

		//		{toParse: `{}`, want: []interface{}{}},

		//		{toParse: `{"a": "b"}`, want: []interface{}{[]interface{}{"a", "b"}}},
		{toParse: `{"[2;{}45]"  ;d:e;f;g:{h:{i:j}}}`, want: []interface{}{[]interface{}{"a", "b"}, "c", []interface{}{"d", "e"}}},
		{toParse: `{"c";"d":"e";"f":"g";"h"}`, want: []interface{}{[]interface{}{"a", "b"}, "c", []interface{}{"d", "e"}}},
		//		{toParse: `{"a":5 ; "b" : "string"}`, want: map[string]interface{}{"a": float64(5), "b": "string"}},

		//{toParse: `{"a" "b"}`, wantError: true},
		//{toParse: `{"a": "b";}`, wantError: true},
		//{toParse: `{"a":"b","c" "b"}`, wantError: true},
		//{toParse: `{"a": "b","c":"d";}`, wantError: true},
		//{toParse: `{;}`, wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}

		got := l.Interface()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("[%d, %q] Interface() = %v; want %v", i, test.toParse, got, test.want)
		}
		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Interface() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Interface() ok; want error", i, test.toParse)
		}
	}
}

func TestConsumed(t *testing.T) {
	for i, test := range []struct {
		toParse   string
		wantError bool
	}{
		{toParse: "", wantError: false},
		{toParse: "   ", wantError: false},
		{toParse: "\r\n", wantError: false},
		{toParse: "\t\t", wantError: false},

		{toParse: "{", wantError: true},
	} {
		l := Lexer{Data: []byte(test.toParse)}
		l.Consumed()

		err := l.Error()
		if err != nil && !test.wantError {
			t.Errorf("[%d, %q] Consumed() error: %v", i, test.toParse, err)
		} else if err == nil && test.wantError {
			t.Errorf("[%d, %q] Consumed() ok; want error", i, test.toParse)
		}
	}
}

func TestFetchStringUnterminatedString(t *testing.T) {
	for _, test := range []struct {
		data []byte
	}{
		{data: []byte(`"sting without trailing quote`)},
		{data: []byte(`"\"`)},
		{data: []byte{'"'}},
	} {
		l := Lexer{Data: test.data}
		l.fetchString()
		if l.pos > len(l.Data) {
			t.Errorf("fetchString(%s): pos=%v should not be greater than length of Data = %v", test.data, l.pos, len(l.Data))
		}
		if l.Error() == nil {
			t.Errorf("fetchString(%s): should add parsing error", test.data)
		}
	}
}
