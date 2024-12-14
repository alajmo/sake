package run

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"

	"github.com/alajmo/sake/core/dao"
)

var ResetColor = "\033[0m"

// Client is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	conn *ssh.Client

	Name         string
	User         string
	Host         string
	Port         uint16
	IdentityFile string
	Password     string
	AuthMethod   []ssh.AuthMethod

	connString string
	connOpened bool

	Sessions []SSHSession
}

type SSHSession struct {
	sess         *ssh.Session
	remoteStdin  io.WriteCloser
	remoteStdout io.Reader
	remoteStderr io.Reader
	sessOpened   bool
	running      bool
}

type Identity struct {
	IdentityFile *string
	Password     *string
}

// SSHDialFunc can dial an ssh server and return a client
type SSHDialFunc func(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// Connect creates SSH connection to a specified host.
func (c *SSHClient) Connect(
	dialer SSHDialFunc,
	disableVerifyHost bool,
	knownHostsFile string,
	defaultTimeout uint,
	mu *sync.Mutex,
) *ErrConnect {
	return c.ConnectWith(dialer, disableVerifyHost, knownHostsFile, defaultTimeout, mu)
}

// ConnectWith creates a SSH connection to a specified host. It will use dialer to establish the
// connection.
func (c *SSHClient) ConnectWith(
	dialer SSHDialFunc,
	disableVerifyHost bool,
	knownHostsFile string,
	defaultTimeout uint,
	mu *sync.Mutex,
) *ErrConnect {
	if c.connOpened {
		return &ErrConnect{
			Name:   c.Name,
			User:   c.User,
			Host:   c.Host,
			Port:   c.Port,
			Reason: "Already connected",
		}
	}

	c.connString = net.JoinHostPort(c.Host, fmt.Sprint(c.Port))

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: c.AuthMethod,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if !disableVerifyHost {
				return VerifyHost(knownHostsFile, mu, hostname, remote, key)
			}
			return nil
		},
		Timeout: time.Duration(defaultTimeout) * time.Second,
	}

	var err error
	c.conn, err = dialer("tcp", c.connString, config)
	if err != nil {
		return &ErrConnect{
			Name:   c.Name,
			User:   c.User,
			Host:   c.Host,
			Port:   c.Port,
			Reason: err.Error(),
		}
	}
	c.connOpened = true

	return nil
}

// Run runs a command remotely on c.host.
func (c *SSHClient) Run(i int, env []string, workDir string, shell string, cmdStr string) error {
	// TODO: What to do about these?
	// if c.Sessions[i].running {
	// 	return fmt.Errorf("Session already running")
	// }
	// if c.Sessions[i].sessOpened {
	// 	return fmt.Errorf("Session already connected")
	// }

	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	c.Sessions[i].remoteStdin, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	c.Sessions[i].remoteStdout, err = sess.StdoutPipe()
	if err != nil {
		return err
	}

	c.Sessions[i].remoteStderr, err = sess.StderrPipe()
	if err != nil {
		return err
	}

	exportedEnv := AsExport(env)

	var cmdString string
	if workDir != "" {
		cmdString = fmt.Sprintf("cd %s; %s", workDir, exportedEnv)
	} else {
		cmdString = exportedEnv
	}

	if shell != "" {
		cmdString = fmt.Sprintf("%s %s '%s'", cmdString, shell, cmdStr)
	} else {
		cmdString = fmt.Sprintf("%s %s", cmdString, cmdStr)
	}

	// Start the remote command.
	if err := sess.Start(cmdString); err != nil {
		return err
	}

	c.Sessions[i].sess = sess
	c.Sessions[i].sessOpened = true
	c.Sessions[i].running = true

	return nil
}

// Wait waits until the remote command finishes and exits.
// It closes the SSH session.
func (c *SSHClient) Wait(i int) error {
	if !c.Sessions[i].running {
		return fmt.Errorf("trying to wait on stopped session")
	}

	err := c.Sessions[i].sess.Wait()
	c.Sessions[i].sess.Close()
	c.Sessions[i].running = false
	c.Sessions[i].sessOpened = false

	return err
}

// Close closes the underlying SSH connection and session.
func (c *SSHClient) Close(i int) error {
	if c.Sessions[i].sessOpened {
		c.Sessions[i].sess.Close()
		c.Sessions[i].sessOpened = false
	}
	if !c.connOpened {
		return fmt.Errorf("trying to close the already closed connection")
	}

	err := c.conn.Close()
	c.connOpened = false
	c.Sessions[i].running = false

	return err
}

func (c *SSHClient) Stdin(i int) io.WriteCloser {
	return c.Sessions[i].remoteStdin
}

func (c *SSHClient) Stderr(i int) io.Reader {
	return c.Sessions[i].remoteStderr
}

func (c *SSHClient) Stdout(i int) io.Reader {
	return c.Sessions[i].remoteStdout
}

// DialThrough will create a new connection from the ssh server c is connected to. DialThrough is an SSHDialer.
func (c *SSHClient) DialThrough(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := c.conn.Dial(net, addr)
	if err != nil {
		return nil, err
	}
	client, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(client, chans, reqs), nil
}

func (c *SSHClient) Prefix() (string, string, string, uint16) {
	return c.Name, c.Host, c.User, c.Port
}

func (c *SSHClient) Write(i int, p []byte) (n int, err error) {
	return c.Sessions[i].remoteStdin.Write(p)
}

func (c *SSHClient) WriteClose(i int) error {
	return c.Sessions[i].remoteStdin.Close()
}

