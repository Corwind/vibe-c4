package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// extractTypesFromFile walks an AST file and extracts structs, interfaces, and functions.
func extractTypesFromFile(fset *token.FileSet, file *ast.File, relFilePath string, pkgInfo *PackageInfo) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					pos := fset.Position(ts.Pos())
					switch t := ts.Type.(type) {
					case *ast.StructType:
						pkgInfo.Structs = append(pkgInfo.Structs, extractStructInfo(ts.Name.Name, t, relFilePath, pos.Line))
					case *ast.InterfaceType:
						pkgInfo.Interfaces = append(pkgInfo.Interfaces, extractInterfaceInfo(ts.Name.Name, t, relFilePath, pos.Line))
					}
				}
			}
		case *ast.FuncDecl:
			pos := fset.Position(d.Pos())
			if d.Recv != nil {
				// Method - store temporarily as a function with a special marker.
				// We'll attach it to its struct later.
				method := extractMethodInfoFromFunc(d, relFilePath, pos.Line)
				recvType := receiverTypeName(d.Recv)
				pkgInfo.Functions = append(pkgInfo.Functions, FunctionInfo{
					Name:     recvType + "." + method.Name,
					Params:   method.Params,
					Returns:  method.Returns,
					FilePath: relFilePath,
					Line:     pos.Line,
				})
			} else {
				pkgInfo.Functions = append(pkgInfo.Functions, extractFunctionInfo(d, relFilePath, pos.Line))
			}
		}
	}
}

func extractStructInfo(name string, st *ast.StructType, filePath string, line int) StructInfo {
	info := StructInfo{
		Name:     name,
		FilePath: filePath,
		Line:     line,
	}

	if st.Fields != nil {
		for _, field := range st.Fields.List {
			typeName := typeToString(field.Type)
			tag := ""
			if field.Tag != nil {
				tag = field.Tag.Value
			}

			if len(field.Names) == 0 {
				// Embedded field
				info.Fields = append(info.Fields, FieldInfo{
					Name:     typeName,
					Type:     typeName,
					Tag:      tag,
					Embedded: true,
				})
			} else {
				for _, n := range field.Names {
					info.Fields = append(info.Fields, FieldInfo{
						Name: n.Name,
						Type: typeName,
						Tag:  tag,
					})
				}
			}
		}
	}

	return info
}

func extractInterfaceInfo(name string, it *ast.InterfaceType, filePath string, line int) InterfaceInfo {
	info := InterfaceInfo{
		Name:     name,
		FilePath: filePath,
		Line:     line,
	}

	if it.Methods != nil {
		for _, method := range it.Methods.List {
			ft, ok := method.Type.(*ast.FuncType)
			if !ok {
				continue
			}
			if len(method.Names) == 0 {
				continue // embedded interface
			}
			info.Methods = append(info.Methods, MethodInfo{
				Name:    method.Names[0].Name,
				Params:  extractParamTypes(ft.Params),
				Returns: extractReturnTypes(ft.Results),
			})
		}
	}

	return info
}

func extractFunctionInfo(fn *ast.FuncDecl, filePath string, line int) FunctionInfo {
	return FunctionInfo{
		Name:     fn.Name.Name,
		Params:   extractParamTypes(fn.Type.Params),
		Returns:  extractReturnTypes(fn.Type.Results),
		FilePath: filePath,
		Line:     line,
	}
}

func extractMethodInfoFromFunc(fn *ast.FuncDecl, filePath string, line int) MethodInfo {
	return MethodInfo{
		Name:     fn.Name.Name,
		Params:   extractParamTypes(fn.Type.Params),
		Returns:  extractReturnTypes(fn.Type.Results),
		FilePath: filePath,
		Line:     line,
	}
}

// attachMethodsToStructs moves methods from Functions to their receiver struct's Methods slice.
func attachMethodsToStructs(pkgInfo *PackageInfo) {
	structMap := make(map[string]*StructInfo)
	for i := range pkgInfo.Structs {
		structMap[pkgInfo.Structs[i].Name] = &pkgInfo.Structs[i]
	}

	var standaloneFunctions []FunctionInfo
	for _, fn := range pkgInfo.Functions {
		parts := strings.SplitN(fn.Name, ".", 2)
		if len(parts) == 2 {
			recvType := parts[0]
			methodName := parts[1]
			if s, ok := structMap[recvType]; ok {
				s.Methods = append(s.Methods, MethodInfo{
					Name:     methodName,
					Params:   fn.Params,
					Returns:  fn.Returns,
					FilePath: fn.FilePath,
					Line:     fn.Line,
				})
				continue
			}
		}
		standaloneFunctions = append(standaloneFunctions, fn)
	}

	pkgInfo.Functions = standaloneFunctions
}

func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	name := typeToString(recv.List[0].Type)
	// Strip pointer prefix to match struct name
	return strings.TrimPrefix(name, "*")
}

func extractParamTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var types []string
	for _, field := range fields.List {
		typeName := typeToString(field.Type)
		if len(field.Names) == 0 {
			types = append(types, typeName)
		} else {
			for range field.Names {
				types = append(types, typeName)
			}
		}
	}
	return types
}

func extractReturnTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var types []string
	for _, field := range fields.List {
		typeName := typeToString(field.Type)
		if len(field.Names) == 0 {
			types = append(types, typeName)
		} else {
			for range field.Names {
				types = append(types, typeName)
			}
		}
	}
	return types
}

// typeToString converts an AST type expression to its string representation.
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + typeToString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", typeToString(t.Len), typeToString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeToString(t.Key), typeToString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan " + typeToString(t.Value)
	case *ast.Ellipsis:
		return "..." + typeToString(t.Elt)
	case *ast.BasicLit:
		return t.Value
	default:
		return "unknown"
	}
}
