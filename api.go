package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"html"
	"reflect"
	"strings"
)

var (
	firstParagraph = `
# API Docs

This Document documents the types introduced by the %s to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.`
)

var (
	links = map[string]string{
		"metav1.ObjectMeta":                    "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta",
		"metav1.ListMeta":                      "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#listmeta-v1-meta",
		"metav1.LabelSelector":                 "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#labelselector-v1-meta",
		"v1.ResourceRequirements":              "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#resourcerequirements-v1-core",
		"v1.LocalObjectReference":              "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core",
		"v1.SecretKeySelector":                 "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#secretkeyselector-v1-core",
		"v1.PersistentVolumeClaim":             "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#persistentvolumeclaim-v1-core",
		"v1.EmptyDirVolumeSource":              "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#emptydirvolumesource-v1-core",
		"v1.Volume":                            "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#volume-v1-core",
		"v1.VolumeMount":                       "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#volumemount-v1-core",
		"v1.Affinity":                          "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#affinity-v1-core",
		"v1.Toleration":                        "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#toleration-v1-core",
		"v1.Container":                         "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#container-v1-core",
		"v1.EnvVar":                            "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#envvar-v1-core",
		"v1.PersistentVolumeClaimSpec":         "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#persistentvolumeclaimspec-v1-core",
		"v1.PodSecurityContext":                "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#podsecuritycontext-v1-core",
		"v1.DNSPolicy":                         "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#pod-v1-core",
		"v1.TopologySpreadConstraint":          "https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/",
		"appsv1.StatefulSetUpdateStrategyType": "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#statefulsetupdatestrategy-v1-apps",
		"v1.PersistentVolumeClaimStatus":       "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#persistentvolumeclaimstatus-v1-core",
		"v1.PullPolicy":                        "https://kubernetes.io/docs/concepts/containers/images#updating-images",
		"*appsv1.DeploymentStrategyType":       "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#deploymentstrategy-v1-apps",
		"appsv1.DeploymentStrategyType":        "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#deploymentstrategy-v1-apps",
		"*appsv1.RollingUpdateDeployment":      "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#rollingupdatedeployment-v1-apps",
		"appsv1.RollingUpdateDeployment":       "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#rollingupdatedeployment-v1-apps",
		"v12.IngressRule":                      "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ingressrule-v1-networking-k8s-io",
		"v12.IngressTLS":                       "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ingresstls-v1-networking-k8s-io",
		"*v1.Probe":                            "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#probe-v1-core",
		"v1.Probe":                             "https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#probe-v1-core",
	}

	selfLinks              = map[string]string{}
	structsByName          = map[string][]Pair{}
	wellKnownExternalPairs = map[string][]Pair{
		"[v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core)": []Pair{
			{Name: "name", Doc: "name", Type: "string", Mandatory: true},
		},
	}
)

func toSectionLink(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "-", -1)
	return name
}

func printTOC(types []KubeTypes) {
	fmt.Printf("\n## Table of Contents\n")
	for _, t := range types {
		strukt := t[0]
		if len(t) > 1 {
			fmt.Printf("* [%s](#%s)\n", strukt.Name, toSectionLink(strukt.Name))
		}
	}
}

func printAPIDocs(paths []string, docOwner string) {
	fmt.Println(fmt.Sprintf(firstParagraph, docOwner))

	types := ParseDocumentationFrom(paths, true)
	for _, t := range types {
		st := t[0]
		link := fmt.Sprintf("#%s", strings.ToLower(st.Name))
		selfLinks[st.Name] = link
	}

	// we need to parse once more to now add the self links
	types = ParseDocumentationFrom(paths, false)
	printTOC(types)
	for _, t := range types {
		st := t[0]
		// update self link and references for inlined structs
		link := fmt.Sprintf("#%s", strings.ToLower(st.Name))
		selfLinks[st.Name] = link
		wr := wrapInLink(st.Name, link)
		structsByName[wr] = t
		structsByName[st.Name] = t
	}

	for _, t := range types {
		st := t[0]
		if len(t) > 1 {
			fmt.Printf("\n## %s\n\n%s\n\n", st.Name, st.Doc)

			fmt.Println("| Field | Description | Scheme | Required |")
			fmt.Println("| ----- | ----------- | ------ | -------- |")
			fields := t[1:]
			printPairs(fields)
			fmt.Println("")
			fmt.Println("[Back to TOC](#table-of-contents)")
		}
	}
}

func printPairs(fields []Pair) {
	for _, f := range fields {
		if f.EmbedLink != nil {
			// special case
			if *f.EmbedLink == "metav1.TypeMeta" {
				continue
			}
			emb, ok := structsByName[*f.EmbedLink]
			if !ok {
				if known, ok := wellKnownExternalPairs[*f.EmbedLink]; ok {
					emb = known
				} else {
					panic(fmt.Sprintf("possible bug, link: %s not found at exist and well known pairs", *f.EmbedLink))
				}
			}
			printPairs(emb[1:])
		} else {
			fmt.Println("|", f.Name, "|", html.EscapeString(f.Doc), "|", f.Type, "|", f.Mandatory, "|")
		}

	}
}

// Pair of strings. We need the name of fields and the doc
type Pair struct {
	Name, Doc, Type string
	Mandatory       bool
	EmbedLink       *string
}

// KubeTypes is an array to represent all available types in a parsed file. [0] is for the type itself
type KubeTypes []Pair

