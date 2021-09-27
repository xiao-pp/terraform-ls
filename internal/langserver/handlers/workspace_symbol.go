package handlers

import (
	"context"

	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	"github.com/hashicorp/terraform-ls/internal/decoder"
	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

func (h *logHandler) WorkspaceSymbol(ctx context.Context, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	var symbols []lsp.SymbolInformation

	mf, err := lsctx.ModuleFinder(ctx)
	if err != nil {
		return symbols, err
	}

	cc, err := lsctx.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	modules, err := mf.ListModules()
	if err != nil {
		return nil, err
	}

	for _, mod := range modules {
		d, err := decoder.DecoderForModule(ctx, mod)
		if err != nil {
			return symbols, err
		}

		h.logger.Printf("%s: loaded files: %q", mod.Path, d.Filenames())

		schema, _ := mf.SchemaForModule(mod.Path)
		d.SetSchema(schema)

		modSymbols, err := d.Symbols(params.Query)
		if err != nil {
			h.logger.Printf("%s: errors: %s", mod.Path, err)
			continue
		}

		h.logger.Printf("%s: symbols: %#v", mod.Path, modSymbols)

		symbols = append(symbols, ilsp.SymbolInformation(mod.Path, modSymbols,
			cc.Workspace.Symbol)...)
	}

	return symbols, nil
}
