package lsp

import (
	"github.com/hashicorp/hcl-lang/lang"
	lsp "github.com/hashicorp/terraform-ls/internal/protocol"
)

func ToSignatureHelp(signature *lang.FuncSignature) *lsp.SignatureHelp {
	if signature == nil {
		return nil
	}

	parameters := make([]lsp.ParameterInformation, 0)
	for _, p := range signature.Parameters {
		parameters = append(parameters, lsp.ParameterInformation{
			Label:         p.Name,
			Documentation: p.Description.Value, // TODO? clean
		})
	}

	return &lsp.SignatureHelp{
		Signatures: []lsp.SignatureInformation{
			{
				Label:           signature.Name,
				Documentation:   signature.Description.Value, // TODO? clean
				Parameters:      parameters,
				ActiveParameter: signature.ActiveParameter,
			},
		},
		ActiveSignature: 0,
	}
}