func (c *SSHClient) Signal(i int, sig os.Signal) error {
	if !c.Sessions[i].sessOpened {
		return fmt.Errorf("session is not open")
	}

	switch sig {
	case os.Interrupt:
		return c.Sessions[i].sess.Signal(ssh.SIGINT)
	default:
		return fmt.Errorf("%v not supported", sig)
	}
}

func (c *SSHClient) GetName() string {
	return c.Name
}

// VerifyHost validates that the host is found in known_hosts file
func VerifyHost(knownHostsFile string, mu *sync.Mutex, host string, remote net.Addr, key ssh.PublicKey) error {
	// Return error if host not found or known host but key has changed
	hostFound, err := CheckKnownHost(host, remote, key, knownHostsFile)

	// Host in known hosts but key mismatch (possible man in the middle attack)
	if hostFound && err != nil {
		return err
	}

	// Host verified
	if hostFound && err == nil {
		return nil
	}

	// Host not found, ask user to check if he trust the host public key
	if !askIsHostTrusted(host, key, mu) {
		return errors.New("you typed no, aborted")
	}

	// Add the new host to known hosts file
	return AddKnownHost(host, key, knownHostsFile)
}

func CheckKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {
	var keyErr *knownhosts.KeyError

	// Get host key hostKeyCallback
	hostKeyCallback, err := knownhosts.New(knownFile)

	if err != nil {
		// TODO: if known_hosts malformed, return error to user
		// Need to check type of error, for instance: illegal base64 data at input byte 0
		return false, err
	}

	// TODO: For some reason hashed ip6 with port 22 does not work, all other combinations work
	err = hostKeyCallback(host, remote, key)

	// Known host already exists
	if err == nil {
		return true, nil
	}

	// If length of keyErr.Want is greater than 0, this means host has different key
	if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
		return true, keyErr
	}

	// Key not found in file and is therefor not trusted
	return false, nil
}

func askIsHostTrusted(host string, key ssh.PublicKey, mu *sync.Mutex) bool {
	mu.Lock()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type (y)es or (n)o: ")

	a, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	mu.Unlock()

	return strings.ToLower(strings.TrimSpace(a)) == "yes" || strings.ToLower(strings.TrimSpace(a)) == "y"
}

func AddKnownHost(host string, key ssh.PublicKey, knownFile string) (err error) {
	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	line := Line(host, key)
	_, err = f.WriteString(line + "\n")

	return err
}

// TODO: Replace this method with known_hosts Line method when issue with ip6 formats is fixed.
// Supported Host formats:
//
//	172.24.2.3
//	172.24.2.3:333 # custom port
//	2001:3984:3989::10
//	[2001:3984:3989::10]:333 # custom port
func Line(address string, key ssh.PublicKey) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "22"
	}

	if port != "22" {
		if strings.Contains(host, ":") {
			// ip6
			host = "[" + host + "]" + ":" + port
		} else {
			// ip4
			host = host + ":" + port
		}
	}

	var entry string
	hash := ssh_config.Get(host, "HashKnownHosts")
	if hash == "yes" {
		entry = knownhosts.HashHostname(host)
	} else {
		entry = host
	}

	return entry + " " + serialize(key)
}

func serialize(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal())
}

// Process all ENVs into a string of form
// Example output:
// export FOO="bar"; export BAR="baz";
func AsExport(env []string) string {
	exports := ``

	for _, v := range env {
		kv := strings.Split(v, "=")
		exports += `export ` + kv[0] + `="` + kv[1] + `";`
	}

	return exports
}

func GetSSHAgentSigners() ([]ssh.Signer, error) {
	// Load keys from SSH Agent if it's running
	sockPath, found := os.LookupEnv("SSH_AUTH_SOCK")
	if found {
		sock, err := net.Dial("unix", sockPath)
		if err != nil {
			return []ssh.Signer{}, err
		} else {
			agent := agent.NewClient(sock)
			s, err := agent.Signers()
			return s, err
		}
	}

	return []ssh.Signer{}, nil
}

func GetPasswordAuth(server dao.Server) (ssh.AuthMethod, error) {
	password, err := dao.EvaluatePassword(*server.Password)
	if err != nil {
		return nil, err
	}

	return ssh.Password(password), nil
}

// Password protected key
func GetPasswordIdentitySigner(server dao.Server) (ssh.Signer, error) {
	var signer ssh.Signer

	data, err := os.ReadFile(*server.IdentityFile)
	if err != nil {
		return nil, err
	}

	var pass *string
	pw, err := dao.EvaluatePassword(*server.Password)
	pass = &pw
	if err != nil {
		return nil, err
	}

	signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(*pass))
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func GetFingerprintPubKey(server dao.Server) (string, error) {
	data, err := os.ReadFile(*server.PubFile)
	if err != nil {
		return "", err
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return "", err
	}

	return ssh.FingerprintSHA256(pk), nil
}

func GetSigner(identityFile string) (ssh.Signer, error) {
	var signer ssh.Signer
	data, err := os.ReadFile(identityFile)
	if err != nil {
		return nil, err
	}

	signer, err = ssh.ParsePrivateKey(data)
	if err != nil {
		switch e := err.(type) {
		case *ssh.PassphraseMissingError:
			// TODO: Let user enter password 3 times, then fail
			fmt.Printf("Enter passphrase for %s: ", identityFile)
			pass, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return nil, err
			}

			signer, err = ssh.ParsePrivateKeyWithPassphrase(data, pass)
			if err != nil {
				return nil, err
			}
		default:
			return nil, e
		}
	}

	return signer, nil
}

func (c *SSHClient) Connected() bool {
	return c.connOpened
}
