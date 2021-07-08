// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"github.com/earthly/earthly/buildkitd"
	"github.com/earthly/earthly/conslogging"
	"github.com/earthly/earthly/domain"
	"github.com/earthly/earthly/util/llbutil"
	"github.com/earthly/earthly/util/llbutil/pllb"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"golang.org/x/sync/errgroup"
	"io"
	// "github.com/earthly/earthly/util/llbutil/pllb"
	_ "github.com/joho/godotenv/autoload"
)

//func (app *earthlyApp) actionBuildImp(c *cli.Context, flagArgs, nonFlagArgs []string) error {
//	var target domain.Target
//	var artifact domain.Artifact
//	destPath := "./"
//	target, _ = domain.ParseTarget("targetName")
//
//	bkClient, err := buildkitd.NewClient(c.Context, app.console, app.buildkitdImage, app.buildkitdSettings)
//	if err != nil {
//		return errors.Wrap(err, "build new buildkitd client")
//	}
//	defer bkClient.Close()
//	isLocal := buildkitd.IsLocal(app.buildkitdSettings.BuildkitAddress)
//
//	bkIP, err := buildkitd.GetContainerIP(c.Context, app.buildkitdSettings)
//	if err != nil {
//		return errors.Wrap(err, "get buildkit container IP")
//	}
//
//	platformsSlice := make([]*specs.Platform, 0, len(app.platformsStr.Value()))
//	for _, p := range app.platformsStr.Value() {
//		platform, err := llbutil.ParsePlatform(p)
//		if err != nil {
//			return errors.Wrapf(err, "parse platform %s", p)
//		}
//		platformsSlice = append(platformsSlice, platform)
//	}
//	if len(platformsSlice) == 0 {
//		platformsSlice = []*specs.Platform{nil}
//	}
//
//	dotEnvMap := make(map[string]string)
//	if fileutil.FileExists(dotEnvPath) {
//		dotEnvMap, err = godotenv.Read(dotEnvPath)
//		if err != nil {
//			return errors.Wrapf(err, "read %s", dotEnvPath)
//		}
//	}
//	secretsMap, err := processSecrets(app.secrets.Value(), app.secretFiles.Value(), dotEnvMap)
//	if err != nil {
//		return err
//	}
//
//	debuggerSettings := debuggercommon.DebuggerSettings{
//		DebugLevelLogging: app.debug,
//		Enabled:           app.interactiveDebugging,
//		RepeaterAddr:      fmt.Sprintf("%s:8373", bkIP),
//		Term:              os.Getenv("TERM"),
//	}
//
//	debuggerSettingsData, err := json.Marshal(&debuggerSettings)
//	if err != nil {
//		return errors.Wrap(err, "debugger settings json marshal")
//	}
//	secretsMap[debuggercommon.DebuggerSettingsSecretsKey] = debuggerSettingsData
//
//	sc, err := secretsclient.NewClient(app.apiServer, app.sshAuthSock, app.authToken, app.console.Warnf)
//	if err != nil {
//		return errors.Wrap(err, "failed to create secretsclient")
//	}
//
//	localhostProvider, err := localhostprovider.NewLocalhostProvider()
//	if err != nil {
//		return errors.Wrap(err, "failed to create localhostprovider")
//	}
//
//	cacheLocalDir, err := ioutil.TempDir("", "earthly-cache")
//	if err != nil {
//		return errors.Wrap(err, "make temp dir for cache")
//	}
//	defer os.RemoveAll(cacheLocalDir)
//	defaultLocalDirs := make(map[string]string)
//	defaultLocalDirs["earthly-cache"] = cacheLocalDir
//	buildContextProvider := provider.NewBuildContextProvider(app.console)
//	buildContextProvider.AddDirs(defaultLocalDirs)
//	attachables := []session.Attachable{
//		llbutil.NewSecretProvider(sc, secretsMap),
//		authprovider.NewDockerAuthProvider(os.Stderr),
//		buildContextProvider,
//		localhostProvider,
//	}
//
//	gitLookup := buildcontext.NewGitLookup(app.console, app.sshAuthSock)
//	err = app.updateGitLookupConfig(gitLookup)
//	if err != nil {
//		return err
//	}
//
//	if app.sshAuthSock != "" {
//		ssh, err := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{{
//			Paths: []string{app.sshAuthSock},
//		}})
//		if err != nil {
//			return errors.Wrap(err, "ssh agent provider")
//		}
//		attachables = append(attachables, ssh)
//	}
//
//	var enttlmnts []entitlements.Entitlement
//	if app.allowPrivileged {
//		enttlmnts = append(enttlmnts, entitlements.EntitlementSecurityInsecure)
//	}
//	cleanCollection := cleanup.NewCollection()
//	defer cleanCollection.Close()
//
//	go func() {
//		// Dialing doesnt accept URLs, it accepts an address and a "network". These cannot be handled as URL schemes.
//		// Since Shellrepeater hard-codes TCP, we drop it here and log the error if we fail to connect.
//
//		u, err := url.Parse(app.debuggerHost)
//		if err != nil {
//			panic("debugger host was not a URL")
//		}
//
//		debugTermConsole := app.console.WithPrefix("internal-term")
//		err = terminal.ConnectTerm(c.Context, u.Host, debugTermConsole)
//		if err != nil {
//			debugTermConsole.Warnf("Failed to connect to terminal: %s", err.Error())
//		}
//	}()
//
//	dotEnvVars := variables.NewScope()
//	for k, v := range dotEnvMap {
//		dotEnvVars.AddInactive(k, v)
//	}
//	buildArgs := append([]string{}, app.buildArgs.Value()...)
//	buildArgs = append(buildArgs, flagArgs...)
//	overridingVars, err := variables.ParseCommandLineArgs(buildArgs)
//	if err != nil {
//		return errors.Wrap(err, "parse build args")
//	}
//	overridingVars = variables.CombineScopes(overridingVars, dotEnvVars)
//	imageResolveMode := llb.ResolveModePreferLocal
//	if app.pull {
//		imageResolveMode = llb.ResolveModeForcePull
//	}
//
//	cacheImports := make(map[string]bool)
//	if app.remoteCache != "" {
//		cacheImports[app.remoteCache] = true
//	}
//	var cacheExport string
//	var maxCacheExport string
//	if app.remoteCache != "" && app.push {
//		if app.maxRemoteCache {
//			maxCacheExport = app.remoteCache
//		} else {
//			cacheExport = app.remoteCache
//		}
//	}
//	var parallelism *semaphore.Weighted
//	if app.conversionParllelism != 0 {
//		parallelism = semaphore.NewWeighted(int64(app.conversionParllelism))
//	}
//	localRegistryAddr := ""
//	if isLocal && app.cfg.Global.LocalRegistryHost != "" {
//		lrURL, err := url.Parse(app.cfg.Global.LocalRegistryHost)
//		if err != nil {
//			return errors.Wrapf(err, "parse local registry host %s", app.cfg.Global.LocalRegistryHost)
//		}
//		localRegistryAddr = lrURL.Host
//	}
//	builderOpts := builder.Opt{
//		BkClient:               bkClient,
//		Console:                app.console,
//		Verbose:                app.verbose,
//		Attachables:            attachables,
//		Enttlmnts:              enttlmnts,
//		NoCache:                app.noCache,
//		CacheImports:           states.NewCacheImports(cacheImports),
//		CacheExport:            cacheExport,
//		MaxCacheExport:         maxCacheExport,
//		UseInlineCache:         app.useInlineCache,
//		SaveInlineCache:        app.saveInlineCache,
//		SessionID:              app.sessionID,
//		ImageResolveMode:       imageResolveMode,
//		CleanCollection:        cleanCollection,
//		OverridingVars:         overridingVars,
//		BuildContextProvider:   buildContextProvider,
//		GitLookup:              gitLookup,
//		UseFakeDep:             !app.noFakeDep,
//		Strict:                 app.strict,
//		DisableNoOutputUpdates: app.interactiveDebugging,
//		ParallelConversion:     (app.conversionParllelism != 0),
//		Parallelism:            parallelism,
//		LocalRegistryAddr:      localRegistryAddr,
//	}
//	b, err := builder.NewBuilder(c.Context, builderOpts)
//	if err != nil {
//		return errors.Wrap(err, "new builder")
//	}
//
//	if len(platformsSlice) != 1 {
//		return errors.Errorf("multi-platform builds are not yet supported on the command line. You may, however, create a target with the instruction BUILD --plaform ... --platform ... %s", target)
//	}
//	buildOpts := builder.BuildOpt{
//		PrintSuccess:          true,
//		Push:                  app.push,
//		NoOutput:              app.noOutput,
//		OnlyFinalTargetImages: app.imageMode,
//		Platform:              platformsSlice[0],
//
//		// explicitly set this to true at the top level (without granting the entitlements.EntitlementSecurityInsecure buildkit option),
//		// to differentiate between a user forgetting to run earthly -P, versus a remotely referening an earthfile that requires privileged.
//		AllowPrivileged: true,
//	}
//	if app.artifactMode {
//		buildOpts.OnlyArtifact = &artifact
//		buildOpts.OnlyArtifactDestPath = destPath
//	}
//	_, err = b.BuildTarget(c.Context, target, buildOpts)
//	if err != nil {
//		return errors.Wrap(err, "build target")
//	}
//	return nil
//}

