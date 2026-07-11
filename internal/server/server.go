// Package server registers dive-mcp's MCP tools on an mcp.Server.
package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codenio/dive-mcp/internal/dive"
)

// New creates an mcp.Server named "dive-mcp" with all dive tools registered.
func New(version string, runner *dive.Runner) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "dive-mcp",
		Version: version,
	}, nil)

	registerAnalyzeImage(s, runner)
	registerListLayers(s, runner)
	registerGetWastedSpace(s, runner)
	registerCICheck(s, runner)

	return s
}

// ---- analyze_image ----

type analyzeImageInput struct {
	Image  string `json:"image" jsonschema:"the container image reference to analyze, e.g. alpine:latest or myrepo/myimage:tag"`
	Source string `json:"source,omitempty" jsonschema:"the container engine to fetch the image from: docker, podman, or docker-archive (default: docker)"`
}

type analyzeImageOutput struct {
	Image            string  `json:"image"`
	SizeBytes        int64   `json:"sizeBytes"`
	InefficientBytes int64   `json:"inefficientBytes"`
	EfficiencyScore  float64 `json:"efficiencyScore"`
	LayerCount       int     `json:"layerCount"`
	Summary          string  `json:"summary"`
}

func registerAnalyzeImage(s *mcp.Server, runner *dive.Runner) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "analyze_image",
		Description: "Analyze a container image with dive and return a summary: total size, " +
			"efficiency score (0-1, higher is better), inefficient (wasted) bytes, and layer count.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in analyzeImageInput) (*mcp.CallToolResult, analyzeImageOutput, error) {
		if in.Image == "" {
			return nil, analyzeImageOutput{}, fmt.Errorf("image must not be empty")
		}
		a, err := runner.Analyze(ctx, in.Image, in.Source)
		if err != nil {
			return nil, analyzeImageOutput{}, err
		}
		out := analyzeImageOutput{
			Image:            in.Image,
			SizeBytes:        a.Image.SizeBytes,
			InefficientBytes: a.Image.InefficientBytes,
			EfficiencyScore:  a.Image.EfficiencyScore,
			LayerCount:       len(a.Layers),
		}
		out.Summary = fmt.Sprintf(
			"Image %q: %s across %d layers, efficiency score %.4f, %s inefficient (wasted).",
			in.Image, humanBytes(out.SizeBytes), out.LayerCount, out.EfficiencyScore, humanBytes(out.InefficientBytes),
		)
		return nil, out, nil
	})
}

// ---- list_layers ----

type listLayersInput struct {
	Image  string `json:"image" jsonschema:"the container image reference to analyze, e.g. alpine:latest"`
	Source string `json:"source,omitempty" jsonschema:"the container engine to fetch the image from: docker, podman, or docker-archive (default: docker)"`
}

type layerOutput struct {
	Index     int    `json:"index"`
	DigestID  string `json:"digestId"`
	SizeBytes int64  `json:"sizeBytes"`
	Command   string `json:"command"`
}

type listLayersOutput struct {
	Image  string        `json:"image"`
	Layers []layerOutput `json:"layers"`
}

func registerListLayers(s *mcp.Server, runner *dive.Runner) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_layers",
		Description: "List every layer of a container image with its index, digest, size, and the command that produced it.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listLayersInput) (*mcp.CallToolResult, listLayersOutput, error) {
		if in.Image == "" {
			return nil, listLayersOutput{}, fmt.Errorf("image must not be empty")
		}
		a, err := runner.Analyze(ctx, in.Image, in.Source)
		if err != nil {
			return nil, listLayersOutput{}, err
		}
		out := listLayersOutput{Image: in.Image}
		for _, l := range a.Layers {
			out.Layers = append(out.Layers, layerOutput{
				Index:     l.Index,
				DigestID:  l.DigestID,
				SizeBytes: l.SizeBytes,
				Command:   l.Command,
			})
		}
		return nil, out, nil
	})
}

// ---- get_wasted_space ----

type getWastedSpaceInput struct {
	Image  string `json:"image" jsonschema:"the container image reference to analyze, e.g. alpine:latest"`
	Source string `json:"source,omitempty" jsonschema:"the container engine to fetch the image from: docker, podman, or docker-archive (default: docker)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum number of wasted file entries to return, sorted by total wasted bytes descending (default: 20)"`
}

type wastedFileOutput struct {
	File             string `json:"file"`
	Count            int    `json:"count"`
	SizeBytes        int64  `json:"sizeBytes"`
	TotalWastedBytes int64  `json:"totalWastedBytes"`
}

type getWastedSpaceOutput struct {
	Image            string             `json:"image"`
	InefficientBytes int64              `json:"inefficientBytes"`
	Files            []wastedFileOutput `json:"files"`
}

