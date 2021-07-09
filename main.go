// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/compiler"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
	_ "github.com/joho/godotenv/autoload"
	"io/ioutil"
)

func main() {
	//command.Command()
	testCompile("engine/compiler/testdata/serial.yml", "engine/compiler/testdata/serial.json")
}

// helper function parses and compiles the source file and then
// compares to a golden json file.
func testCompile(source, golden string) *engine.Spec {
	// replace the default random function with one that
	// is deterministic, for testing purposes.

	manifest, _ := manifest.ParseFile(source)

	compiler := &compiler.Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{Target: "master"},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{Machine: "github.com", Login: "octocat", Password: "correct-horse-battery-staple"},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}
	var nocontext = context.Background()
	got := compiler.Compile(nocontext, args)

	raw, _ := ioutil.ReadFile(golden)

	want := new(engine.Spec)
	json.Unmarshal(raw, want)
	//err.(*exec.ExitError)
	bb := got.(*engine.Spec)

	config := engine.ToConfig(bb, bb.Steps[1])
	fmt.Print(config)
	return got.(*engine.Spec)
}
