package foundrygen

import (
	"os"
	"path/filepath"
)

// checkPathAliases complements the project file's lexical containment checks.
// Filesystems may resolve different spellings to the same existing object.
func (s *session) checkPathAliases() error {
	type namedPath struct{ name, path string }
	inputs := []namedPath{{ProjectFile, filepath.Join(s.root, ProjectFile)}}
	outputs := make([]namedPath, len(s.proj.Outputs))
	for i, o := range s.proj.Outputs {
		model, out, err := s.paths(o)
		if err != nil {
			return err
		}
		inputs = append(inputs, namedPath{o.Model, model})
		outputs[i] = namedPath{o.Out, out}
	}

	cache := map[string]os.FileInfo{}
	inputAncestors := make([][]os.FileInfo, len(inputs))
	outputAncestors := make([][]os.FileInfo, len(outputs))
	for _, group := range []struct {
		paths     []namedPath
		ancestors [][]os.FileInfo
	}{
		{inputs, inputAncestors}, {outputs, outputAncestors},
	} {
		for i, p := range group.paths {
			ancestors, err := pathAncestors(s.root, p.path, cache)
			if err != nil {
				return err
			}
			group.ancestors[i] = ancestors
		}
	}
	for i, out := range outputs {
		info := cache[out.path]
		if info == nil {
			continue
		}
		contains := func(ancestors []os.FileInfo) bool {
			for _, ancestor := range ancestors {
				if os.SameFile(info, ancestor) {
					return true
				}
			}
			return false
		}
		for j, input := range inputs {
			if contains(inputAncestors[j]) {
				return refuse("output %q aliases input %q or one of its parent directories", out.name, input.name)
			}
		}
		for j, other := range outputs {
			if i != j && contains(outputAncestors[j]) {
				return refuse("output %q aliases output %q or one of its parent directories", out.name, other.name)
			}
		}
	}
	return nil
}

// Include existing parents even when the final path does not yet exist. A nil
// cache entry records an absent path; no probing files or directories are made.
func pathAncestors(root, path string, cache map[string]os.FileInfo) ([]os.FileInfo, error) {
	var ancestors []os.FileInfo
	for {
		info, known := cache[path]
		if !known {
			var err error
			info, err = os.Lstat(path)
			if err != nil && !os.IsNotExist(err) {
				return nil, fail("inspect %s: %v", path, err)
			}
			cache[path] = info
		}
		if info != nil {
			ancestors = append(ancestors, info)
		}
		if path == root {
			return ancestors, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return nil, fail("%s is outside project %s", path, root)
		}
		path = parent
	}
}
