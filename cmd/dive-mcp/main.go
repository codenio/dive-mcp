// Command dive-mcp is an MCP (Model Context Protocol) stdio server that
// wraps the dive CLI (github.com/wagoodman/dive) so AI tools can analyze
// container image layers and wasted space.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codenio/dive-mcp/internal/dive"
	"github.com/codenio/dive-mcp/internal/server"
)

// version is set via -ldflags "-X main.version=..." at release build time.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("dive-mcp " + version)
		return
	}

	runner, err := dive.NewRunner()
	if err != nil {
		log.Fatalf("dive-mcp: %v", err)
	}

	s := server.New(version, runner)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "dive-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}