func main() {
	//command.Command()
	//targets, _ := earthfile2llb.GetTargets("Earthfile")
	target, _ := domain.ParseTarget("+build")
	fmt.Print(target)
	//fmt.Print(targets)

	console := conslogging.Current(conslogging.ForceColor, conslogging.DefaultPadding, false)
	// Bootstrap buildkit - pulls image and starts daemon.
	//ctx, _ := context.WithTimeout(ctx, 100*time.Millisecond)
	buildkitdImage := "earthly/buildkitd:main"
	ctx := context.Background()
	bkClient, _ := buildkitd.NewClient(ctx, console, buildkitdImage, buildkitd.Settings{BuildkitAddress: "docker-container://earthly-buildkitd", DebuggerAddress: "tcp://127.0.0.1:8373", LocalRegistryAddress: "tcp://127.0.0.1:8371", UseTCP: false, UseTLS: false})
	state := pllb.Image("earthly/buildkitd:main", llb.Platform(llbutil.DefaultPlatform()))

	dt, _ := state.Marshal(ctx, llb.Platform(llbutil.DefaultPlatform()))
	solveOpt := &client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterDocker,
				Attrs: map[string]string{
					"name":                  "",
					"containerimage.config": "",
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return nil, nil
				},
			},
		},
	}
	ch := make(chan *client.SolveStatus)
	con := conslogging.Current(conslogging.ForceColor, conslogging.DefaultPadding, false)
	sm := newSolverMonitor(con, true, true)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		_, _ = bkClient.Solve(ctx, dt, *solveOpt, ch)
		return nil
	})
	sm.PrintTiming()
	var vertexFailureOutput string

	eg.Go(func() error {
		var err error
		vertexFailureOutput, err = sm.monitorProgress(ctx, ch, "", true)
		return err
	})

	eg.Wait()
	fmt.Print(vertexFailureOutput)
	fmt.Print(bkClient)
}
