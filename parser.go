package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/leaanthony/webview2generator/types"
	"os"
)

var (
	idlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Comment", `(?:#|//)[^\n]*\n?`},
		{"String", `"(\\"|[^"])*"`},
		{"UUID", `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`},
		{"Hex", `0x[a-fA-F0-9]+`},
		{"Int", `[0-9]+`},
		{"Ident", `[a-zA-Z]\w*`},
		{"Number", `(?:@Int\.)?@Int`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		{"Whitespace", `[ \t\n\r]+`},
	})
	Parser = participle.MustBuild[types.IDL](
		participle.UseLookahead(4),
		participle.Elide("Comment", "Whitespace"),
		participle.Lexer(idlLexer),
	)
)

func ParseIDL(inputFile string, outputDir string) error {

	data, err := os.Open(inputFile)
	if err != nil {
		return err
	}

	idl, err := Parser.Parse("", data)
	if err != nil {
		return err
	}

	err = idl.Process()
	if err != nil {
		return err
	}

	err = idl.Generate(outputDir)
	if err != nil {
		return err
	}
	return nil
}
