// Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var BuildVersion string = "dev"

var (
	printLicense    bool   = false
	noOutputLicense bool   = false
	udpMode         bool   = false
	binaryMode      bool   = false
	proxyAddr       string = ""
	bufferLen       int    = 8
	linebreak       string = "lf"
	linebreakCh     string
	listenAt        string = ""
	targetAddr      string
)

func init() {
	flag.BoolVar(&printLicense, "license", printLicense, "Print the license")
	flag.BoolVar(&noOutputLicense, "no-license", noOutputLicense, "Do not output the license (but you still have to follow it)")
	flag.BoolVar(&udpMode, "udp", udpMode, "Use UDP mode instead of default TCP")
	flag.BoolVar(&udpMode, "u", udpMode, "Use UDP mode instead of default TCP")
	flag.BoolVar(&binaryMode, "binary", binaryMode, `Use hex format input/output for binary data, aka "-b"`)
	flag.BoolVar(&binaryMode, "b", binaryMode, `Use hex format input/output for binary data, aka "-binary"`)
	flag.IntVar(&bufferLen, "buffer", bufferLen, `The maximum output bytes per line when under binary mode, aka "-B"`)
	flag.IntVar(&bufferLen, "B", bufferLen, `The maximum output bytes per line when under binary mode, aka "-buffer"`)
	flag.StringVar(&proxyAddr, "proxy", proxyAddr, `The socks5 proxy address, aka "-x"`)
	flag.StringVar(&proxyAddr, "x", proxyAddr, `The socks5 proxy address, aka "-proxy"`)
	flag.StringVar(&linebreak, "linebreak", linebreak, `The line break, cr='\r', lf='\n', crlf='\r\n'`)
	flag.StringVar(&listenAt, "listen", listenAt, `Local bind address, aka "-l"`)
	flag.StringVar(&listenAt, "l", listenAt, `Local bind address, aka "-listen"`)
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		fmt.Fprint(out, helpMessage)
		fmt.Fprintln(out, "Flags:")
		flag.PrintDefaults()
	}
}

func printErr(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
}

func parseFlags() {
	flag.Parse()
	if printLicense {
		fmt.Print(LICENSE)
		os.Exit(0)
	}
	switch linebreak {
	case "cr":
		linebreakCh = "\r"
	default:
		fallthrough
	case "lf":
		linebreakCh = "\n"
	case "crlf":
		linebreakCh = "\r\n"
	}
	targetAddr = flag.Arg(0)
}

type Dialer interface {
	Dial(network string, addr string) (c net.Conn, err error)
	ListenPacket(network string, addr string) (c net.PacketConn, err error)
}

func main() {
	parseFlags()
	printErr(LICENSE)
	var dialer Dialer = netDialer{}
	if proxyAddr != "" {
		var err error
		if dialer, err = NewSocks5(proxyAddr); err != nil {
			printErr("Cannot create proxy:", err)
			os.Exit(1)
		}
	}
	if udpMode {
		handleUDP(dialer)
	} else {
		handleTCP(dialer)
	}
}

func handleTCP(dialer Dialer) {
	conn, err := dialer.Dial("tcp", targetAddr)
	if err != nil {
		printErr("Err when dialing:", err)
		os.Exit(1)
	}
	defer conn.Close()
	input := bufio.NewScanner(os.Stdin)
	if binaryMode {
		go func() {
			var err error
			defer conn.Close()
			for input.Scan() {
				line := input.Text()
				if len(line) == 0 {
					continue
				}
				if line[0] == '!' {
					if _, err = fmt.Fprintln(conn, line[1:]); err != nil {
						printErr("Err when writing:", err)
						os.Exit(1)
					}
				} else {
					bts, err := parseBytes(line)
					if err != nil {
						printErr("Invaild input:", err)
					} else if _, err = conn.Write(bts); err != nil {
						printErr("Err when writing:", err)
						os.Exit(1)
					}
				}
			}
		}()
		buf := make([]byte, bufferLen)
		var (
			n   int
			err error
		)
		for {
			n, err = conn.Read(buf)
			if n != 0 {
				fmt.Println(formatBytes(buf[:n]))
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					printErr("Err when reading:", err)
					os.Exit(1)
				}
				return
			}
		}
	} else {
		go func() {
			var err error
			defer conn.Close()
			for input.Scan() {
				line := input.Text()
				if _, err = fmt.Fprintln(conn, line); err != nil {
					printErr("Err when writing:", err)
					os.Exit(1)
				}
			}
		}()
		sc := bufio.NewScanner(conn)
		for sc.Scan() {
			fmt.Println(sc.Text())
		}
		if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
			printErr("Err when reading:", err)
			os.Exit(1)
		}
	}
}

