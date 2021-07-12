//// Copyright 2019 Drone.IO Inc. All rights reserved.
//// Use of this source code is governed by the Polyform License
//// that can be found in the LICENSE file.
//
//package main
//
//import (
//	"bufio"
//	"fmt"
//	_ "github.com/joho/godotenv/autoload"
//	"io"
//	"log"
//	"os/exec"
//	"strconv"
//	"time"
//)
//
////func main() {
////	//ExampleCmd_StdinPipe()
////	cmd := exec.Command("./earthly", "+build333")
////	cmdReader, err := cmd.StdoutPipe()
////	if err != nil {
////		log.Fatal(err)
////	}
////	scanner := bufio.NewScanner(cmdReader)
////	go func() {
////		for scanner.Scan() {
////			fmt.Println(scanner.Text())
////		}
////	}()
////	if err := cmd.Start(); err != nil {
////		log.Fatal(err)
////	}
////	if err := cmd.Wait(); err != nil {
////		log.Fatal(err)
////	}
////	//cmd.Stdin = strings.NewReader("some input")
////	//var out bytes.Buffer
////	//cmd.Stdout = &out
////	//cmd.Run()
////	////if err != nil {
////	////	log.Fatal(err)
////	////}
////	//fmt.Printf("in all caps: %q\n", out.String())
////
////	//./earthly bootstrap
////	//command.Command()
////	//command.RunApp(nil, nil)
////}
//
//func ExampleCmd_StdinPipe() {
//	cmd := exec.Command("./earthly", "+build")
//	stdin, _ := cmd.StdinPipe()
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	go func() {
//		defer stdin.Close()
//		io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
//	}()
//
//	out, _ := cmd.CombinedOutput()
//
//	fmt.Printf("%s\n", out)
//}
//
//
//func main() {
//	for i := 10; i < 20; i++ {
//		go printName(`My name is Bob, I am ` + strconv.Itoa(i) + ` years old`)
//		// Adding delay so as to see incremental output
//		time.Sleep(60 * time.Millisecond)
//	}
//	// Adding delay so as to let program complete
//	// Please use channels or wait groups
//	time.Sleep(100 * time.Millisecond)
//}
////
//func printName(jString string) {
//	cmd := exec.Command("echo", "-n", jString)
//	cmdReader, err := cmd.StdoutPipe()
//	if err != nil {
//		log.Fatal(err)
//	}
//	scanner := bufio.NewScanner(cmdReader)
//	go func() {
//		for scanner.Scan() {
//			fmt.Println(scanner.Text())
//		}
//	}()
//	if err := cmd.Start(); err != nil {
//		log.Fatal(err)
//	}
//	if err := cmd.Wait(); err != nil {
//		log.Fatal(err)
//	}
//}

package main

import (
	"bufio"
	"os/exec"

	"github.com/pieterclaerhout/go-log"
)

func main() {
	// Print the log timestamps
	log.PrintTimestamp = true

	// The command you want to run along with the argument
	//cmd := exec.Command("brew", "info", "golang")

	cmd := exec.Command("go run ../cmd/earthly/main.go", "+build")

	// Get a pipe to read from standard out
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			log.Info(line)
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()

	// Start the command and check for errors
	err := cmd.Start()
	log.CheckError(err)

	// Wait for all output to be processed
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	log.CheckError(err)

}