// ParseDocumentationFrom gets all types' documentation and returns them as an
// array. Each type is again represented as an array (we have to use arrays as we
// need to be sure for the order of the fields). This function returns fields and
// struct definitions that have no documentation as {name, ""}.
func ParseDocumentationFrom(srcs []string, mustRegisterEmbed bool) []KubeTypes {
	var docForTypes []KubeTypes

	for _, src := range srcs {
		pkg := astFrom(src)

		for _, kubType := range pkg.Types {
			if structType, ok := kubType.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType); ok {
				var ks KubeTypes
				ks = append(ks, Pair{kubType.Name, fmtRawDoc(kubType.Doc), "", false, nil})

				for _, field := range structType.Fields.List {
					typeString := fieldType(field.Type)
					fieldMandatory := fieldRequired(field)
					// handle inlined structs
					if isInline(field) {
						typeString = strings.TrimPrefix(typeString, "*")
						ks = append(ks, Pair{"", "", "", false, &typeString})
						continue
					}
					if n := fieldName(field); n != "-" {
						fieldDoc := fmtRawDoc(field.Doc.Text())
						ks = append(ks, Pair{n, fieldDoc, typeString, fieldMandatory, nil})
					}
				}
				docForTypes = append(docForTypes, ks)
			}
		}
	}

	return docForTypes
}

func astFrom(filePath string) *doc.Package {
	fset := token.NewFileSet()
	m := make(map[string]*ast.File)

	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	m[filePath] = f
	apkg, _ := ast.NewPackage(fset, m, nil, nil)

	return doc.New(apkg, "", 0)
}

func fmtRawDoc(rawDoc string) string {
	var buffer bytes.Buffer
	delPrevChar := func() {
		if buffer.Len() > 0 {
			buffer.Truncate(buffer.Len() - 1) // Delete the last " " or "\n"
		}
	}

	// Ignore all lines after ---
	rawDoc = strings.Split(rawDoc, "---")[0]

	for _, line := range strings.Split(rawDoc, "\n") {
		line = strings.TrimRight(line, " ")
		leading := strings.TrimLeft(line, " ")
		switch {
		case len(line) == 0: // Keep paragraphs
			delPrevChar()
			buffer.WriteString("\n\n")
		case strings.HasPrefix(leading, "TODO"): // Ignore one line TODOs
		case strings.HasPrefix(leading, "+"): // Ignore instructions to go2idl
		default:
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				delPrevChar()
				line = "\n" + line + "\n" // Replace it with newline. This is useful when we have a line with: "Example:\n\tJSON-someting..."
			} else {
				line += " "
			}
			buffer.WriteString(line)
		}
	}

	postDoc := strings.TrimRight(buffer.String(), "\n")
	postDoc = strings.Replace(postDoc, "\\\"", "\"", -1) // replace user's \" to "
	postDoc = strings.Replace(postDoc, "\"", "\\\"", -1) // Escape "
	postDoc = strings.Replace(postDoc, "\n", "\\n", -1)
	postDoc = strings.Replace(postDoc, "\t", "\\t", -1)
	postDoc = strings.Replace(postDoc, "|", "\\|", -1)

	return postDoc
}

func toLink(typeName string) string {
	selfLink, hasSelfLink := selfLinks[typeName]
	if hasSelfLink {
		return wrapInLink(typeName, selfLink)
	}

	link, hasLink := links[typeName]
	if hasLink {
		return wrapInLink(typeName, link)
	}

	return typeName
}

func wrapInLink(text, link string) string {
	return fmt.Sprintf("[%s](%s)", text, link)
}

// fieldName returns the name of the field as it should appear in JSON format
// "-" indicates that this field is not part of the JSON representation
func fieldName(field *ast.Field) string {
	jsonTag := ""
	if field.Tag != nil {
		jsonTag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json") // Delete first and last quotation
		// skip json field
		if strings.HasPrefix(jsonTag, "-") {
			return "-"
		}
	}

	jsonTag = strings.Split(jsonTag, ",")[0] // This can return "-"
	if jsonTag == "" {
		if field.Names != nil {
			return field.Names[0].Name
		}
		return field.Type.(*ast.Ident).Name
	}
	return jsonTag
}

// fieldRequired returns whether a field is a required field.
func fieldRequired(field *ast.Field) bool {
	jsonTag := ""
	if field.Tag != nil {
		jsonTag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json") // Delete first and last quotation
		return !strings.Contains(jsonTag, "omitempty")
	}

	return false
}

// isInline checks if struct inlined with json inline.
func isInline(field *ast.Field) bool {
	jsonTag := ""
	if field.Tag != nil {
		jsonTag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("json") // Delete first and last quotation
		return strings.HasPrefix(jsonTag, ",inline")
	}

	return false
}

func fieldType(typ ast.Expr) string {
	switch typ := typ.(type) {
	case *ast.Ident:
		return toLink(typ.Name)
	case *ast.StarExpr:
		return "*" + toLink(fieldType(typ.X))
	case *ast.SelectorExpr:
		e := typ
		pkg := e.X.(*ast.Ident)
		t := e.Sel
		return toLink(pkg.Name + "." + t.Name)
	case *ast.ArrayType:
		return "[]" + toLink(fieldType(typ.Elt))
	case *ast.MapType:
		mapType := typ
		return "map[" + toLink(fieldType(mapType.Key)) + "]" + toLink(fieldType(mapType.Value))
	default:
		return ""
	}
}
