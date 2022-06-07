package run

// Source: https://github.com/pressly/sup/blob/be6dff41589b713547415b72660885dd7a045f8f/sup.go

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

var ResetColor = "\033[0m"
var DefaultTimeout = 20 * time.Second

// Client is a wrapper over the SSH connection/sessions.
type SSHClient struct {
	conn *ssh.Client
	sess *ssh.Session

	Name         string
	User         string
	Host         string
	Port         uint8
	IdentityFile string
	Password     string
	AuthMethod   ssh.AuthMethod

	connString   string
	remoteStdin  io.WriteCloser
	remoteStdout io.Reader
	remoteStderr io.Reader
	connOpened   bool
	sessOpened   bool
	running      bool
}

type Identity struct {
	IdentityFile *string
	Password     *string
}

// InitAuthMethod initiates SSH authentication method.
// if identity_file, use that file
// if identity_file + passphrase, use that file with the passphrase
// if passphrase, use passphrase connect
// if nothing, attempt to use SSH Agent
func InitAuthMethod(globalIdentityFile string, globalPassword string, identities []Identity) (ssh.AuthMethod, error) {
	var signers []ssh.Signer

	// Load keys from SSH Agent if it's running
	sockPath, found := os.LookupEnv("SSH_AUTH_SOCK")
	if found {
		sock, err := net.Dial("unix", sockPath)
		if err != nil {
			return ssh.PublicKeys(), err
		} else {
			agent := agent.NewClient(sock)
			signers, _ = agent.Signers()
		}
	}

	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		agent := agent.NewClient(sock)
		signers, _ = agent.Signers()
	}

	// User provides global identity/password via flag/env
	if globalIdentityFile != "" {
		data, err := ioutil.ReadFile(globalIdentityFile)
		if err != nil {
			return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", globalIdentityFile, err)
		}

		if globalPassword != "" {
			signer, err := ssh.ParsePrivateKeyWithPassphrase(data, []byte(globalPassword))
			if err != nil {
				return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", globalIdentityFile, err)
			}

			return ssh.PublicKeys(signer), nil
		} else {
			signer, err := ssh.ParsePrivateKey(data)
			if err != nil {
				return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", globalIdentityFile, err)
			}

			return ssh.PublicKeys(signer), nil
		}
	}

	// User provides identity/passphrase via config
	for _, identity := range identities {
		var signer ssh.Signer

		if identity.IdentityFile != nil {
			// Identity IdentityFile
			data, err := ioutil.ReadFile(*identity.IdentityFile)
			if err != nil {
				return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", *identity.IdentityFile, err)
			}

			if identity.Password != nil {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(*identity.Password))
				if err != nil {
					return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", *identity.IdentityFile, err)
				}
			} else {
				signer, err = ssh.ParsePrivateKey(data)
				if err != nil {
					return ssh.PublicKeys(), fmt.Errorf("failed to parse `%s`\n  %w", *identity.IdentityFile, err)
				}
			}

			signers = append(signers, signer)
		}
	}

	return ssh.PublicKeys(signers...), nil
}

// SSHDialFunc can dial an ssh server and return a client
type SSHDialFunc func(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error)

// Connect creates SSH connection to a specified host.
// It expects the host of the form "[ssh://]host[:port]".
func (c *SSHClient) Connect(disableVerifyHost bool, knownHostsFile string, mu *sync.Mutex) *ErrConnect {
	return c.ConnectWith(ssh.Dial, disableVerifyHost, knownHostsFile, mu)
}

