// Package main provdires a simple docker implementation in go for linux
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func init() { log.SetFlags(log.Lshortfile) }

// the directory for the temporary container files to be in
const cDir = "/opt/minidocker"

func main() {
	// if you are being called inside a container then execute container
	// instead of main
	if dir := os.Getenv("CONTAINER_DIR"); dir != "" {
		container(dir)
		return
	}

	// check that they executed us correctly
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <command> [args...]\n", os.Args[0])
		return
	}

	// create a temporary directory for the command to be executed in
	dir, err := ioutil.TempDir(cDir, "CONTAINER-")
	if err != nil {
		log.Println(err)
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"CONTAINER_DIR=" + dir}

	cmd.SysProcAttr = &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}

	// cleanup the directory we're using as root, container can't because
	// it will have chrooted into it
	defer func() {
		if err := unix.Unmount(dir+"/main", 0); err != nil {
			log.Printf("Unmount failed: %v", err)
		}

		if err := os.RemoveAll(dir); err != nil {
			log.Printf("Remove failed: %v", err)
		}
	}()

	if err := cmd.Run(); err != nil {
		log.Println(err)
	}
}

// called if the program is being run from inside a container
func container(dir string) {
	// mount the executeable into /main
	if err := mount(os.Args[1], dir+"/main"); err != nil {
		log.Println(err)
		return
	}

	// set the hostname
	sep := strings.LastIndex(dir, "/")
	hostname := dir[sep+1:]
	if err := unix.Sethostname([]byte(hostname)); err != nil {
		log.Println(err)
		return
	}
	// chroot into the system
	if err := unix.Chroot(dir); err != nil {
		log.Println(err)
		return
	}

	// now run the command
	cmd := exec.Command("/main", os.Args[2:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println(err)
		return
	}
}
