package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/leaanthony/webview2generator/types"
	"os"
)

func ParseIDL(inputFile string, outputDir string) error {

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}
	idlLexer := stateful.MustSimple([]stateful.Rule{
		{"Comment", `(?:#|//)[^\n]*\n?`, nil},
		{"String", `"(\\"|[^"])*"`, nil},
		{"UUID", `[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}`, nil},
		{"Hex", `0x[a-fA-F0-9]+`, nil},
		{"Ident", `[a-zA-Z]\w*`, nil},
		{"Number", `(?:\d*\.)?\d+`, nil},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`, nil},
		{"Whitespace", `[ \t\n\r]+`, nil},
	})
	parser := participle.MustBuild(&types.IDL{},
		participle.UseLookahead(4),
		participle.Elide("Comment", "Whitespace"),
		participle.Lexer(idlLexer),
	)

	idl := &types.IDL{}
	err = parser.ParseString("", string(data), idl)
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
