package main

import (
	"fmt"
	"net"
	"os"

	"github.com/brownsys/tracing-framework-go/xtrace/client"
	xtgrpc "github.com/brownsys/tracing-framework-go/xtrace/grpc"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	xtraceServerFlag string
	nextServerFlag   string
	listenAddrFlag   string
	addFlag          int64
	subFlag          int64
	mulFlag          int64
	divFlag          int64
	inputFlag        int64
)

func init() {
	pflag.StringVar(&xtraceServerFlag, "xtrace-server", "", "The X-Trace server to send logs to.")
	pflag.StringVar(&nextServerFlag, "next-server", "", "The next server to invoke RPCs on, if any.")
	pflag.StringVar(&listenAddrFlag, "listen-addr", "", "The address to listen for RPC calls on.")
	pflag.Int64Var(&addFlag, "add", 0, "Add this value to the argument.")
	pflag.Int64Var(&subFlag, "sub", 0, "Subtract this value from the argument.")
	pflag.Int64Var(&mulFlag, "mul", 0, "Multiply the argument by this value.")
	pflag.Int64Var(&divFlag, "div", 0, "Divide the argument by this value.")
	pflag.Int64Var(&inputFlag, "in", 0, "Instead of hosting an RPC server, call --next-server with this argument.")
}

type server struct {
	op       func(int64) int64
	opname   string
	next     OperatorClient
	nextname string
}

func (s *server) PerformOp(ctx context.Context, in *Int) (*Int, error) {
	err := xtgrpc.ExtractIDs(ctx)
	if err != nil {
		client.Log(fmt.Sprintf("could not extract Task ID in RPC handler: %v", err))
	}

	out := &Int{s.op(in.Num)}
	client.Log(fmt.Sprintf("applying operation %v to argument %v: %v", s.opname, in.Num, out.Num))
	if s.next != nil {
		client.Log(fmt.Sprintf("passing to %v", s.nextname))
		out, err = s.next.PerformOp(ctx, out)
		if err != nil {
			client.Log(fmt.Sprintf("error invoking RPC: %v", err))
			return nil, err
		}
		return out, nil
	}
	client.Log(fmt.Sprintf("returning value %v", out.Num))
	return out, nil
}

func main() {
	pflag.Parse()

	usage := func() { pflag.PrintDefaults(); os.Exit(1) }

	// number of operations flags set
	numSet := 0
	for _, f := range []string{"add", "sub", "mul", "div"} {
		if pflag.Lookup(f).Changed {
			numSet++
		}
	}

	switch {
	case !pflag.Lookup("xtrace-server").Changed:
		fmt.Fprintln(os.Stderr, "Must provide --xtrace-server")
		usage()
	case pflag.Lookup("in").Changed && (numSet > 0 || pflag.Lookup("listen-addr").Changed):
		fmt.Fprintln(os.Stderr, "Cannot provide --in and any of --listen-addr, --add, --sub, --mul, or --div")
		usage()
	case !pflag.Lookup("in").Changed && (numSet == 0 || numSet > 1 || !pflag.Lookup("listen-addr").Changed):
		fmt.Fprintln(os.Stderr, "Must provide --in or exactly one of --add, --sub, --mul, --div, and also --listen-addr")
		usage()
	case pflag.Lookup("in").Changed && !pflag.Lookup("next-server").Changed:
		fmt.Fprintln(os.Stderr, "Cannot provide --in without --next-server")
		usage()
	}

	if err := client.Connect(xtraceServerFlag); err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to X-Trace server: %v\n", err)
		os.Exit(2)
	}

	var next OperatorClient
	if pflag.Lookup("next-server").Changed {
		var err error
		cc, err := grpc.Dial(nextServerFlag, grpc.WithInsecure())
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to next gRPC server: %v\n", err)
			os.Exit(2)
		}
		next = NewOperatorClient(cc)
	}

	if pflag.Lookup("in").Changed {
		client.Log(fmt.Sprintf("passing --in (%v) to %v", inputFlag, nextServerFlag))
		out, err := next.PerformOp(context.Background(), &Int{inputFlag})
		if err != nil {
			client.Log(fmt.Sprintf("error invoking RPC: %v", err))
			os.Exit(2)
		}
		fmt.Println("Got value", out.Num)
		return
	}

	var op func(int64) int64
	var opname string
	switch {
	case pflag.Lookup("add").Changed:
		op = func(in int64) int64 { return in + addFlag }
		opname = fmt.Sprintf("+%v", addFlag)
	case pflag.Lookup("sub").Changed:
		op = func(in int64) int64 { return in - subFlag }
		opname = fmt.Sprintf("-%v", subFlag)
	case pflag.Lookup("mul").Changed:
		op = func(in int64) int64 { return in * mulFlag }
		opname = fmt.Sprintf("*%v", mulFlag)
	case pflag.Lookup("div").Changed:
		op = func(in int64) int64 { return in / divFlag }
		opname = fmt.Sprintf("/%v", divFlag)
	}

	srv := server{
		op:       op,
		opname:   opname,
		next:     next,
		nextname: nextServerFlag,
	}

	lis, err := net.Listen("tcp", listenAddrFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to listen:", err)
		os.Exit(2)
	}
	s := grpc.NewServer()
	RegisterOperatorServer(s, &srv)
	s.Serve(lis)
}
