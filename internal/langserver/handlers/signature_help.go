package handlers

import (
	"context"

	ilsp "github.com/hashicorp/terraform-ls/internal/lsp"
	"github.com/hashicorp/terraform-ls/internal/protocol"
	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

func (svc *service) SignatureHelp(ctx context.Context, params lsp.SignatureHelpParams) (*lsp.SignatureHelp, error) {
	_, err := ilsp.ClientCapabilities(ctx)
	if err != nil {
		return nil, err
	}

	dh := ilsp.HandleFromDocumentURI(params.TextDocument.URI)
	doc, err := svc.stateStore.DocumentStore.GetDocument(dh)
	if err != nil {
		return nil, err
	}

	d, err := svc.decoderForDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	pos, err := ilsp.HCLPositionFromLspPosition(params.Position, doc)
	if err != nil {
		return nil, err
	}

	sig, err := d.SignatureAtPos(doc.Filename, pos)
	if err != nil {
		return nil, err
	}
	if sig == nil {
		return nil, nil
	}

	parameters := make([]protocol.ParameterInformation, 0)
	for _, p := range sig.Parameters {
		parameters = append(parameters, lsp.ParameterInformation{
			Label:         p.Name,
			Documentation: p.Description.Value,
		})
	}
	return &lsp.SignatureHelp{
		Signatures: []lsp.SignatureInformation{
			{
				Label:           sig.Name,
				Documentation:   sig.Description.Value,
				Parameters:      parameters,
				ActiveParameter: 1,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: 0,
	}, nil
}
