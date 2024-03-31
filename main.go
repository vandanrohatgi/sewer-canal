package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

var hostname = "sewer-canal󰖌"
var rootFS = "./busybox/"

func init() {
	fmt.Println("initialising...")
	reexec.Register("nsInit", nsInit)
	if reexec.Init() {
		fmt.Println("nsInit already called. stopping infinite fork..")
		os.Exit(0)
	}
}

func main() {
	cmd := reexec.Command("nsInit")
	fmt.Println("creating namespace...")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}

}

func nsInit() {
	fmt.Println("setting container hostname")
	mountProc()
  pivotRoot()
	syscall.Sethostname([]byte(hostname))
	runCmd()
}

func runCmd() {
	fmt.Println("dropping a shell in " + hostname)
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=" + hostname + "->"} // for sick shell pattern

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func pivotRoot() {
	putold := filepath.Join(rootFS, "/.pivot_root")

	if err := syscall.Mount(
		rootFS,
		rootFS,
		"",
		syscall.MS_BIND|syscall.MS_REC,
		"",
	); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(putold, 0700); err != nil {
		panic(err)
	}

	if err := syscall.PivotRoot(rootFS, putold); err != nil {
		panic(err)
	}

	if err := os.Chdir("/"); err != nil {
		panic(err)
	}

	putold = "/.pivot_root"
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(putold); err != nil {
		panic(err)
	}
}

func mountProc() {
  target:=filepath.Join(rootFS,"/proc")
  os.Mkdir(target,0755)
  if err := syscall.Mount(
		"proc",
		target,
    "proc",
		uintptr(0),
		"",
	); err != nil {
		panic(err)
	}
}
