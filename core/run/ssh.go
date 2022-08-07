package run

// Source: https://github.com/pressly/sup/blob/be6dff41589b713547415b72660885dd7a045f8f/sup.go

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/user"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
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
	Port         uint16
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

func GetGlobalIdentitySigner(runFlags *core.RunFlags) (ssh.Signer, error) {
	var identityFile string
	var password string

	if runFlags.IdentityFile != "" {
		identityFile = runFlags.IdentityFile
	} else {
		value, found := os.LookupEnv("SAKE_IDENTITY_FILE")
		if found {
			if strings.HasPrefix(value, "~/") {
				usr, err := user.Current()
				if err != nil {
					panic(err)
				}
				dir := usr.HomeDir
				identityFile = filepath.Join(dir, value[2:])
			} else {
				identityFile = value
			}
		}
	}

	if runFlags.Password != "" {
		password = runFlags.Password
	} else {
		value, found := os.LookupEnv("SAKE_PASSWORD")
		if found {
			password = value
		}
	}

	// User provides global identity/password via flag/env
	if identityFile != "" {
		data, err := ioutil.ReadFile(identityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse `%s`\n  %w", identityFile, err)
		}

		if password != "" {
			signer, err := ssh.ParsePrivateKeyWithPassphrase(data, []byte(password))
			if err != nil {
				return nil, fmt.Errorf("failed to parse `%s`\n  %w", identityFile, err)
			}

			return signer, nil
		} else {
			signer, err := ssh.ParsePrivateKey(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse `%s`\n  %w", identityFile, err)
			}

			return signer, nil
		}
	}

	return nil, nil
}

func GetIdentitySigner(server dao.Server) (ssh.Signer, *ErrConnect) {
	var pass *string
	if server.Password != nil {
		pw, err := dao.EvaluatePassword(*server.Password)
		pass = &pw
		if err != nil {
			errConnect := &ErrConnect{
				Name:   server.Name,
				User:   server.User,
				Host:   server.Host,
				Port:   server.Port,
				Reason: err.Error(),
			}
			return nil, errConnect
		}
	}

	var signer ssh.Signer
	if server.IdentityFile != nil {
		// Identity IdentityFile
		data, err := ioutil.ReadFile(*server.IdentityFile)
		if err != nil {
			errConnect := &ErrConnect{
				Name:   server.Name,
				User:   server.User,
				Host:   server.Host,
				Port:   server.Port,
				Reason: fmt.Errorf("failed to parse `%s`\n  %w", *server.IdentityFile, err).Error(),
			}
			return nil, errConnect
		}

		if pass != nil {
			// Password protected key
			signer, err = ssh.ParsePrivateKeyWithPassphrase(data, []byte(*pass))
			if err != nil {
				errConnect := &ErrConnect{
					Name:   server.Name,
					User:   server.User,
					Host:   server.Host,
					Port:   server.Port,
					Reason: fmt.Errorf("failed to parse `%s`\n  %w", *server.IdentityFile, err).Error(),
				}
				return nil, errConnect
			}
		} else {
			// Unprotected key
			signer, err = ssh.ParsePrivateKey(data)
			if err != nil {
				errConnect := &ErrConnect{
					Name:   server.Name,
					User:   server.User,
					Host:   server.Host,
					Port:   server.Port,
					Reason: fmt.Errorf("failed to parse `%s`\n  %w", *server.IdentityFile, err).Error(),
				}
				return nil, errConnect
			}
		}

		return signer, nil
	}

	return nil, nil
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

	c.connString = net.JoinHostPort(c.Host, fmt.Sprint(c.Port))

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
func (c *SSHClient) Run(env []string, workDir string, shell string, cmdStr string) error {
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

// TODO: Refactor this
func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (err error) {
	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	remoteNormalized := knownhosts.Normalize(remote.String())
	hostNormalized := knownhosts.Normalize(host)
	r, _, err := net.SplitHostPort(remote.String())
	if err != nil {
		return err
	}

	addresses := []string{r}

	// If it's a hostname, then add the resolved IP as well
	// For instance, user specifies:
	// host: server-1.lan
	// Then the complete line in known_hosts will be:
	// x.x.x.x,server-1.lan
	if hostNormalized != remoteNormalized {
		h, _, err := net.SplitHostPort(host)
		if err != nil {
			return err
		}
		addresses = append(addresses, h)
	}

	// Note: knownhosts.Line will append square brackets [] to IP6 addresses
	line := Line(addresses, key)
	_, err = f.WriteString(line + "\n")

	return err
}

func Line(addresses []string, key ssh.PublicKey) string {
	var trimmed []string
	trimmed = append(trimmed, addresses...)

	return strings.Join(trimmed, ",") + " " + serialize(key)
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
