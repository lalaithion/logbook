package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/achew22/logbook/parser"
)

func trim(s string) string {
	return strings.Trim(s, " \t\n")
}

// makeFakeHome creates a direcory shaped approximately like what a home dir would look like.
func makeFakeHome(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "logbook")
	if err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		os.RemoveAll(dir) // clean up
	}
}

func makeLogEntry(t *testing.T, homeDir, fileName, contents string) {
	// MkdirAll instead of Mkdir so that if you make it multiple times no
	// one cares.
	err := os.MkdirAll(filepath.Join(homeDir, "logbook"), 0700)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(homeDir, "logbook", fileName+".md"), []byte(contents), 0600)
	if err != nil {
		t.Error(err)
	}
}

func assertLogEntry(t *testing.T, homeDir, fileName, want string) {
	fullPath := filepath.Join(homeDir, "logbook", fileName+".md")
	got, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(strings.Split(string(got), "\n"), strings.Split(want, "\n")); diff != "" {
		t.Errorf("Diff for %s:\n!!! (- = got, + = want)\n%s", fullPath, diff)
	}
}

func helperCommandContext(t *testing.T, fakeHome string, s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)

	cmd := exec.Command(os.Args[0], cs...)

	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		"HOME=" + fakeHome,
	}
	return cmd
}

func helperCommand(t *testing.T, fakeHome string, s ...string) *exec.Cmd {
	return helperCommandContext(t, fakeHome, s...)
}

// TestHelperProcess isn't a real test.
//
// Some details elided for this blog post.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	// Cleaning up the args before handing off arg parsing. We know we were invoked with 2 unnecessary args ("-test.run=..." and "--"). Strip em out.
	newArgs := []string{os.Args[0]}
	if len(os.Args) > 3 {
		newArgs = append(newArgs, os.Args[3:]...)
	}
	os.Args = newArgs

	// Invoke the main for the command.
	main()
}

func TestBasicInvocation(t *testing.T) {
	dir, cleanup := makeFakeHome(t)
	defer cleanup()
	t.Logf("Dir: %s", dir)

	dayBeforeYesterday := parser.TimeToDate(time.Now().AddDate(0, 0, -2)).ToYmd()
	makeLogEntry(t, dir, dayBeforeYesterday, `# Day before yesterday

It was an okay day
`)

	yesterday := parser.TimeToDate(time.Now().AddDate(0, 0, -1)).ToYmd()
	makeLogEntry(t, dir, yesterday, `# Yesterday stuff

It was important!
`)

	got, err := helperCommand(t, dir).CombinedOutput()
	gotString := trim(string(got))
	today := parser.TimeToDate(time.Now()).ToYmd()
	want := "Writing log entry for " + today + "\nWrote file"
	if err != nil {
		t.Errorf("Invocation failed: %v\nwant: %q\ngot:  %q", err, want, gotString)
	}
	if gotString != want {
		t.Errorf("Inequal stderr/out:\nwant: %X\ngot:  %X", want, gotString)
	}

	assertLogEntry(t, dir, today, "# Andrew Allen - "+today+`

Reminders:
From 2018-12-29: Tomorrow text
From 2018-12-29: Tomorrow text
From 2018-12-29: Tomorrow text

`)
}

func TestInvalidDateOverride(t *testing.T) {
	dir, cleanup := makeFakeHome(t)
	defer cleanup()
	t.Logf("Dir: %s", dir)

	got, err := helperCommand(t, dir, "--date_override=2001-02-29").CombinedOutput()
	gotString := trim(string(got))
	want := "Invalid --date_override provided. parsing time \"2001-02-29\": day out of range"
	if err == nil {
		t.Errorf("Invocation succeeded when it shouldn't have: %v\nwant: %q\ngot:  %q", err, want, gotString)
	}
	if gotString != want {
		t.Errorf("Inequal stderr/out:\nwant: %q\ngot:  %q", want, gotString)
	}
}

func TestCreateLogDirectory(t *testing.T) {
	dir, cleanup := makeFakeHome(t)
	defer cleanup()
	t.Logf("Dir: %s", dir)

	got, err := helperCommand(t, dir, "--date_override=2000-01-01").CombinedOutput()
	gotString := trim(string(got))
	want := fmt.Sprintf("Writing log entry for 2000-01-01\nCreating %s/logbook\nWrote file", dir)

	if err != nil {
		t.Errorf("Invocation failed:\nerr: %s\noutput: %s\n", err.Error(), got)
	}
	if gotString != want {
		t.Errorf("Inequal stderr/out:\nwant: %q\ngot:  %q", want, gotString)
	}
}
