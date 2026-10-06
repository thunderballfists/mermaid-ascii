package cmd

import "testing"

func TestUnicodeShapeClosingDelimiterAtEdge(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, testCase := range []struct {
		source string
		shape  Shape
	}{
		{"A(Start)-->B", ShapeRoundedRectangle},
		{"A(Start)-->|label|B", ShapeRoundedRectangle},
		{"A([Start])-->|loop|A", ShapeRoundedRectangle},
		{"A((Start))==>B", ShapeCircle},
		{"A(((Start)))-->B", ShapeDoubleCircle},
	} {
		t.Run(testCase.source, func(t *testing.T) {
			properties, err := mermaidFileToMap("graph LR\n"+testCase.source, "plain")
			if err != nil {
				t.Fatal(err)
			}
			definition := properties.nodeDefinitions["A"]
			if definition.label != "Start" || definition.shape != testCase.shape {
				t.Fatalf("closing delimiter lost label/shape: %+v", definition)
			}
			edges, _ := properties.data.Get("A")
			if len(edges) != 1 || edges[0].leftCap != arrowCapNone {
				t.Fatalf("closing delimiter became an edge cap: %+v", edges)
			}
		})
	}
}

func TestShapeDelimiterASCIIParsingPreserved(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = true
	properties, err := mermaidFileToMap("graph LR\nA(Start)-->B", "cli")
	if err != nil {
		t.Fatal(err)
	}
	definition := properties.nodeDefinitions["A"]
	if definition.label != "A" || definition.shape != ShapeRectangle {
		t.Fatalf("legacy ASCII shape parsing changed: %+v", definition)
	}
}

func TestUnicodeShapeDelimiterKeepsExplicitCaps(t *testing.T) {
	originalASCII := useAscii
	t.Cleanup(func() { useAscii = originalASCII })
	useAscii = false
	for _, source := range []string{"A)-->B", "A(Start))-->B", "A(Start))-->|label|B"} {
		spec, ok := parseEdgeSpecification(source)
		if !ok || spec.operator != ")-->" {
			t.Errorf("explicit cap lost in %s: %+v", source, spec)
		}
	}
}
