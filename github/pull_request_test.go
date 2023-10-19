package github

import (
	_ "embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-github/v47/github"
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

func (file *pullRequestFile) String() string {
	if file == nil {
		return "<nil file>"
	}
	return fmt.Sprintf("%s with patches on %s", file.filename, lineBoundsString(file.lineBounds))
}

func pullRequestFilesString(files []*pullRequestFile) string {
	combined := "<none>"
	for i, file := range files {
		if i == 0 {
			combined = file.String()
			continue
		}
		combined = fmt.Sprintf("%s, %s", combined, file)
	}
	return combined
}

func TestFilterAnnotations(t *testing.T) {
	main := "src/main.go"
	test := "src/main_test.go"
	pullRequestFile_1 := pullRequestFile{
		filename:   main,
		patch:      "...",
		lineBounds: []lineBound{{5, 15}},
	}
	pullRequestFile_2 := pullRequestFile{
		filename:   test,
		patch:      "...",
		lineBounds: []lineBound{{1, 5}, {11, 15}},
	}
	pr_empty := pullRequest{
		files: []*pullRequestFile{},
	}
	pr_main_and_test := pullRequest{
		files: []*pullRequestFile{&pullRequestFile_1, &pullRequestFile_2},
	}
	pr_test := pullRequest{
		files: []*pullRequestFile{&pullRequestFile_2},
	}

	main_annotation_in_bound := Annotation{
		fileName:  main,
		startLine: 6,
		endLine:   7,
	}
	main_annotation_out_of_bounds := Annotation{
		fileName:  main,
		startLine: 2,
		endLine:   3,
	}
	main_annotation_covering := Annotation{
		fileName:  main,
		startLine: 1,
		endLine:   100,
	}

	test_annotation_start_in_bound := Annotation{
		fileName:  test,
		startLine: 5,
		endLine:   6,
	}
	test_annotation_end_in_bound := Annotation{
		fileName:  test,
		startLine: 8,
		endLine:   11,
	}
	test_annotation_out_of_bounds := Annotation{
		fileName:  test,
		startLine: 6,
		endLine:   9,
	}

	other_annotation := Annotation{
		fileName:  "other_file.py",
		startLine: 1,
		endLine:   100,
	}

	tests := []struct {
		name                             string
		pr                               pullRequest
		annotations, filteredAnnotations []*Annotation
	}{
		{
			"empty PR",
			pr_empty,
			[]*Annotation{&main_annotation_in_bound, &test_annotation_start_in_bound, &other_annotation},
			[]*Annotation{},
		},
		{
			"no overlap with",
			pr_main_and_test,
			[]*Annotation{&other_annotation},
			[]*Annotation{},
		},
		{
			"multiple files multiple overlaps",
			pr_main_and_test,
			[]*Annotation{&main_annotation_in_bound, &main_annotation_out_of_bounds, &test_annotation_start_in_bound, &test_annotation_end_in_bound, &test_annotation_out_of_bounds, &other_annotation},
			[]*Annotation{&main_annotation_in_bound, &test_annotation_start_in_bound, &test_annotation_end_in_bound},
		},
		{
			"one file with annotations that match lines but are in other file",
			pr_test,
			[]*Annotation{&main_annotation_in_bound, &other_annotation},
			[]*Annotation{},
		},
		{
			"covering annotation",
			pr_main_and_test,
			[]*Annotation{&main_annotation_covering},
			[]*Annotation{&main_annotation_covering},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pr.filterAnnotations(tt.annotations)

			for _, expectedAnnotation := range tt.filteredAnnotations {
				found := false
				for _, gotAnnotation := range got {
					if *gotAnnotation == *expectedAnnotation {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected annotation %s but did not find it.", expectedAnnotation)
				}
			}

			if len(tt.filteredAnnotations) != len(got) {
				t.Errorf("expected %d annotations but got %d", len(tt.filteredAnnotations), len(got))
			}
		})
	}
}

func TestSdkFilesToInternalFiles(t *testing.T) {
	filename_1 := "src/main.go"
	patch_1 := `@@ -0,0 +1,2 @@
+1
+2`
	bounds_1, _ := patchToLineBounds(patch_1)

	filename_2 := "src/main_2.go"
	patch_2 := `@@ -3,2 +3,2 @@ hi there
-1
+1
+2`
	bounds_2, _ := patchToLineBounds(patch_2)

	github_file_no_patch := github.CommitFile{
		Filename: &filename_1,
	}
	github_file_no_name := github.CommitFile{
		Patch: &patch_1,
	}

	github_file_1 := github.CommitFile{
		Filename: &filename_1,
		Patch:    &patch_1,
	}
	github_file_2 := github.CommitFile{
		Filename: &filename_2,
		Patch:    &patch_2,
	}

	internal_file_1 := pullRequestFile{
		filename:   filename_1,
		patch:      patch_1,
		lineBounds: bounds_1,
	}
	internal_file_2 := pullRequestFile{
		filename:   filename_2,
		patch:      patch_2,
		lineBounds: bounds_2,
	}

	tests := []struct {
		name          string
		sdkFiles      []*github.CommitFile
		internalFiles []*pullRequestFile
	}{
		{
			"nil file",
			[]*github.CommitFile{nil},
			[]*pullRequestFile{},
		},
		{
			"one file",
			[]*github.CommitFile{&github_file_1},
			[]*pullRequestFile{&internal_file_1},
		},
		{
			"multiple files",
			[]*github.CommitFile{&github_file_1, &github_file_2},
			[]*pullRequestFile{&internal_file_1, &internal_file_2},
		},
		{
			"one file without patch",
			[]*github.CommitFile{&github_file_no_patch},
			[]*pullRequestFile{},
		},
		{
			"one file without name",
			[]*github.CommitFile{&github_file_no_name},
			[]*pullRequestFile{},
		},
		{
			"four files, two of which are valid",
			[]*github.CommitFile{&github_file_1, &github_file_2, &github_file_no_name, &github_file_no_patch},
			[]*pullRequestFile{&internal_file_1, &internal_file_2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sdkFilesToInternalFiles(tt.sdkFiles)
			if err != nil {
				t.Errorf("expected no error but received %q", err)
			}

			for _, internalFile := range tt.internalFiles {
				found := false
				for _, foundFile := range got {
					if reflect.DeepEqual(internalFile, foundFile) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected file %s but did not find it: found %s", internalFile, pullRequestFilesString(got))
				}
			}

			if len(got) != len(tt.internalFiles) {
				t.Errorf("expected %d total files but received %d", len(tt.internalFiles), len(got))
			}
		})
	}
}

//go:embed multi_stage.patch
var multiStagePatch string

func TestPatchToLineBounds(t *testing.T) {
	tests := []struct {
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
		{
			"file simulating multiple changes in sequence",
			multiStagePatch,
			[]lineBound{
				{11,14},
				{1,1},
			},
		},
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
