package analyzer

import "strings"

type methodSig struct {
	Name    string
	Params  string
	Returns string
}

// matchInterfaceImpls finds all struct types that implement interfaces by comparing method sets.
func matchInterfaceImpls(packages []PackageInfo) []InterfaceImpl {
	type ifaceEntry struct {
		pkg     string
		name    string
		methods []methodSig
	}

	type structEntry struct {
		pkg     string
		name    string
		methods map[string]methodSig
	}

	// Collect all interfaces
	var ifaces []ifaceEntry
	for _, pkg := range packages {
		for _, iface := range pkg.Interfaces {
			var methods []methodSig
			for _, m := range iface.Methods {
				methods = append(methods, methodSig{
					Name:    m.Name,
					Params:  strings.Join(m.Params, ","),
					Returns: strings.Join(m.Returns, ","),
				})
			}
			if len(methods) > 0 {
				ifaces = append(ifaces, ifaceEntry{
					pkg:     pkg.ImportPath,
					name:    iface.Name,
					methods: methods,
				})
			}
		}
	}

	// Collect all structs with their method sets
	var structs []structEntry
	for _, pkg := range packages {
		for _, s := range pkg.Structs {
			methodSet := make(map[string]methodSig)
			for _, m := range s.Methods {
				methodSet[m.Name] = methodSig{
					Name:    m.Name,
					Params:  strings.Join(m.Params, ","),
					Returns: strings.Join(m.Returns, ","),
				}
			}
			if len(methodSet) > 0 {
				structs = append(structs, structEntry{
					pkg:     pkg.ImportPath,
					name:    s.Name,
					methods: methodSet,
				})
			}
		}
	}

	var impls []InterfaceImpl
	for _, iface := range ifaces {
		for _, s := range structs {
			if implementsInterface(s.methods, iface.methods) {
				impls = append(impls, InterfaceImpl{
					StructPkg:     s.pkg,
					StructName:    s.name,
					InterfacePkg:  iface.pkg,
					InterfaceName: iface.name,
				})
			}
		}
	}

	return impls
}

func implementsInterface(structMethods map[string]methodSig, ifaceMethods []methodSig) bool {
	for _, im := range ifaceMethods {
		sm, ok := structMethods[im.Name]
		if !ok {
			return false
		}
		if sm.Params != im.Params || sm.Returns != im.Returns {
			return false
		}
	}
	return true
}
