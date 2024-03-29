package admin

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"reflect"
	"strings"
)

var (
	// ErrUnsupportedEncoder unsupported encoder error
	ErrUnsupportedEncoder = errors.New("unsupported encoder")
	// ErrUnsupportedDecoder unsupported decoder error
	ErrUnsupportedDecoder = errors.New("unsupported decoder")
)

// DefaultTransformer registered encoders, decoders for admin
var DefaultTransformer = &Transformer{
	Encoders: map[string][]EncoderInterface{},
	Decoders: map[string][]DecoderInterface{},
}

func init() {
	DefaultTransformer.RegisterTransformer("xml", &XMLTransformer{})
	DefaultTransformer.RegisterTransformer("json", &JSONTransformer{})
	DefaultTransformer.RegisterTransformer("csv", &CSVTransformer{})
}

// Transformer encoder & decoder transformer
type Transformer struct {
	Encoders map[string][]EncoderInterface
	Decoders map[string][]DecoderInterface
}

// RegisterTransformer register transformers for encode, decode
func (transformer *Transformer) RegisterTransformer(format string, transformers ...interface{}) error {
	format = "." + strings.TrimPrefix(format, ".")

	for _, e := range transformers {
		valid := false

		if encoder, ok := e.(EncoderInterface); ok {
			valid = true
			transformer.Encoders[format] = append(transformer.Encoders[format], encoder)
		}

		if decoder, ok := e.(DecoderInterface); ok {
			valid = true
			transformer.Decoders[format] = append(transformer.Decoders[format], decoder)
		}

		if !valid {
			return errors.New("invalid encoder/decoder")
		}
	}

	return nil
}

// EncoderInterface encoder interface
type EncoderInterface interface {
	CouldEncode(*Encoder) bool
	Encode(writer io.Writer, encoder *Encoder) error
	IsType(t reflect.Type) bool
}

// Encoder encoder struct used for encode
type Encoder struct {
	Layout   string
	Resource *Resource
	Context  *Context
	Result   interface{}
}

// Encode encode data based on request accept type
func (transformer *Transformer) Encode(wrap func(e EncoderInterface) (done func()), writer io.Writer, encoder *Encoder) (err error) {
	for _, format := range getFormats(encoder.Context.Request) {
		if encoders, ok := transformer.Encoders[format]; ok {
			for _, e := range encoders {
				if e.CouldEncode(encoder) {
					if wrap != nil {
						func() {
							defer wrap(e)()
							if err = e.Encode(writer, encoder); err == ErrUnsupportedEncoder {
								err = nil
							}
						}()
					} else {
						if err = e.Encode(writer, encoder); err == ErrUnsupportedEncoder {
							err = nil
						}
					}

					return
				}
			}
		}
	}

	return ErrUnsupportedEncoder
}

// DecoderInterface decoder interface
type DecoderInterface interface {
	CouldDecode(Decoder) bool
	Decode(writer io.Writer, decoder Decoder) error
}

// Decoder decoder struct used for decode
type Decoder struct {
	Layout   string
	Resource *Resource
	Context  *Context
	Result   interface{}
}

// Decode decode data based on request content type #FIXME
func (transformer *Transformer) Decode(writer io.Writer, decoder Decoder) error {
	for _, format := range getFormats(decoder.Context.Request) {
		if decoders, ok := transformer.Decoders[format]; ok {
			for _, d := range decoders {
				if d.CouldDecode(decoder) {
					if err := d.Decode(writer, decoder); err != ErrUnsupportedDecoder {
						return err
					}
				}
			}
		}
	}

	return ErrUnsupportedDecoder
}

func getFormats(request *http.Request) (formats []string) {
	if format := path.Ext(request.URL.Path); format != "" {
		formats = append(formats, format)
	}

	for _, ext := range strings.Split(request.Header.Get("X-Accept"), ",") {
		formats = append(formats, "."+ext)
	}

	if extensions, err := mime.ExtensionsByType(request.Header.Get("Accept")); err == nil {
		formats = append(formats, extensions...)
	} else {
		// get request format from Accept
		for _, accept := range strings.FieldsFunc(request.Header.Get("Accept"), func(s rune) bool { return string(s) == "," || string(s) == ";" }) {
			if extensions, err := mime.ExtensionsByType(accept); err == nil {
				formats = append(formats, extensions...)
			}
		}
	}

	return
}
