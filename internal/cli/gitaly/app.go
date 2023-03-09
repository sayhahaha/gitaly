package gitaly

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"gitlab.com/gitlab-org/gitaly/v15/internal/version"
)

func init() {
	// Override the version printer so the output format matches what Gitaly
	// used before the introduction of the CLI toolkit.
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Fprintln(ctx.App.Writer, version.GetVersionString("Gitaly"))
	}
}

// NewApp returns a new gitaly app.
func NewApp() *cli.App {
	return &cli.App{
		Name:    "gitaly",
		Usage:   "a git server",
		Version: version.GetVersionString("Gitaly"),
		// serveAction is also here in the root to keep the CLI backwards compatible
		// with the previous way to launch Gitaly with just `gitaly <configfile>`. We
		// may want to deprecate this eventually.
		Action: serveAction,
		Commands: []*cli.Command{
			newServeCommand(),
			newCheckCommand(),
		},
	}
}
