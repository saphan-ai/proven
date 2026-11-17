package codecomments

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type goAdapter struct{}

func (goAdapter) Name() string { return "go" }
func (goAdapter) Matches(p string) bool {
	return strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go")
}
func (goAdapter) Extract(path string, data []byte) ([]Unit, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var units []Unit
	if ast.IsGenerated(f) {
		return units, nil
	}
	lines := strings.Split(string(data), "\n")
	used := map[token.Pos]bool{}
	add := func(key, name string, public bool, node ast.Node, doc *ast.CommentGroup, params []string) {
		first, last := fset.Position(node.Pos()), fset.Position(node.End())
		// Keep full lines so a model quote can be checked exactly.
		text, line := "", first.Line
		if doc != nil {
			text = doc.Text()
			line = fset.Position(doc.Pos()).Line
			used[doc.Pos()] = true
		}
		units = append(units, Unit{ID: path + "::" + key, Path: path, Symbol: name, Public: public, Line: line, CodeLine: first.Line, Text: text, Code: strings.Join(lines[first.Line-1:last.Line], "\n"), Params: params})
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			key := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recv := d.Recv.List[0].Type
				if p, ok := recv.(*ast.StarExpr); ok {
					recv = p.X
				}
				key = fmt.Sprintf("%s.%s", string(data[fset.Position(recv.Pos()).Offset:fset.Position(recv.End()).Offset]), key)
			}
			params := []string{}
			for _, field := range d.Type.Params.List {
				for _, n := range field.Names {
					params = append(params, n.Name)
				}
			}
			add("func:"+key, d.Name.Name, d.Name.IsExported(), d, d.Doc, params)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					doc := s.Doc
					if doc == nil {
						doc = d.Doc
					}
					add("type:"+s.Name.Name, s.Name.Name, s.Name.IsExported(), s, doc, nil)
				case *ast.ValueSpec:
					doc := s.Doc
					if doc == nil {
						doc = d.Doc
					}
					for _, n := range s.Names {
						add("value:"+n.Name, n.Name, n.IsExported(), s, doc, nil)
					}
				}
			}
		}
	}
	for i, c := range f.Comments {
		if used[c.Pos()] {
			continue
		}
		first, last := fset.Position(c.Pos()).Line, fset.Position(c.End()).Line
		lo, hi := max(0, first-5), min(len(lines), last+5)
		u := Unit{ID: fmt.Sprintf("%s::comment:%d", path, i), Path: path, Line: first, CodeLine: lo + 1, Text: c.Text(), Code: strings.Join(lines[lo:hi], "\n")}
		units = append(units, u)
	}
	return units, nil
}