func registerGetWastedSpace(s *mcp.Server, runner *dive.Runner) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "get_wasted_space",
		Description: "Return the top files that waste space in a container image (files duplicated across layers), " +
			"sorted by total wasted bytes (count * size) descending.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getWastedSpaceInput) (*mcp.CallToolResult, getWastedSpaceOutput, error) {
		if in.Image == "" {
			return nil, getWastedSpaceOutput{}, fmt.Errorf("image must not be empty")
		}
		a, err := runner.Analyze(ctx, in.Image, in.Source)
		if err != nil {
			return nil, getWastedSpaceOutput{}, err
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 20
		}
		top := dive.TopWasted(a.Image.FileReference, limit)
		out := getWastedSpaceOutput{
			Image:            in.Image,
			InefficientBytes: a.Image.InefficientBytes,
		}
		for _, f := range top {
			out.Files = append(out.Files, wastedFileOutput{
				File:             f.File,
				Count:            f.Count,
				SizeBytes:        f.SizeBytes,
				TotalWastedBytes: f.TotalWastedBytes(),
			})
		}
		return nil, out, nil
	})
}

// ---- ci_check ----

type ciCheckInput struct {
	Image                    string  `json:"image" jsonschema:"the container image reference to analyze, e.g. alpine:latest"`
	Source                   string  `json:"source,omitempty" jsonschema:"the container engine to fetch the image from: docker, podman, or docker-archive (default: docker)"`
	LowestEfficiency         float64 `json:"lowest_efficiency,omitempty" jsonschema:"minimum allowed efficiency score, 0-1 (default: 0.9)"`
	HighestWastedBytes       int64   `json:"highest_wasted_bytes,omitempty" jsonschema:"maximum allowed inefficient (wasted) bytes; 0 disables this check (default: disabled)"`
	HighestUserWastedPercent float64 `json:"highest_user_wasted_percent,omitempty" jsonschema:"maximum allowed fraction of image size that is wasted, 0-1 (default: 0.1)"`
}

type ciCheckOutput struct {
	Image            string   `json:"image"`
	Pass             bool     `json:"pass"`
	Reasons          []string `json:"reasons,omitempty"`
	EfficiencyScore  float64  `json:"efficiencyScore"`
	InefficientBytes int64    `json:"inefficientBytes"`
	WastedPercent    float64  `json:"wastedPercent"`
}

func registerCICheck(s *mcp.Server, runner *dive.Runner) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "ci_check",
		Description: "Validate a container image against dive's CI-style thresholds: lowest allowed efficiency score, " +
			"highest allowed wasted bytes, and highest allowed wasted percentage. Mirrors `dive --ci` semantics but " +
			"returns structured pass/fail data instead of an exit code.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ciCheckInput) (*mcp.CallToolResult, ciCheckOutput, error) {
		if in.Image == "" {
			return nil, ciCheckOutput{}, fmt.Errorf("image must not be empty")
		}
		a, err := runner.Analyze(ctx, in.Image, in.Source)
		if err != nil {
			return nil, ciCheckOutput{}, err
		}

		lowestEfficiency := in.LowestEfficiency
		if lowestEfficiency == 0 {
			lowestEfficiency = 0.9
		}
		highestUserWastedPercent := in.HighestUserWastedPercent
		if highestUserWastedPercent == 0 {
			highestUserWastedPercent = 0.1
		}
		highestWastedBytes := in.HighestWastedBytes // 0 == disabled

		var wastedPercent float64
		if a.Image.SizeBytes > 0 {
			wastedPercent = float64(a.Image.InefficientBytes) / float64(a.Image.SizeBytes)
		}

		out := ciCheckOutput{
			Image:            in.Image,
			Pass:             true,
			EfficiencyScore:  a.Image.EfficiencyScore,
			InefficientBytes: a.Image.InefficientBytes,
			WastedPercent:    wastedPercent,
		}

		if a.Image.EfficiencyScore < lowestEfficiency {
			out.Pass = false
			out.Reasons = append(out.Reasons, fmt.Sprintf(
				"efficiency score %.4f is below the lowest allowed efficiency %.4f", a.Image.EfficiencyScore, lowestEfficiency))
		}
		if highestWastedBytes > 0 && a.Image.InefficientBytes > highestWastedBytes {
			out.Pass = false
			out.Reasons = append(out.Reasons, fmt.Sprintf(
				"inefficient bytes %d exceed the highest allowed wasted bytes %d", a.Image.InefficientBytes, highestWastedBytes))
		}
		if wastedPercent > highestUserWastedPercent {
			out.Pass = false
			out.Reasons = append(out.Reasons, fmt.Sprintf(
				"wasted percent %.4f exceeds the highest allowed user wasted percent %.4f", wastedPercent, highestUserWastedPercent))
		}

		return nil, out, nil
	})
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
