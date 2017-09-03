package generic_types

// This file contains function definitions needed for the template package to
// be compilable but that should be replaced by other function calls in the templates

func genericHash(x interface{}) uint32 {
	return interfaceHash(x)
}
