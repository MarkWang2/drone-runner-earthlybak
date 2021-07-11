// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"io"
	"os/exec"
)

func main() {
	ExampleCmd_StdinPipe()
	//cmd := exec.Command("./earthly", "+build")
	//cmd.Stdin = strings.NewReader("some input")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Run()
	////if err != nil {
	////	log.Fatal(err)
	////}
	//fmt.Printf("in all caps: %q\n", out.String())

	//./earthly bootstrap
	//command.Command()
	//command.RunApp(nil, nil)
}

func ExampleCmd_StdinPipe() {
	cmd := exec.Command("./earthly", "+build")
	stdin, _ := cmd.StdinPipe()
	//if err != nil {
	//	log.Fatal(err)
	//}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
	}()

	out, _ := cmd.CombinedOutput()

	fmt.Printf("%s\n", out)
}
