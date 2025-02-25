package command

import (
	"context"
	"fmt"
	"sort"

	"github.com/creachadair/jrpc2/code"
	"github.com/hashicorp/terraform-ls/internal/langserver/cmd"
	"github.com/hashicorp/terraform-ls/internal/uri"
)

const moduleCallersVersion = 0

type moduleCallersResponse struct {
	FormatVersion int            `json:"v"`
	Callers       []moduleCaller `json:"callers"`
}

type moduleCaller struct {
	URI string `json:"uri"`
}

func (h *CmdHandler) ModuleCallersHandler(ctx context.Context, args cmd.CommandArgs) (interface{}, error) {
	modUri, ok := args.GetString("uri")
	if !ok || modUri == "" {
		return nil, fmt.Errorf("%w: expected module uri argument to be set", code.InvalidParams.Err())
	}

	if !uri.IsURIValid(modUri) {
		return nil, fmt.Errorf("URI %q is not valid", modUri)
	}

	modPath, err := uri.PathFromURI(modUri)
	if err != nil {
		return nil, err
	}

	modCallers, err := h.StateStore.Modules.CallersOfModule(modPath)
	if err != nil {
		return nil, err
	}

	callers := make([]moduleCaller, 0)
	for _, caller := range modCallers {
		callers = append(callers, moduleCaller{
			URI: uri.FromPath(caller.Path),
		})
	}
	sort.SliceStable(callers, func(i, j int) bool {
		return callers[i].URI < callers[j].URI
	})
	return moduleCallersResponse{
		FormatVersion: moduleCallersVersion,
		Callers:       callers,
	}, nil
}
