package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// TODO: Maybe change ext4 to tmpfs for ramdisk speed, might be a secuity hole
// though.

// mount mounts a file as read only
func mount(src, dst string) error {
	// debug
	//fmt.Printf("Mount %q to %q\n", src, dst)

	// if the file does not exist create it
	f, err := os.Create(dst)
	f.Close()
	if err != nil {
		return fmt.Errorf("Create %q: %v", err)
	}

	// Read only here does not work because MS_BIND ignores other flags.
	// to solve this requires use of a remount
	err = unix.Mount(src, dst, "ext4",
		unix.MS_BIND, "")
	if err != nil {
		return fmt.Errorf("Mount %q: %v", dst, err)
	}
	// Now we remount it as MS_RDONLY
	err = unix.Mount("none", dst, "",
		unix.MS_RDONLY|unix.MS_REMOUNT|unix.MS_BIND, "")
	if err != nil {
		return fmt.Errorf("Remount %q: %v", dst, err)
	}

	return nil
}
