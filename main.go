package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"unsafe"

	"golang.org/x/sys/unix"
)

func fork() (isChild bool, err error) {
	inParent, _, errno := unix.Syscall(unix.SYS_FORK, 0, 0, 0)
	if inParent == 0 {
		isChild = true
	} else {
		isChild = false
	}
	if errors.Is(errno, fs.ErrNotExist) {
		fmt.Println(errno)
		err = errors.New("failed to create child")
	} else {
		err = nil
	}
	return isChild, err

}

func getCWD() (string, error) {
	buffer := make([]byte, 255)
	buff := unsafe.Pointer(&buffer[0])
	// getcwd takes a pointer to a buffer, and the size of that buffer
	// it populates the buffer with the current working directory
	_, _, err := unix.Syscall(unix.SYS_GETCWD, uintptr(buff), uintptr(len(buffer)), 0)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(err)
		return "", err
	}
	return string(buffer), nil
}

func chdir(path string) error {
	pathptr, err := unix.BytePtrFromString(path)
	if err != nil {
		return err
	}
	_, _, err = unix.Syscall(unix.SYS_CHDIR, uintptr(unsafe.Pointer(pathptr)), 0, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func main() {
	isChild, err := fork()
	if err != nil {
		log.Fatal(err)
	}
	if !isChild {
		fmt.Println("In Parent")
		dir, err := getCWD()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(dir)

		path := ".."
		err = chdir(path)
		if err != nil {
			log.Fatal(err)
		}

		dir, _ = getCWD()
		fmt.Println(dir)
	} else {
		fmt.Println("In Child")
	}
}

// neat straces. Makes sense
//
// fork()                                  = 927567
// futex(0x540220, FUTEX_WAKE_PRIVATE, 1)  = 1
// In child
// write(1, "In parent\n", 10In parent
// /home/nick/Code/go/syscall_exps
// )             = 10
// /home/nick/Code/go
// exit_group(0)                           = ?
// +++ exited with 0 +++
//
//
// and
//
//
// fork()                                  = 927910
// write(1, "In Parent\n", 10In Parent
// In Child
// )             = 10
// getcwd("/home/nick/Code/go/syscall_exps", 255) = 32
// write(1, "/home/nick/Code/go/syscall_exps\0"..., 256/home/nick/Code/go/syscall_exps
// ) = 256
// chdir("..")                             = 0
// getcwd("/home/nick/Code/go", 255)       = 19
// write(1, "/home/nick/Code/go\0\0\0\0\0\0\0\0\0\0\0\0\0\0"..., 256/home/nick/Code/go
// ) = 256
// exit_group(0)                           = ?
// +++ exited with 0 +++
