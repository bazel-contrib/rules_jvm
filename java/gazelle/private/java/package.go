package java

type Package struct {
	Name string

	Imports []string
	Mains   []string

	// Especially useful for module mode
	Files       []string
	TestPackage bool
}
