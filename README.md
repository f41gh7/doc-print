# doc-print

## Description

 Simple binary for printing docs, it works with k8s golang based operator and helps creating docs for it.
 

## Usage


first install it with command, you must have golang installed at your system
```bash
go install github.com/f41gh7/doc-print
```

 create some go file with struct:
 
 ```go
package  main
// main struct
type TestK8sApi struct {
	//this is some string, comment will be included at doc
	SomeField string
	// this object is nested
	Nested HelperObject
}

// helper struct, will be included at TestK8sApi
type HelperObject struct {
	// this is nested field at TestK8sApi
	SomeNestedField int
}
```

run 
```bash
doc-print  --paths api_test.go  --owner someOwnerName```
```

it will produce markdown doc for your structs:
```markdown
# API Docs

This Document documents the types introduced by the someOwnerName to be consumed by users.

> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## Table of Contents
* [HelperObject](#helperobject)
* [TestK8sApi](#testk8sapi)

## HelperObject

helper struct, will be included at TestK8sApi

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| SomeNestedField | this is nested field at TestK8sApi | int | false |

[Back to TOC](#table-of-contents)

## TestK8sApi

main struct

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| SomeField | this is some string, comment will be included at doc | string | false |
| Nested | this object is nested | [HelperObject](#helperobject) | false |

[Back to TOC](#table-of-contents)

```