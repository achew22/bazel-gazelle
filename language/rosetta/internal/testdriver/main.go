// Package main is a sample "driver" for the gazelle rosetta plugin. It is used
// to validate the functionality of the rosetta plugin but can also be used as a
// template for implementing a language plugin in another programming language.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/jsonpb"

	pb "github.com/bazelbuild/bazel-gazelle/language/rosetta/proto"
)

func logf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func main() {
	unmarshaler := jsonpb.Unmarshaler{}
	marshaler := jsonpb.Marshaler{}

	// Block reading on standard in until a response is able to happen.
	recvCh := make(chan *pb.Request, 1)
	sendCh := make(chan *pb.Response, 1)
	doneCh := make(chan error, 1)

	// Listener loop.
	go func() {
		var inputBuffer bytes.Buffer
		r := io.TeeReader(os.Stdin, &inputBuffer)

		for {
			var msg pb.Request
			if err := unmarshaler.Unmarshal(r, &msg); err == io.EOF {
				logf("Got EOF, terminating.")
				close(recvCh)
				doneCh <- err
				return
			} else if err != nil {
				logf("Unable to unmarshal: %v\nData:\n%s\n", err, inputBuffer.String())
				close(recvCh)
				doneCh <- err
				return
			}

			recvCh <- &msg
		}
	}()

	// Sender loop.
	go func() {
		for msg := range sendCh {
			if err := marshaler.Marshal(os.Stdout, msg); err != nil {
				panic(fmt.Sprintf("err marshalling: %v", err))
			}
		}
	}()

	for {
		select {
		case err := <-doneCh:
			close(sendCh)
			if err == nil {
				os.Exit(0)
			} else if err == io.EOF {
				os.Exit(0)
			} else {
				logf("Error: %v", err)
				os.Exit(1)
			}
		case msg := <-recvCh:
			logf("Got msg: %v", msg)
			if msg.GetGenerateRules() != nil {
				logf("Got dir as %q", msg.GetGenerateRules().GenerateArgs.GetDir())
			}
			sendCh <- &pb.Response{
				Return: &pb.Response_GenerateRules{
					GenerateRules: &pb.GenerateRulesResponse{
						GenerateResult: &pb.GenerateResult{
							Gen: []*pb.Rule{
								&pb.Rule{Kind: "Kind!"},
							},
						},
					},
				},
			}
		}
	}
}
