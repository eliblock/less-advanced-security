package github

import (
	"fmt"
	"testing"
)

func (bound lineBound) String() string {
	return fmt.Sprintf("(%d, %d)", bound.start, bound.end)
}

func lineBoundsString(bounds []lineBound) string {
	combined := "<none>"
	for i, bound := range bounds {
		if i == 0 {
			combined = bound.String()
			continue
		}
		combined = fmt.Sprintf("%s, %s", combined, bound)
	}
	return combined
}

func TestPatchToLineBounds(t *testing.T) {
	var tests = []struct {
		name, patch string
		lineBounds  []lineBound
	}{
		{"empty", "", []lineBound{}},
		{"new file", `
@@ -0,0 +1,4 @@
+1
+2
+3
+4
`, []lineBound{{1, 4}},
		},
		{"file with a change in the middle", `
@@ -4,4 +4,5 @@ something_arbitrary_here

def foo():
-     print('hi')
+     print("hello")
+     print("there")

`, []lineBound{{4, 8}}},
		{"file with multiple changes", `
@@ -4,7 +4,7 @@
 logger = logging.getLogger(__name__)

 def print_something():
-    print('some')
+    print('something')


 def log_something():
@@ -24,3 +24,4 @@ def print_something_else():
     logger.info(f"Great {'news'}!")

     print("a print statement")
+    print("and another.")`, []lineBound{{4, 10}, {24, 27}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := patchToLineBounds(tt.patch)
			if err != nil {
				t.Errorf("expected no error but received %q", err)
			}

			for _, expectedBound := range tt.lineBounds {
				found := false
				for _, foundBound := range got {
					if expectedBound == foundBound {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected bound %s but did not find it: found %s", expectedBound, lineBoundsString(got))
				}
			}

			if len(got) != len(tt.lineBounds) {
				t.Errorf("expected %d total bounds but received %d", len(tt.lineBounds), len(got))
			}
		})
	}
}
