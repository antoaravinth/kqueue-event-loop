package main

import (
	"net"
	"os"
	"strings"
	"fmt"
	"syscall"
	"time"
	"strconv"
	"sync"
)

type listener struct {
	ln      net.Listener
	lnaddr  net.Addr
	pconn   net.PacketConn
	f       *os.File
	fd      int
	network string
	addr    string
}

func (ln *listener) system() error {
	var err error
	switch netln := ln.ln.(type) {
	default:
		panic("invalid listener type")
	case *net.TCPListener:
		ln.f, err = netln.File()
	}

	if err != nil {
		return err
	}

	ln.fd = int(ln.f.Fd())
	return syscall.SetNonblock(ln.fd, true)
}

func Serve(addr string) error {
	var lns []*listener

	var ln listener
	ln.network, ln.addr = parseAddr(addr)

	ln.ln, _ = net.Listen(ln.network, ln.addr)
	ln.lnaddr = ln.ln.Addr()

	fmt.Println(ln)

	ln.system()

	lns = append(lns, &ln)

	serve(lns)
	return nil
}

func serve(lns []*listener) error {
	p, err := MakePoll()
	var accepted bool;
	accepted = false

	if err != nil {
		return err
	}
	defer syscall.Close(p)
	for _, ln := range lns {
		if err := AddRead(p, ln.fd); err != nil {
			return err
		}
	}

	var evs = MakeEvents(64)
	var rsa syscall.Sockaddr
	var packet [0xFFFF]byte
	nextTicker := time.Now()
	var mu sync.Mutex
	lock := func() { mu.Lock() }
	unlock := func() { mu.Unlock() }
	for {
		delay := nextTicker.Sub(time.Now())
		if delay < 0 {
			delay = 0
		} else if delay > time.Second/4 {
			delay = time.Second / 4
		}
		pn, err := Wait(p, evs,delay)
		if err != nil && err != syscall.EINTR {
			return err
		}

		remain := nextTicker.Sub(time.Now())
		if remain < 0 {
			var tickerDelay time.Duration
			nextTicker = time.Now().Add(tickerDelay + remain)
		}

		accepted = false;
		lock()
		for i := 0 ; i< pn; i++ {
			var nfd int
			var fd = GetFD(evs, i)
			if accepted {
				goto read
			} else {
				goto accept
			}

			accept:
				nfd, rsa, err = syscall.Accept(fd)
				fmt.Println("ndf",nfd,fd)
				if err = syscall.SetNonblock(nfd, true); err != nil {
					fmt.Println("failed......",err)
				}

				fmt.Println("rsa",rsa)

				AddWrite(p, nfd)
				accepted = true
				goto opened

			opened:
				AddRead(p, nfd)
				fmt.Println("asdsad")
			read:
				n, err := syscall.Read(nfd, packet[:])
				fmt.Println(err)
				if err == nil {
					fmt.Println(string(packet[:n]))
				}
				fmt.Println("read",n)
				goto write

			//next:
			//	fmt.Println("in next")
			//	goto write

			write:
				var b []byte
				var result = appendresp(b, "200 OK", "", "Hello World!\r\n")
				fmt.Println(string(result))
				syscall.Write(nfd,result)
			//	goto close
			//close:
			//	syscall.Close(nfd)
		}
		unlock()
		time.Sleep(250000000)
	}

	return nil
}

func parseAddr(addr string) (network, address string) {
	network = "tcp"
	address = addr
	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}
	return
}

func appendresp(b []byte, status, head, body string) []byte {
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, status...)
	b = append(b, '\r', '\n')
	b = append(b, "Server: evio\r\n"...)
	b = append(b, "Date: "...)
	b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, "Content-Length: "...)
		b = strconv.AppendInt(b, int64(len(body)), 10)
		b = append(b, '\r', '\n')
	}
	b = append(b, head...)
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, body...)
	}
	return b
}

func main() {
	Serve("tcp://:8080")
}