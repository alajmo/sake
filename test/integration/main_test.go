package integration

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kr/pretty"
)

var tmpDir = "../tmp/"

var debug = flag.Bool("debug", false, "debug")
var update = flag.Bool("update", false, "update golden files")
var clean = flag.Bool("clean", false, "Clean tmp directory after run")

type TemplateTest struct {
	TestName string
	TestCmd  string
	Golden   string
	Ignore   bool
	WantErr  bool
	Index    int
}

func (tt TemplateTest) GoldenOutput(output []byte) []byte {
	out := string(output)
	golden := fmt.Sprintf(
		"Index: %d\nName: %s\nCmd: %s\nWantErr: %t\n\n---\n%s",
		tt.Index, tt.TestName, tt.TestCmd, tt.WantErr, out,
	)

	return []byte(golden)
}

func clearGolden(file string) {
	// Guard against accidentally deleting outside directory
	if strings.Contains(file, "golden") {
		os.RemoveAll(file)
	}
}

func clearTmp() {
	files, _ := os.ReadDir(".")
	for _, f := range files {
		filepath := path.Join(tmpDir, f.Name())
		os.Remove(filepath)
	}
}

func diff(expected, actual any) []string {
	return pretty.Diff(expected, actual)
}

// This function only runs once.
func TestMain(m *testing.M) {
	// Create ./test/tmp if it doesn't exist
	err := os.MkdirAll(tmpDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// Change to ./test/tmp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		panic(err)
	}

	clearTmp()

	os.Exit(m.Run())
}

func Run(t *testing.T, tt TemplateTest) {
	var err error
	log.SetFlags(0)
	// var goldenFile = filepath.Join(tmpDir, tt.Golden)
	if _, err := os.Stat(tt.Golden); os.IsNotExist(err) {
		err = os.WriteFile(tt.Golden, []byte{}, os.ModePerm)
		if err != nil {
			t.Fatalf("could not create golden file at %s: %v", tt.Golden, err)
		}
	}

	// Run test command
	cmd := exec.Command("bash", "-c", tt.TestCmd)
	// export GPG_TTY=$(tty)
	// sockPath, found := os.LookupEnv("SSH_AUTH_SOCK")
	cmd.Env = append(os.Environ(), "GPG_TTY=/dev/pts/32")
	output, err := cmd.CombinedOutput()

	// TEST: Check we get error if we want error
	if (err != nil) != tt.WantErr {
		t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.WantErr, err != nil, err)
	}

	if *debug {
		fmt.Println(tt.TestCmd)
		fmt.Println(string(output))
	}

	// Write output to tmp file which will be used to compare with golden files
	err = os.WriteFile(tt.Golden, tt.GoldenOutput(output), 0644)
	if err != nil {
		t.Fatalf("could not write %s: %v", tt.Golden, err)
	}

	goldenFilePath := filepath.Join("../integration/golden", tt.Golden)
	if *update {
		clearGolden(goldenFilePath)

		// Write stdout of test command to golden file
		err = os.WriteFile(goldenFilePath, tt.GoldenOutput(output), os.ModePerm)
		if err != nil {
			t.Fatalf("could not write %s: %v", goldenFilePath, err)
		}
	} else {
		actual, err := os.ReadFile(tt.Golden)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		expected, err := os.ReadFile(goldenFilePath)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if !tt.Ignore && !reflect.DeepEqual(actual, expected) {
			fmt.Println(text.FgGreen.Sprintf("EXPECTED:"))
			fmt.Println("<---------------------")
			fmt.Println(string(expected))
			fmt.Println("--------------------->")

			fmt.Println()

			fmt.Println(text.FgRed.Sprintf("ACTUAL:"))
			fmt.Println("<---------------------")
			fmt.Println(string(actual))
			fmt.Println("--------------------->")

			t.Fatalf("\nfile: %v\ndiff: %v", text.FgBlue.Sprint(goldenFilePath), diff(expected, actual))
		}

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}

	t.Cleanup(func() {
		if *clean && err == nil {
			clearGolden(tt.Golden)
		}
	})
}
