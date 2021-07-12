package buildcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/drone-runners/drone-runner-docker/domain"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone/runner-go/manifest"
	"github.com/earthly/earthly/ast"
	"github.com/earthly/earthly/ast/spec"
	"github.com/earthly/earthly/cleanup"
	"github.com/earthly/earthly/conslogging"
	"github.com/earthly/earthly/util/gitutil"
	"github.com/earthly/earthly/util/llbutil/pllb"
	"github.com/earthly/earthly/util/syncutil/synccache"
	"path/filepath"
	"strings"

	gwclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/pkg/errors"
)

// DockerfileMetaTarget is a target name prefix which signals the resolver that the build file is a
// dockerfile. The DockerfileMetaTarget is really not a valid Earthly target otherwise.
const DockerfileMetaTarget = "@dockerfile:"

// Data represents a resolved target's build context data.
type Data struct {
	// The parsed Earthfile AST.
	Earthfile spec.Earthfile
	// BuildFilePath is the local path where the Earthfile or Dockerfile can be found.
	BuildFilePath string

	// BuildContext is the state to use for the build.
	BuildContext pllb.State
	// GitMetadata contains git metadata information.
	GitMetadata *gitutil.GitMetadata
	// Target is the earthly reference.
	Ref domain.Reference
	// LocalDirs is the local dirs map to be passed as part of the buildkit solve.
	LocalDirs map[string]string
}

// Resolver is a build context resolver.
type Resolver struct {
	gr *gitResolver
	lr *localResolver

	parseCache *synccache.SyncCache // local path -> AST
	console    conslogging.ConsoleLogger
}

// NewResolver returns a new NewResolver.
func NewResolver(sessionID string, cleanCollection *cleanup.Collection, gitLookup *GitLookup, console conslogging.ConsoleLogger) *Resolver {
	return &Resolver{
		gr: &gitResolver{
			cleanCollection: cleanCollection,
			projectCache:    synccache.New(),
			buildFileCache:  synccache.New(),
			gitLookup:       gitLookup,
		},
		lr: &localResolver{
			gitMetaCache: synccache.New(),
			sessionID:    sessionID,
			console:      console,
		},
		parseCache: synccache.New(),
		console:    console,
	}
}

// Resolve returns resolved context data for a given Earthly reference. If the reference is a target,
// then the context will include a build context and possibly additional local directories.
func (r *Resolver) Resolve(ctx context.Context, gwClient gwclient.Client, ref domain.Reference) (*Data, error) {
	if ref.IsUnresolvedImportReference() {
		return nil, errors.Errorf("cannot resolve non-dereferenced import ref %s", ref.String())
	}
	var d *Data
	var err error
	localDirs := make(map[string]string)
	if ref.IsRemote() {
		// Remote.
		d, err = r.gr.resolveEarthProject(ctx, gwClient, ref)
		if err != nil {
			return nil, err
		}
	} else {
		// Local.
		if _, isTarget := ref.(domain.Target); isTarget {
			localDirs[ref.GetLocalPath()] = ref.GetLocalPath()
		}

		d, err = r.lr.resolveLocal(ctx, ref)
		if err != nil {
			return nil, err
		}
	}
	d.Ref = gitutil.ReferenceWithGitMeta(ref, d.GitMetadata)
	d.LocalDirs = localDirs
	target := ref.(domain.Target)
	if !strings.HasPrefix(ref.GetName(), DockerfileMetaTarget) {
		d.Earthfile, err = r.parseTargetJsonfile(ctx, target.TargetASTJson)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

// move drone ci transtlate code
func (r *Resolver) parseEarthfile(ctx context.Context, path string) (spec.Earthfile, error) {
	path = filepath.Clean(path)
	efValue, err := r.parseCache.Do(ctx, path, func(ctx context.Context, k interface{}) (interface{}, error) {
		return ast.Parse(ctx, k.(string), true)
	})
	if err != nil {
		return spec.Earthfile{}, err
	}
	ef := efValue.(spec.Earthfile)

	ef = parseDronefile("engine/compiler/testdata/serial.yml")
	return ef, nil
}

func (r *Resolver) parseTargetJsonfile(ctx context.Context, jsonTarget string) (spec.Earthfile, error) {
	var efile spec.Earthfile
	err := json.Unmarshal([]byte(jsonTarget), &efile)
	return efile, err
}

// Target ==> step
func parseDronefile(source string) spec.Earthfile {
	manifest, _ := manifest.ParseFile(source)
	pipline := manifest.Resources[0].(*resource.Pipeline)

	targets := []spec.Target{}
	for _, step := range pipline.Steps {
		rp := spec.Block{}
		imageCmd := spec.Command{Name: "FROM", Args: []string{step.Image}}
		imageSM := spec.Statement{&imageCmd, nil, nil, nil, nil}
		rp = append(rp, imageSM)

		workDirCmd := spec.Command{Name: "WORKDIR", Args: []string{"/drone-runner-earthly"}} // step.WorkingDir
		workDirSM := spec.Statement{&workDirCmd, nil, nil, nil, nil}
		rp = append(rp, workDirSM)

		// done yaml add copy make drone use as dockerfile way.
		cpCmd := spec.Command{Name: "COPY", Args: []string{"go.mod", "go.sum", "./"}}
		cpSM := spec.Statement{&cpCmd, nil, nil, nil, nil}
		rp = append(rp, cpSM)

		for _, cmd := range step.Commands {
			statement := spec.Statement{&spec.Command{Name: "RUN", Args: []string{cmd}}, nil, nil, nil, nil}
			rp = append(rp, statement)
		}
		target := spec.Target{step.Name, rp, nil}
		targets = append(targets, target)
	}
	efile := spec.Earthfile{nil, nil, targets, nil, nil}
	fmt.Print(efile)

	return efile
}
