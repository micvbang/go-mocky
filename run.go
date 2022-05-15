package mocky

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Run(flags Flags) {
	if flags.OutputPath == nil {
		flags.OutputPath = DefaultOutputPath
	}

	interfaces, err := Parse(flags.InterfaceDir)
	if err != nil {
		log.Fatalf("failed to parse dir %s: %s", flags.InterfaceDir, err)
	}

	var ifaceToGenerate *Interface
	for _, iface := range interfaces {
		if flags.InterfaceName == iface.Name {
			ifaceToGenerate = &iface
			break
		}
	}
	if ifaceToGenerate == nil {
		log.Fatalf("interface '%s' not found", flags.InterfaceName)
	}

	fpath := flags.OutputPath(flags)
	log.Printf("writing to file %s", fpath)

	f, err := os.Create(fpath)
	if err != nil {
		log.Fatalf("failed to open file %s: %s", fpath, err)
	}
	defer f.Close()

	err = Generate(f, *ifaceToGenerate)
	if err != nil {
		log.Fatalf("failed to generate: %s", err)
	}
}

func DefaultOutputPath(f Flags) string {
	fname := fmt.Sprintf("mock_%s.go", strings.ToLower(f.InterfaceName))
	return filepath.Join(f.InterfaceDir, fname)
}

type Flags struct {
	InterfaceDir  string
	InterfaceName string

	OutputPath func(Flags) string
}

func ParseFlags() (Flags, error) {
	fs := flag.NewFlagSet("mocky", flag.ExitOnError)

	output := Flags{}

	fs.StringVar(&output.InterfaceDir, "d", "", "directory containing .go file with interface")
	fs.StringVar(&output.InterfaceName, "i", "", "name of interface to mock (case sensitive)")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return Flags{}, err
	}

	if len(output.InterfaceDir) == 0 {
		d, err := os.Getwd()
		if err != nil {
			return Flags{}, err
		}
		output.InterfaceDir = d
	}

	if len(output.InterfaceName) == 0 {
		fmt.Printf("You must provide an interface name\n")
		fs.Usage()
		os.Exit(1)
	}

	return output, nil
}
