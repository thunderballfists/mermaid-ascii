package cmd

import "testing"

func TestUnicodeClassAssignments(t *testing.T) {
	for _, testCase := range []struct {
		name, source, equivalent string
	}{
		{"after edges", "A-->B\nclass A,B box", "A:::box --> B:::box"},
		{"before edges", "class A,B box\nA-->B", "A:::box --> B:::box"},
		{"standalone", "A\nclass A box", "A:::box"},
		{"whitespace and semicolon", "A-->B\nclass\tA,B\tbox;", "A:::box --> B:::box"},
		{"spaced selectors", "A-->B\nclass A, B box ;", "A:::box --> B:::box"},
		{"label preservation", "A[東京]-->B[Café]\nclass A,B box", "A[東京]:::box --> B[Café]:::box"},
		{"nested group", "subgraph Outer\nsubgraph Inner\nA-->B\nclass A,B box\nend\nend", "subgraph Outer\nsubgraph Inner\nA:::box --> B:::box\nend\nend"},
		{"missing reference", "A-->B\nclass Missing box", "A-->B"},
		{"later assignment", "A-->B\nclass A,B box\nclass A ghost", "A:::ghost --> B:::box"},
	} {
		for _, direction := range []string{"LR", "TD"} {
			t.Run(direction+"/"+testCase.name, func(t *testing.T) {
				prefix := "graph " + direction + "\nclassDef box color:#3366cc,stroke:#00aa66\n"
				actualGraph, actual := renderHTMLColors(t, prefix+testCase.source)
				expectedGraph, expected := renderHTMLColors(t, prefix+testCase.equivalent)
				if actual != expected || len(actualGraph.nodes) != len(expectedGraph.nodes) || len(actualGraph.edges) != len(expectedGraph.edges) {
					t.Fatalf("class assignment differs from inline class:\n%s\nwant:\n%s", actual, expected)
				}
			})
		}
	}
}

func TestUnicodeClassDefinitionAfterAssignment(t *testing.T) {
	definition := "classDef box color:#3366cc,stroke:#00aa66"
	_, actual := renderHTMLColors(t, "graph LR\nA-->B\nclass A,B box\n"+definition)
	_, expected := renderHTMLColors(t, "graph LR\n"+definition+"\nA:::box --> B:::box")
	if actual != expected {
		t.Fatalf("late class definition changed output:\n%s", actual)
	}
}

func TestUnicodeClassAssignmentValidation(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, directive := range []string{"class", "class A", "class A,B", "class A, box", "class ,B box", "class A ;"} {
		if _, err := mermaidFileToMap("graph LR\nA-->B\n"+directive, "html"); err == nil {
			t.Errorf("accepted malformed class assignment %q", directive)
		}
	}
}

func TestASCIIClassAssignmentPreservesLegacyParsing(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = true
	properties, err := mermaidFileToMap("graph LR\nA-->B\nclass A,B box", "cli")
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := properties.data.Get("class A,B box"); !exists {
		t.Fatal("Unicode class assignment changed ASCII parsing")
	}
}
