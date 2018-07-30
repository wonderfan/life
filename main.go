package main

import (
	"flag"
	"fmt"
	"github.com/perlin-network/life/exec"
	"io/ioutil"
	"time"
)

type Resolver struct {
	tempRet0 int64
}

func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	fmt.Printf("Resolve func: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_ping":
			return func(vm *exec.VirtualMachine) int64 {
				return vm.GetCurrentFrame().Locals[0] + 1
			}

		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	fmt.Printf("Resolve global: %s %s\n", module, field)
	switch module {
	case "env":
		switch field {
		case "__life_magic":
			return 424
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

func main() {
	entryFunctionFlag := flag.String("entry", "app_main", "entry function id")
	jitFlag := flag.Bool("jit", false, "enable jit")
	flag.Parse()

	input, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	vm, err := exec.NewVirtualMachine(input, exec.VMConfig{
		EnableJIT:          *jitFlag,
		DefaultMemoryPages: 128,
		DefaultTableSize:   65536,
	}, &Resolver{})
	if err != nil {
		panic(err)
	}

	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", *entryFunctionFlag)
		entryID = 0
	}

	start := time.Now()

	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}

	ret, err := vm.Run(entryID)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	end := time.Now()
	fmt.Printf("return value = %d, duration = %v\n", ret, end.Sub(start))
}
