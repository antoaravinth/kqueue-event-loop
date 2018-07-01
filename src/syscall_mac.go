// +build darwin netbsd freebsd openbsd dragonfly

package main

import (
	"syscall"
	"time"
	"fmt"
)

func AddRead(p, fd int) error {
	//fmt.Println("add read")
	//if readon != nil {
	//	if *readon {
	//		return nil
	//	}
	//	*readon = true
	//}
	_, err := syscall.Kevent(p,
		[]syscall.Kevent_t{{Ident: uint64(fd),
			Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ}},
		nil, nil)
	return err
}
func DelRead(p, fd int, readon, writeon *bool) error {
	fmt.Println("deal read")
	if readon != nil {
		if !*readon {
			return nil
		}
		*readon = false
	}
	_, err := syscall.Kevent(p,
		[]syscall.Kevent_t{{Ident: uint64(fd),
			Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ}},
		nil, nil)
	return err
}

func AddWrite(p, fd int) error {
	fmt.Println("add write")
	//if writeon != nil {
	//	if *writeon {
	//		return nil
	//	}
	//	*writeon = true
	//}
	_, err := syscall.Kevent(p,
		[]syscall.Kevent_t{{Ident: uint64(fd),
			Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE}},
		nil, nil)
	return err
}
func DelWrite(p, fd int) error {
	fmt.Println("del write")
	//if writeon != nil {
	//	if !*writeon {
	//		return nil
	//	}
	//	*writeon = false
	//}
	_, err := syscall.Kevent(p,
		[]syscall.Kevent_t{{Ident: uint64(fd),
			Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_WRITE}},
		nil, nil)
	return err
}

func MakePoll() (p int, err error) {
	fmt.Println("making poll")
	return syscall.Kqueue()
}
func MakeEvents(n int) interface{} {
	fmt.Println("making events")
	return make([]syscall.Kevent_t, n)
}
func Wait(p int, evs interface{}, timeout time.Duration) (n int, err error) {
	//fmt.Println("waiting for an event ")
	if timeout < 0 {
		timeout = 0
	}
	ts := syscall.NsecToTimespec(int64(timeout))
	return syscall.Kevent(p, nil, evs.([]syscall.Kevent_t), &ts)
}
func GetFD(evs interface{}, i int) int {
	fmt.Println("getting file descriptor",int(evs.([]syscall.Kevent_t)[i].Ident))
	return int(evs.([]syscall.Kevent_t)[i].Ident)
}
