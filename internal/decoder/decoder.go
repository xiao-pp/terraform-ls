package decoder

import (
	"context"
	"log"

	"github.com/hashicorp/hcl-lang/decoder"
	"github.com/hashicorp/hcl-lang/lang"
	lsctx "github.com/hashicorp/terraform-ls/internal/context"
	"github.com/hashicorp/terraform-ls/internal/terraform/ast"
	"github.com/hashicorp/terraform-ls/internal/terraform/module"
)

func DecoderForModule(ctx context.Context, mod module.Module) (*decoder.Decoder, error) {
	d := decoder.NewDecoder()

	d.SetReferenceTargetReader(func() lang.ReferenceTargets {
		return mod.RefTargets
	})

	d.SetReferenceOriginReader(func() lang.ReferenceOrigins {
		return mod.RefOrigins
	})

	d.SetUtmSource("terraform-ls")
	d.UseUtmContent(true)

	clientName, ok := lsctx.ClientName(ctx)
	if ok {
		d.SetUtmMedium(clientName)
	}

	for name, f := range mod.ParsedModuleFiles {
		err := d.LoadFile(name.String(), f)
		if err != nil {
			// skip unreadable files
			continue
		}
	}

	return d, nil
}

func DecoderForVariables(varsFiles ast.VarsFiles, logger *log.Logger) (*decoder.Decoder, error) {
	d := decoder.NewDecoder()

	logger.Printf("loading %d varfiles", len(varsFiles))
	for name, f := range varsFiles {
		logger.Printf("loading varfile: %q", name)
		err := d.LoadFile(name.String(), f)
		if err != nil {
			// skip unreadable files
			// return nil, err
			logger.Printf("skipping varfile %q: %s", name, err)
			continue
		} else {
			logger.Printf("loaded varfile %q", name)
		}
	}

	return d, nil
}