func handleUDP(dialer Dialer) {
	target := &stringAddr{"udp", targetAddr}
	conn, err := dialer.ListenPacket("udp", listenAt)
	if err != nil {
		printErr("Err when dialing:", err)
		os.Exit(1)
	}
	defer conn.Close()
	printErr("# local addr =", conn.LocalAddr().String())
	input := bufio.NewScanner(os.Stdin)
	var (
		n    int
		addr net.Addr
		buf  = make([]byte, 65536)
	)
	if binaryMode {
		go func() {
			var err error
			defer conn.Close()
			for input.Scan() {
				line := input.Text()
				if len(line) == 0 {
					continue
				}
				if line[0] == '!' {
					line += linebreakCh
					if _, err = conn.WriteTo(([]byte)(line[1:]), target); err != nil {
						printErr("Err when writing:", err)
						os.Exit(1)
					}
				} else {
					bts, err := parseBytes(line)
					if err != nil {
						printErr("Invaild input:", err)
					} else if _, err = conn.WriteTo(bts, target); err != nil {
						printErr("Err when writing:", err)
						os.Exit(1)
					}
				}
			}
		}()
		for {
			n, addr, err = conn.ReadFrom(buf)
			if n != 0 {
				printErr("# recv from", addr.String(), "leng =", n)
				for i, j := 0, bufferLen; i < n; i, j = j, j+bufferLen {
					if j > n {
						j = n
					}
					fmt.Println(formatBytes(buf[i:j]))
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					printErr("Err when reading:", err)
					os.Exit(1)
				}
				return
			}
		}
	} else {
		go func() {
			var err error
			defer conn.Close()
			for input.Scan() {
				line := input.Text()
				line += linebreakCh
				if _, err = conn.WriteTo(([]byte)(line), target); err != nil {
					printErr("Err when writing:", err)
					os.Exit(1)
				}
			}
		}()
		for {
			n, addr, err = conn.ReadFrom(buf)
			if n != 0 {
				printErr("# recv from", addr.String(), "leng =", n)
				fmt.Println((string)(buf[:n]))
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					printErr("Err when reading:", err)
					os.Exit(1)
				}
				return
			}
		}
	}
}

func formatBytes(buf []byte) string {
	s := make([]byte, len(buf)*3)
	for i, v := range buf {
		if i != 0 {
			s = append(s, ' ')
		}
		if v < 16 {
			s = append(s, '0')
		}
		s = strconv.AppendUint(s, (uint64)(v), 16)
	}
	return (string)(s)
}

func parseBytes(line string) (bts []byte, err error) {
	fields := strings.Fields(line)
	bts = make([]byte, len(fields))
	for i, v := range fields {
		var (
			b    uint64
			ok   bool
			base = 16
		)
		if v, ok = strings.CutPrefix(v, "0b"); ok {
			base = 2
		} else if v, ok = strings.CutPrefix(v, "0o"); ok {
			base = 8
		} else if v, ok = strings.CutPrefix(v, "0x"); ok {
			base = 16
		}
		b, err = strconv.ParseUint(v, base, 8)
		if err != nil {
			return
		}
		bts[i] = (byte)(b)
	}
	return
}

type netDialer struct{}

func (netDialer) Dial(network string, addr string) (c net.Conn, err error) {
	return net.Dial(network, addr)
}

func (netDialer) ListenPacket(network string, addr string) (c net.PacketConn, err error) {
	return net.ListenPacket(network, addr)
}

type stringAddr struct {
	network string
	addr    string
}

var _ net.Addr = (*stringAddr)(nil)

func (a *stringAddr) Network() string { return a.network }
func (a *stringAddr) String() string  { return a.addr }
