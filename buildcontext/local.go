package buildcontext

import (
	"context"
	"github.com/drone-runners/drone-runner-docker/analytics"
	"github.com/drone-runners/drone-runner-docker/conslogging"
	"github.com/drone-runners/drone-runner-docker/domain"
	"github.com/drone-runners/drone-runner-docker/util/gitutil"
	"github.com/drone-runners/drone-runner-docker/util/llbutil"
	"github.com/drone-runners/drone-runner-docker/util/llbutil/pllb"
	"github.com/drone-runners/drone-runner-docker/util/syncutil/synccache"
	"github.com/moby/buildkit/client/llb"
	"github.com/pkg/errors"
)

type localResolver struct {
	gitMetaCache *synccache.SyncCache // local path -> *gitutil.GitMetadata
	sessionID    string
	console      conslogging.ConsoleLogger
}

func (lr *localResolver) resolveLocal(ctx context.Context, ref domain.Reference) (*Data, error) {
	analytics.Count("localResolver.resolveLocal", "local-reference")
	if ref.IsRemote() {
		return nil, errors.Errorf("unexpected remote target %s", ref.String())
	}

	metadataValue, err := lr.gitMetaCache.Do(ctx, ref.GetLocalPath(), func(ctx context.Context, _ interface{}) (interface{}, error) {
		metadata, err := gitutil.Metadata(ctx, ref.GetLocalPath())
		if err != nil {
			if errors.Is(err, gitutil.ErrNoGitBinary) ||
				errors.Is(err, gitutil.ErrNotAGitDir) ||
				errors.Is(err, gitutil.ErrCouldNotDetectRemote) ||
				errors.Is(err, gitutil.ErrCouldNotDetectGitHash) ||
				errors.Is(err, gitutil.ErrCouldNotDetectGitBranch) {
				// Keep going anyway. Either not a git dir, or git not installed, or
				// remote not detected.
				if errors.Is(err, gitutil.ErrNoGitBinary) {
					lr.console.Warnf("Warning: %s\n", err.Error())
				}
			} else {
				return nil, err
			}
		}
		return metadata, nil
	})
	if err != nil {
		return nil, err
	}
	metadata := metadataValue.(*gitutil.GitMetadata)
	var buildFilePath string
	//target := ref.(domain.Target)
	//if !target.FromArgs {
	//	buildFilePath, err = detectBuildFile(ref, filepath.FromSlash(ref.GetLocalPath()))
	//}
	if err != nil {
		return nil, err
	}

	var buildContext pllb.State
	if _, isTarget := ref.(domain.Target); isTarget {
		excludes, err := readExcludes(ref.GetLocalPath())
		if err != nil {
			return nil, err
		}
		buildContext = pllb.Local(
			ref.GetLocalPath(),
			llb.ExcludePatterns(excludes),
			llb.SessionID(lr.sessionID),
			llb.Platform(llbutil.DefaultPlatform()),
			llb.WithCustomNamef("[context %s] local context %s", ref.GetLocalPath(), ref.GetLocalPath()),
		)
	} else {
		// Commands don't come with a build context.
	}

	return &Data{
		BuildFilePath: buildFilePath,
		BuildContext:  buildContext,
		GitMetadata:   metadata,
	}, nil
}