// ConnectWith creates a SSH connection to a specified host. It will use dialer to establish the
// connection.
func (c *SSHClient) ConnectWith(dialer SSHDialFunc, disableVerifyHost bool, knownHostsFile string, mu *sync.Mutex) *ErrConnect {
	if c.connOpened {
		return &ErrConnect{
			Name:   c.Name,
			User:   c.User,
			Host:   c.Host,
			Port:   c.Port,
			Reason: "Already connected",
		}
	}

	c.connString = fmt.Sprintf("%s:%d", c.Host, c.Port)

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			c.AuthMethod,
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if !disableVerifyHost {
				return VerifyHost(knownHostsFile, mu, hostname, remote, key)
			}
			return nil
		},
		Timeout: DefaultTimeout,
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
func (c *SSHClient) Run(env []string, cmdStr string) error {
	if c.running {
		return fmt.Errorf("Session already running")
	}
	if c.sessOpened {
		return fmt.Errorf("Session already connected")
	}

	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}

	c.remoteStdin, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	c.remoteStdout, err = sess.StdoutPipe()
	if err != nil {
		return err
	}

	c.remoteStderr, err = sess.StderrPipe()
	if err != nil {
		return err
	}

	exportedEnv := AsExport(env)

	// Start the remote command.
	if err := sess.Start(exportedEnv + cmdStr); err != nil {
		return err
	}

	c.sess = sess
	c.sessOpened = true
	c.running = true

	return nil
}

// Wait waits until the remote command finishes and exits.
// It closes the SSH session.
func (c *SSHClient) Wait() error {
	if !c.running {
		return fmt.Errorf("Trying to wait on stopped session")
	}

	err := c.sess.Wait()
	c.sess.Close()
	c.running = false
	c.sessOpened = false

	return err
}

// DialThrough will create a new connection from the ssh server sc is connected to. DialThrough is an SSHDialer.
func (sc *SSHClient) DialThrough(net, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	conn, err := sc.conn.Dial(net, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// Close closes the underlying SSH connection and session.
func (c *SSHClient) Close() error {
	if c.sessOpened {
		c.sess.Close()
		c.sessOpened = false
	}
	if !c.connOpened {
		return fmt.Errorf("Trying to close the already closed connection")
	}

	err := c.conn.Close()
	c.connOpened = false
	c.running = false

	return err
}

func (c *SSHClient) Stdin() io.WriteCloser {
	return c.remoteStdin
}

func (c *SSHClient) Stderr() io.Reader {
	return c.remoteStderr
}

func (c *SSHClient) Stdout() io.Reader {
	return c.remoteStdout
}

func (c *SSHClient) Prefix() string {
	return c.Host
}

func (c *SSHClient) Write(p []byte) (n int, err error) {
	return c.remoteStdin.Write(p)
}

func (c *SSHClient) WriteClose() error {
	return c.remoteStdin.Close()
}

func (c *SSHClient) Signal(sig os.Signal) error {
	if !c.sessOpened {
		return fmt.Errorf("session is not open")
	}

	switch sig {
	case os.Interrupt:
		return c.sess.Signal(ssh.SIGINT)
	default:
		return fmt.Errorf("%v not supported", sig)
	}
}

func (c *SSHClient) GetHost() string {
	return c.Host
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
		return errors.New("you typed no, aborted!")
	}

	// Add the new host to known hosts file
	return AddKnownHost(host, remote, key, knownHostsFile)
}

func CheckKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {
	var keyErr *knownhosts.KeyError

	// Get host key hostKeyCallback
	hostKeyCallback, err := knownhosts.New(knownFile)

	if err != nil {
		return false, err
	}

	// check if host already exists
	err = hostKeyCallback(host, remote, key)

	// Known host already exists
	if err == nil {
		return true, nil
	}

	// If length of keyErr.Want is greater than 0, this means host has different key
	if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
		return true, keyErr
	}

	// Some other error occurred and safest way to handle is to pass it back to user
	if err != nil {
		return false, err
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

func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (err error) {
	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	remoteNormalized := knownhosts.Normalize(remote.String())
	hostNormalized := knownhosts.Normalize(host)
	addresses := []string{remoteNormalized}

	if hostNormalized != remoteNormalized {
		addresses = append(addresses, hostNormalized)
	}

	_, err = f.WriteString(knownhosts.Line(addresses, key) + "\n")

	return err
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
