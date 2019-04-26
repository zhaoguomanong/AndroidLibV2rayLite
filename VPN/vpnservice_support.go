package VPN

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/unix"
	v2net "v2ray.com/core/common/net"
	v2internet "v2ray.com/core/transport/internet"
)

type protectSet interface {
	Protect(int) int
}

type VPNProtectedDialer struct {
	SupportSet    protectSet
	DomainName    string
	PreparedReady bool
	preparedIPs   []net.IP
	preparedPort  int
}

func (d *VPNProtectedDialer) PrepareDomain(pch chan<- bool) {
	log.Printf("Preparing Domain: %s", d.DomainName)
	_host, _port, serr := net.SplitHostPort(d.DomainName)
	_iport, perr := strconv.Atoi(_port)
	if serr != nil || perr != nil {
		log.Printf("PrepareDomain DomainName Err: %v|%v", serr, perr)
		goto PEND
	}
	d.preparedPort = _iport
	if ips, err := net.LookupIP(_host); err == nil {
		d.preparedIPs = ips
		d.PreparedReady = true
	} else {
		log.Printf("PrepareDomain LookupIP Err: %v", err)
	}

PEND:
	pch <- true
	log.Printf("Prepare Result:\n Ready: %v\n Domain: %s\n Port: %s\n IPs: %v\n", d.PreparedReady, _host, _port, d.preparedIPs)
}

func (d *VPNProtectedDialer) prepareFd(network v2net.Network) (fd int, err error) {
	if network == v2net.Network_TCP {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, unix.IPPROTO_TCP)
		d.SupportSet.Protect(fd)
	} else if network == v2net.Network_UDP {
		fd, err = unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
		d.SupportSet.Protect(fd)
	} else {
		err = fmt.Errorf("unknow network")
	}
	return
}

func (d VPNProtectedDialer) Dial(ctx context.Context, src v2net.Address, dest v2net.Destination, sockopt *v2internet.SocketConfig) (net.Conn, error) {
	network := dest.Network.SystemString()
	Address := dest.NetAddr()

	if Address == d.DomainName {
		ip := d.preparedIPs[rand.Intn(len(d.preparedIPs))]
		if fd, err := d.prepareFd(dest.Network); err == nil {
			return d.fdConn(ctx, ip, d.preparedPort, fd)
		} else {
			return nil, err
		}
	}

	var _port int
	var _ip net.IP

	if dest.Network == v2net.Network_TCP {
		addr, err := net.ResolveTCPAddr(network, Address)
		if err != nil {
			return nil, err
		}
		log.Println("Not Using Prepared: TCP,", Address)
		_port = addr.Port
		_ip = addr.IP.To16()
	} else if dest.Network == v2net.Network_UDP {
		addr, err := net.ResolveUDPAddr(network, Address)
		if err != nil {
			return nil, err
		}
		log.Println("Not Using Prepared: UDP,", Address)
		_port = addr.Port
		_ip = addr.IP.To16()
	} else {
		return nil, fmt.Errorf("unsupported network type")
	}

	if fd, err := d.prepareFd(dest.Network); err == nil {
		return d.fdConn(ctx, _ip, _port, fd)
	} else {
		return nil, err
	}
}

func (d VPNProtectedDialer) fdConn(ctx context.Context, ip net.IP, port int, fd int) (net.Conn, error) {

	d.SupportSet.Protect(fd)
	sa := &unix.SockaddrInet6{
		Port: port,
	}
	copy(sa.Addr[:], ip)

	if err := unix.Connect(fd, sa); err != nil {
		log.Printf("fdConn Connect Close Fd: %d Err: %v", fd, err)
		unix.Close(fd)
		return nil, err
	}

	file := os.NewFile(uintptr(fd), "Socket")
	conn, err := net.FileConn(file)
	if err != nil {
		log.Printf("fdConn FileConn Close Fd: %d Err: %v: %d", fd, err)
		file.Close()
		unix.Close(fd)
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			file.Close()
			unix.Close(fd)
		}
		return
	}()

	return conn, nil
}

func init() {
	rand.Seed(time.Now().Unix())
}
