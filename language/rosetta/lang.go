package rosetta

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/golang/protobuf/jsonpb"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"

	pb "github.com/bazelbuild/bazel-gazelle/language/rosetta/proto"
)

const rosettaName = "rosetta"

type rosettaLang struct {
	cmd    *exec.Cmd
	recvCh chan *pb.Response
	sendCh chan *pb.Request
	doneCh chan error
}

// NewLanguage is a thing
func NewLanguage() language.Language {
	bin, ok := bazel.FindBinary("language/rosetta/internal/testdriver", "testdriver")
	if !ok {
		panic(fmt.Sprintf("Unable to find binary: %v", bin))
	}
	cmd := exec.Command(bin)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(fmt.Sprintf("Unable to make a StdinPipe: %v", err))
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("Unable to make a StdoutPipe: %v", err))
	}
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("Unable to start: %v", err))
	}

	unmarshaler := jsonpb.Unmarshaler{}
	marshaler := jsonpb.Marshaler{}

	// Block reading on standard in until a response is able to happen.
	recvCh := make(chan *pb.Response, 1)
	sendCh := make(chan *pb.Request, 1)
	doneCh := make(chan error, 1)

	// Listener loop.
	go func() {
		var inputBuffer bytes.Buffer
		r := io.TeeReader(stdout, &inputBuffer)

		for {
			var msg pb.Response
			if err := unmarshaler.Unmarshal(r, &msg); err == io.EOF {
				close(recvCh)
				doneCh <- err
				return
			} else if err != nil {
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
			fmt.Println("Gazelle Listener loop")
			if err := marshaler.Marshal(stdin, msg); err != nil {
				panic(fmt.Sprintf("err marshalling: %v", err))
			}
		}
	}()

	return &rosettaLang{
		cmd: cmd,

		sendCh: sendCh,
		recvCh: recvCh,
		doneCh: doneCh,
	}
}

func (*rosettaLang) Name() string { return rosettaName }

func (*rosettaLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

func (*rosettaLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error { return nil }

func (*rosettaLang) KnownDirectives() []string { return nil }

func (*rosettaLang) Configure(c *config.Config, rel string, f *rule.File) {}

func (*rosettaLang) Kinds() map[string]rule.KindInfo {
	return kinds
}

func (*rosettaLang) Loads() []rule.LoadInfo { return nil }

func (*rosettaLang) Fix(c *config.Config, f *rule.File) {}

func (*rosettaLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return nil
}

func (*rosettaLang) Embeds(r *rule.Rule, from label.Label) []label.Label { return nil }

func (*rosettaLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
}

var kinds = map[string]rule.KindInfo{
	"filegroup": {
		NonEmptyAttrs:  map[string]bool{"srcs": true, "deps": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
}

func (r *rosettaLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	m := &pb.Request{
		Function: &pb.Request_GenerateRules{
			GenerateRules: &pb.GenerateRulesRequest{
				GenerateArgs: &pb.GenerateArgs{
					Dir: args.Dir,
					//
				},
			},
		},
	}
	r.sendCh <- m

	fmt.Printf("Response: %v\n", <-r.recvCh)

	/*
		r := rule.NewRule("filegroup", "all_files")
		srcs := make([]string, 0, len(args.Subdirs)+len(args.RegularFiles))
		for _, f := range args.RegularFiles {
			srcs = append(srcs, f)
		}
		for _, f := range args.Subdirs {
			pkg := path.Join(args.Rel, f)
			srcs = append(srcs, "//"+pkg+":all_files")
		}
		r.SetAttr("srcs", srcs)
		r.SetAttr("testonly", true)
		if args.File == nil || !args.File.HasDefaultVisibility() {
			r.SetAttr("visibility", []string{"//visibility:public"})
		}
	*/
	return language.GenerateResult{
		//Gen:     []*rule.Rule{r},
		Imports: []interface{}{nil},
	}
}
