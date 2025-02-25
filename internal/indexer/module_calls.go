package indexer

import (
	"context"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-ls/internal/document"
	"github.com/hashicorp/terraform-ls/internal/job"
	"github.com/hashicorp/terraform-ls/internal/schemas"
	"github.com/hashicorp/terraform-ls/internal/terraform/module"
	op "github.com/hashicorp/terraform-ls/internal/terraform/module/operation"
)

func (idx *Indexer) decodeInstalledModuleCalls(modHandle document.DirHandle, ignoreState bool) (job.IDs, error) {
	jobIds := make(job.IDs, 0)

	moduleCalls, err := idx.modStore.ModuleCalls(modHandle.Path())
	if err != nil {
		return jobIds, err
	}

	var errs *multierror.Error

	idx.logger.Printf("indexing installed module calls: %d", len(moduleCalls.Installed))
	for _, mc := range moduleCalls.Installed {
		fi, err := os.Stat(mc.Path)
		if err != nil || !fi.IsDir() {
			multierror.Append(errs, err)
			continue
		}
		err = idx.modStore.Add(mc.Path)
		if err != nil {
			multierror.Append(errs, err)
			continue
		}

		mcHandle := document.DirHandleFromPath(mc.Path)
		// copy path for queued jobs below
		mcPath := mc.Path

		refCollectionDeps := make(job.IDs, 0)

		parseId, err := idx.jobStore.EnqueueJob(job.Job{
			Dir: mcHandle,
			Func: func(ctx context.Context) error {
				return module.ParseModuleConfiguration(ctx, idx.fs, idx.modStore, mcPath)
			},
			Type:        op.OpTypeParseModuleConfiguration.String(),
			IgnoreState: ignoreState,
		})
		if err != nil {
			multierror.Append(errs, err)
		} else {
			jobIds = append(jobIds, parseId)
			refCollectionDeps = append(refCollectionDeps, parseId)
		}

		var metaId job.ID
		if parseId != "" {
			metaId, err = idx.jobStore.EnqueueJob(job.Job{
				Dir:  mcHandle,
				Type: op.OpTypeLoadModuleMetadata.String(),
				Func: func(ctx context.Context) error {
					return module.LoadModuleMetadata(ctx, idx.modStore, mcPath)
				},
				DependsOn:   job.IDs{parseId},
				IgnoreState: ignoreState,
			})
			if err != nil {
				multierror.Append(errs, err)
			} else {
				jobIds = append(jobIds, metaId)
				refCollectionDeps = append(refCollectionDeps, metaId)
			}

			eSchemaId, err := idx.jobStore.EnqueueJob(job.Job{
				Dir: mcHandle,
				Func: func(ctx context.Context) error {
					return module.PreloadEmbeddedSchema(ctx, idx.logger, schemas.FS, idx.modStore, idx.schemaStore, mcPath)
				},
				Type:        op.OpTypePreloadEmbeddedSchema.String(),
				DependsOn:   job.IDs{metaId},
				IgnoreState: ignoreState,
			})
			if err != nil {
				multierror.Append(errs, err)
			} else {
				jobIds = append(jobIds, eSchemaId)
				refCollectionDeps = append(refCollectionDeps, eSchemaId)
			}
		}

		if parseId != "" {
			ids, err := idx.collectReferences(mcHandle, refCollectionDeps, ignoreState)
			if err != nil {
				multierror.Append(errs, err)
			} else {
				jobIds = append(jobIds, ids...)
			}
		}

		varsParseId, err := idx.jobStore.EnqueueJob(job.Job{
			Dir: mcHandle,
			Func: func(ctx context.Context) error {
				return module.ParseVariables(ctx, idx.fs, idx.modStore, mcPath)
			},
			Type:        op.OpTypeParseVariables.String(),
			IgnoreState: ignoreState,
		})
		if err != nil {
			multierror.Append(errs, err)
		} else {
			jobIds = append(jobIds, varsParseId)
		}

		if varsParseId != "" {
			varsRefId, err := idx.jobStore.EnqueueJob(job.Job{
				Dir: mcHandle,
				Func: func(ctx context.Context) error {
					return module.DecodeVarsReferences(ctx, idx.modStore, idx.schemaStore, mcPath)
				},
				Type:        op.OpTypeDecodeVarsReferences.String(),
				DependsOn:   job.IDs{varsParseId},
				IgnoreState: ignoreState,
			})
			if err != nil {
				multierror.Append(errs, err)
			} else {
				jobIds = append(jobIds, varsRefId)
			}
		}
	}

	return jobIds, errs.ErrorOrNil()
}

func (idx *Indexer) collectReferences(modHandle document.DirHandle, dependsOn job.IDs, ignoreState bool) (job.IDs, error) {
	ids := make(job.IDs, 0)

	var errs *multierror.Error

	id, err := idx.jobStore.EnqueueJob(job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return module.DecodeReferenceTargets(ctx, idx.modStore, idx.schemaStore, modHandle.Path())
		},
		Type:        op.OpTypeDecodeReferenceTargets.String(),
		DependsOn:   dependsOn,
		IgnoreState: ignoreState,
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, id)
	}

	id, err = idx.jobStore.EnqueueJob(job.Job{
		Dir: modHandle,
		Func: func(ctx context.Context) error {
			return module.DecodeReferenceOrigins(ctx, idx.modStore, idx.schemaStore, modHandle.Path())
		},
		Type:        op.OpTypeDecodeReferenceOrigins.String(),
		DependsOn:   dependsOn,
		IgnoreState: ignoreState,
	})
	if err != nil {
		errs = multierror.Append(errs, err)
	} else {
		ids = append(ids, id)
	}

	return ids, errs.ErrorOrNil()
}
