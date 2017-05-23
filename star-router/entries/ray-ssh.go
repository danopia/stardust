package entries

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/danopia/stardust/star-router/base"
	"github.com/danopia/stardust/star-router/inmem"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Directory containing the clone function
func getRaySshDriver() base.Folder {
	return inmem.NewFolderOf("ray-ssh",
		inmem.NewFunction("invoke", raySshFunc),
		inmem.NewLink("input-shape", "/rom/shapes/function"),
	).Freeze()
}

// Function that creates a new ray shell when invoked
func raySshFunc(ctx base.Context, input base.Entry) (output base.Entry) {
	service := &raySsh{
		ctx:       ctx,
		rayFunc:   input.(base.Folder), // function shape
		tmpFolder: inmem.NewFolder("ray-ssh"),
	}

	service.configure()
	return service.tmpFolder
}

// Context for a running SSH server
type raySsh struct {
	ctx       base.Context
	sshConfig *ssh.ServerConfig
	listener  net.Listener
	rayFunc   base.Folder
	tmpFolder base.Folder
}

func (e *raySsh) configure() {
	e.sshConfig = &ssh.ServerConfig{
		PasswordCallback: e.checkPassword,
		NoClientAuth:     true,
	}

	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Println("Failed to load private key (./id_rsa)")
		return
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Println("Failed to parse private key")
		return
	}

	e.sshConfig.AddHostKey(private)
	e.start()
}

func (e *raySsh) checkPassword(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	log.Println("Login attempt for", c.User(), "-", string(pass))
	if c.User() == "star" && string(pass) == "dust" {
		return nil, nil
	}
	return nil, nil // fmt.Errorf("password rejected for %q", c.User())
}

func (e *raySsh) start() {
	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatalf("failed to listen for connection: %+v", err)
	}
	e.listener = listener

	log.Println("Listening for SSH on port 2022")
	go e.run()
}

func (e *raySsh) run() {
	for {
		// A ServerConn multiplexes several channels, which must
		// themselves be Accepted.
		tcpConn, err := e.listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, e.sshConfig)
		if err != nil {
			log.Println("Failed to perform SSH handshake", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		go ssh.DiscardRequests(reqs)
		go e.handleChannels(chans, fmt.Sprintf("%s", sshConn.RemoteAddr()))
	}
}

func (e *raySsh) handleChannels(chans <-chan ssh.NewChannel, addr string) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go e.handleChannel(newChannel, addr)
	}
}

func (e *raySsh) handleChannel(ch ssh.NewChannel, addr string) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := ch.ChannelType(); t != "session" {
		ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := ch.Accept()
	if err != nil {
		log.Println("Could not accept channel", err)
		return
	}

	rayFuncIvk, ok := e.rayFunc.Fetch("invoke")
	if !ok {
		panic("wat-a")
	}
	rayFunc, ok := rayFuncIvk.(base.Function)
	if !ok {
		panic("wat-b")
	}

	ray := rayFunc.Invoke(e.ctx, nil).(base.Folder)
	cmdEntry, ok := ray.Fetch("commands")
	if !ok {
		panic("wat0")
	}
	commands := cmdEntry.(base.Queue)
	e.tmpFolder.Put(addr, ray)

	outputEnt, ok := ray.Fetch("output")
	if !ok {
		panic("wat1")
	}
	outputLog := outputEnt.(base.Log)
	output := outputLog.Subscribe(nil)

	cwdEnt, ok := ray.Fetch("cwd")
	if !ok {
		panic("wat2")
	}
	cwd := cwdEnt.(base.String)

	term := terminal.NewTerminal(connection, "> ")

	go func() {
		defer connection.Close()
		defer func() {
			e.tmpFolder.Put(addr, nil)
		}()

		for {
			entry, ok := output.Next()
			if !ok {
				return
			}
			line, ok := entry.(base.String)
			if !ok {
				log.Println("Ray failed to get string from output", entry)
				return
			}

			term.Write([]byte(line.Get() + "\n"))
		}
	}()

	go func(requests <-chan *ssh.Request) {
		hasShell := false

		for req := range requests {
			var width, height int
			var ok bool

			switch req.Type {
			case "shell":
				if !hasShell {
					ok = true
					hasShell = true
				}
			case "pty-req":
				width, height, ok = parsePtyRequest(req.Payload)
				if ok {
					// TODO: Hardcode width to 100000?
					err := term.SetSize(width, height)
					ok = err == nil
				}
			case "window-change":
				width, height, ok = parseWinchRequest(req.Payload)
				if ok {
					// TODO: Hardcode width to 100000?
					err := term.SetSize(width, height)
					ok = err == nil
				}
			}

			if req.WantReply {
				req.Reply(ok, nil)
			}
		}
	}(requests)

	for {
		term.Write([]byte("\n"))
		term.SetPrompt(fmt.Sprintf("%s $ ", cwd.Get()))

		line, err := term.ReadLine()
		if err == io.EOF {
			commands.Close()
			log.Println("Client disconnected", addr)
			return
		}
		if err != nil {
			log.Println("handleChannel readLine err:", err)
			continue
		}
		if len(line) > 0 {
			commands.Push(inmem.NewString("ssh-command", line))
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Helpers below are borrowed from go.crypto circa 2011:

// parsePtyRequest parses the payload of the pty-req message and extracts the
// dimensions of the terminal. See RFC 4254, section 6.2.
func parsePtyRequest(s []byte) (width, height int, ok bool) {
	_, s, ok = parseString(s)
	if !ok {
		return
	}
	width32, s, ok := parseUint32(s)
	if !ok {
		return
	}
	height32, _, ok := parseUint32(s)
	width = int(width32)
	height = int(height32)
	if width < 1 {
		ok = false
	}
	if height < 1 {
		ok = false
	}
	return
}

func parseWinchRequest(s []byte) (width, height int, ok bool) {
	width32, _, ok := parseUint32(s)
	if !ok {
		return
	}
	height32, _, ok := parseUint32(s)
	if !ok {
		return
	}

	width = int(width32)
	height = int(height32)
	if width < 1 {
		ok = false
	}
	if height < 1 {
		ok = false
	}
	return
}

func parseString(in []byte) (out string, rest []byte, ok bool) {
	if len(in) < 4 {
		return
	}
	length := binary.BigEndian.Uint32(in)
	if uint32(len(in)) < 4+length {
		return
	}
	out = string(in[4 : 4+length])
	rest = in[4+length:]
	ok = true
	return
}

func parseUint32(in []byte) (uint32, []byte, bool) {
	if len(in) < 4 {
		return 0, nil, false
	}
	return binary.BigEndian.Uint32(in), in[4:], true
}
