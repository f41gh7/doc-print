package main


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