package VPN

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
	v2net "v2ray.com/core/common/net"
	v2internet "v2ray.com/core/transport/internet"
)

type protectSet interface {
	Protect(int) int
}

type VPNProtectedDialer struct {
	SupportSet   protectSet
	DomainName   string
	DomainIP     string
	preparedIP   net.IP
	preparedPort int
}

func (d *VPNProtectedDialer) PrepareDomain() {
	log.Printf("Preparing Domain: %s", d.DomainName)
	_host, _port, err := net.SplitHostPort(d.DomainName)
	if err != nil {
		log.Fatalf("DomainName Err: %v", err)
	}
	_iport, err := strconv.Atoi(_port)
	if err != nil {
		log.Fatalf("DomainName Err: %v", err)
	}
	d.preparedPort = _iport

	if d.DomainIP != "" {
		d.preparedIP = net.ParseIP(d.DomainIP)
	} else {
		d.preparedIP = net.ParseIP(_host)
	}
}

func (d VPNProtectedDialer) Dial(ctx context.Context, src v2net.Address, dest v2net.Destination, sockopt *v2internet.SocketConfig) (net.Conn, error) {
	network := dest.Network.SystemString()
	Address := dest.NetAddr()

	if dest.Network == v2net.Network_TCP {

		var addr *net.TCPAddr
		var err error

		if d.preparedIP != nil && Address == d.DomainName {
			addr = &net.TCPAddr{
				IP:   d.preparedIP,
				Port: d.preparedPort,
			}
			log.Println("Using Prepared: TCP,", Address)
		} else {
			addr, err = net.ResolveTCPAddr(network, Address)
			log.Println("Not Using Prepared: TCP,", Address)
		}

		fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, unix.IPPROTO_TCP)
		if err != nil {
			return nil, err
		}

		go func() {
			select {
			case <-ctx.Done():
			}
			syscall.Close(fd)
			log.Printf("Closed Fd:%d", fd)
			return
		}()

		d.SupportSet.Protect(fd)

		sa := new(unix.SockaddrInet6)
		sa.Port = addr.Port
		copy(sa.Addr[:], addr.IP.To16())

		if err := unix.Connect(fd, sa); err != nil {
			syscall.Close(fd)
			return nil, err
		}

		file := os.NewFile(uintptr(fd), "Socket")
		conn, err := net.FileConn(file)
		if err != nil {
			syscall.Close(fd)
			return nil, err
		}

		return conn, nil
	}

	if dest.Network == v2net.Network_UDP {
		var addr *net.UDPAddr
		var err error

		if d.preparedIP != nil && Address == d.DomainName {
			addr = &net.UDPAddr{
				IP:   d.preparedIP,
				Port: d.preparedPort,
			}
			log.Println("Using Prepared: UDP,", Address)
		} else {
			addr, err = net.ResolveUDPAddr(network, Address)
			log.Println("Not Using Prepared: UDP,", Address)
		}

		if err != nil {
			return nil, err
		}

		if addr == nil {
			return nil, fmt.Errorf("Fail to resolve address %s/%s", network, Address)
		}

		fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
		if err != nil {
			return nil, err
		}

		go func() {
			select {
			case <-ctx.Done():
			}
			syscall.Close(fd)
			log.Printf("Closed Fd:%d", fd)
			return
		}()

		d.SupportSet.Protect(fd)

		sa := new(unix.SockaddrInet6)
		sa.Port = addr.Port
		copy(sa.Addr[:], addr.IP.To16())

		if err := unix.Connect(fd, sa); err != nil {
			syscall.Close(fd)
			return nil, err
		}

		file := os.NewFile(uintptr(fd), "Socket")
		conn, err := net.FileConn(file)
		if err != nil {
			syscall.Close(fd)
			return nil, err
		}

		return conn, nil
	}
	return nil, errors.New("Pto udf")
}
