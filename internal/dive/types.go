// Package dive wraps invocations of the wagoodman/dive CLI and parses its
// JSON analysis output.
package dive

import "encoding/json"

// Analysis is the top-level structure written by `dive <image> --json <file>`.
type Analysis struct {
	Layers []Layer   `json:"layer"`
	Image  ImageInfo `json:"image"`
}

// Layer describes a single image layer as reported by dive.
type Layer struct {
	Index     int    `json:"index"`
	ID        string `json:"id"`
	DigestID  string `json:"digestId"`
	SizeBytes int64  `json:"sizeBytes"`
	Command   string `json:"command"`
	// FileList holds the (large) per-file tree for the layer. It is not
	// needed by any of the dive-mcp tools, so it is captured loosely and
	// never unmarshalled into a rigid struct.
	FileList []json.RawMessage `json:"fileList,omitempty"`
}

// ImageInfo summarizes the whole image, including candidate wasted files.
type ImageInfo struct {
	SizeBytes        int64           `json:"sizeBytes"`
	InefficientBytes int64           `json:"inefficientBytes"`
	EfficiencyScore  float64         `json:"efficiencyScore"`
	FileReference    []FileReference `json:"fileReference"`
}

// FileReference is a file that appears in more than one layer (a candidate
// for wasted space), along with how many times and how many bytes it costs.
type FileReference struct {
	Count     int    `json:"count"`
	SizeBytes int64  `json:"sizeBytes"`
	File      string `json:"file"`
}

// TotalWastedBytes returns Count * SizeBytes, i.e. the total bytes
// attributable to this file across all of its duplicated occurrences.
func (f FileReference) TotalWastedBytes() int64 {
	return int64(f.Count) * f.SizeBytes
}

// TopWasted returns up to limit FileReferences sorted in descending order of
// TotalWastedBytes. The input slice is not mutated.
func TopWasted(refs []FileReference, limit int) []FileReference {
	sorted := make([]FileReference, len(refs))
	copy(sorted, refs)

	// simple insertion sort (descending) -- input sizes are small (dozens to
	// low hundreds of entries), so this avoids pulling in sort for a trivial
	// comparator while still being O(n^2) worst case, which is fine here.
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].TotalWastedBytes() > sorted[j-1].TotalWastedBytes(); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	if limit <= 0 || limit > len(sorted) {
		return sorted
	}
	return sorted[:limit]
}
