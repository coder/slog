package humanfmt

import (
	"bytes"
	"os"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	jlexers "github.com/alecthomas/chroma/lexers/j"
)

// Adapted from https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/styles/bw.go#L7
// https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/formatters/tty_indexed.go
// https://github.com/alecthomas/chroma/blob/2f5349aa18927368dbec6f8c11608bf61c38b2dd/lexers/j/json.go
var nhooyrJSON = chroma.MustNewStyle("nhooyrJSON", chroma.StyleEntries{
	// Magenta.
	chroma.Keyword: "#7f007f",
	// Magenta.
	chroma.Number: "#7f007f",
	// Magenta.
	chroma.Name: "#00007f",
	// Green.
	chroma.String: "#007f00",
})

func highlightJSON(buf []byte) []byte {
	jsonLexer := chroma.Coalesce(jlexers.JSON)
	it, err := jsonLexer.Tokenise(nil, string(buf))
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to tokenize JSON entry: " + err.Error())
		return buf
	}
	b := bytes.NewBuffer(buf[:0])
	err = formatters.TTY8.Format(b, nhooyrJSON, it)
	if err != nil {
		os.Stderr.WriteString("slogjson: failed to format JSON entry: " + err.Error())
		return buf
	}
	return b.Bytes()
}
