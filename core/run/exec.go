package run

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

type Run struct {
	LocalClients  map[string]Client
	RemoteClients map[string]Client
	Servers       []dao.Server
	Task          *dao.Task
	Config        dao.Config
}

type TaskContext struct {
	rIndex int
	cIndex int
	client Client
	dryRun bool

	desc     string
	name     string
	env      []string
	cmd      string
	numTasks int
}

func (run *Run) RunTask(
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) error {
	servers := run.Servers
	task := run.Task

	err := run.setKnownHostsFile(runFlags.KnownHostsFile)
	if err != nil {
		return err
	}

	configEnv, err := dao.EvaluateEnv(run.Config.Envs)
	if err != nil {
		return err
	}

	err = run.ParseTask(configEnv, userArgs, runFlags, setRunFlags)
	if err != nil {
		return err
	}
	run.CheckTaskNoColor()

	// Remote + Local clients
	numClients := len(servers) * 2
	clientCh := make(chan Client, numClients)
	errCh := make(chan ErrConnect, numClients)

	errConnect, err := run.SetClients(runFlags, numClients, clientCh, errCh)
	if err != nil {
		return err
	}

	if len(errConnect) > 0 {
		unreachableOutput := dao.TableOutput{
			Headers: []string{"server", "host", "user", "port", "error"},
			Rows:    []dao.Row{},
		}

		for _, u := range errConnect {
			unreachableOutput.Rows = append(unreachableOutput.Rows, dao.Row{Columns: []string{u.Name, u.Host, u.User, strconv.Itoa(int(u.Port)), u.Reason}})
		}

		options := print.PrintTableOptions{Theme: task.Theme, OmitEmpty: task.Spec.OmitEmpty, Output: task.Spec.Output, SuppressEmptyColumns: false}
		print.PrintTable("Unreachable Hosts", unreachableOutput.Rows, options, unreachableOutput.Headers[0:1], unreachableOutput.Headers[1:])

		if task.Spec.IgnoreUnreachable == false {
			return nil
		}
	}

	// Get reachable servers
	var reachableServers []dao.Server
	for _, server := range servers {
		if server.Local {
			reachableServers = append(reachableServers, server)
			continue
		}

		_, reachable := run.RemoteClients[server.Host]
		if reachable {
			reachableServers = append(reachableServers, server)
		}
	}
	run.Servers = reachableServers

	// Describe task
	if runFlags.Describe {
		print.PrintTaskBlock([]dao.Task{*task})
	}

	if runFlags.TTY || task.TTY {
		server, err := dao.GetFirstRemoteServer(servers)
		if err != nil {
			return err
		}

		envs := dao.MergeEnvs(userArgs, run.Task.Envs, server.Envs, configEnv)

		return ExecTTY(server.Host, server.User, server.Port, task.Cmd, envs)
	}

	switch task.Spec.Output {
	case "table", "html", "markdown":
		spinner := core.GetSpinner()
		spinner.Start(" Running", 500)

		data := run.Table(runFlags.DryRun)
		options := print.PrintTableOptions{Theme: task.Theme, OmitEmpty: task.Spec.OmitEmpty, Output: task.Spec.Output, SuppressEmptyColumns: false}
		run.CleanupClients()

		spinner.Stop()

		print.PrintTable("", data.Rows, options, data.Headers[0:1], data.Headers[1:])
	default:
		run.Text(runFlags.DryRun)
		run.CleanupClients()
	}

	if runFlags.Attach || task.Attach {
		server, err := dao.GetFirstRemoteServer(servers)
		if err != nil {
			return err
		}

		return SSHToServer(server.Host, server.User, server.Port, run.Config.DisableVerifyHost, run.Config.KnownHostsFile)
	}

	return nil
}

// SetClients establishes connection to server
// InitAuthMethod
// if identity_file, use that file
// if identity_file + passphrase, use that file with the passphrase
// if passphrase, use passphrase connect
// if nothing, attempt to use SSH Agent
func (run *Run) SetClients(
	runFlags *core.RunFlags,
	numChannels int,
	clientCh chan Client,
	errCh chan ErrConnect,
) ([]ErrConnect, error) {
	globalIdentityFile, globalPassword := getGlobalIdentity(runFlags)

	// Iterate through servers and create a singleton AuthMethod, which is used for
	// connecting to all hosts using a identity key. Servers which only use a password
	// are not included here and are handled separately.
	var identities []Identity
	for _, server := range run.Servers {
		if server.Local {
			continue
		}

		var pass string
		var err error
		if server.Password != nil {
			pass, err = dao.EvaluatePassword(*server.Password)
			if err != nil {
				errConnect := &ErrConnect{
					Name:   server.Name,
					User:   server.User,
					Host:   server.Host,
					Port:   server.Port,
					Reason: err.Error(),
				}
				return []ErrConnect{*errConnect}, nil
			}
		}

		identities = append(identities, Identity{
			IdentityFile: server.IdentityFile,
			Password:     &pass,
		})
	}

	// VerifyHost
	authMethod, err := InitAuthMethod(globalIdentityFile, globalPassword, identities)
	if err != nil {
		return []ErrConnect{}, err
	}

	createLocalClient := func(server dao.Server, wg *sync.WaitGroup, mu *sync.Mutex) {
		defer wg.Done()

		local := &LocalhostClient{
			User: server.User,
			Host: server.Host,
		}

		clientCh <- local
	}

	createRemoteClient := func(server dao.Server, wg *sync.WaitGroup, mu *sync.Mutex) {
		defer wg.Done()

		var auth ssh.AuthMethod
		if server.IdentityFile == nil && server.Password != nil {
			// Password only logic
			password, err := dao.EvaluatePassword(*server.Password)
			if err != nil {
				errConnect := &ErrConnect{
					Name:   server.Name,
					User:   server.User,
					Host:   server.Host,
					Port:   server.Port,
					Reason: err.Error(),
				}
				errCh <- *errConnect
			}

			passwordAuth := ssh.Password(password)
			auth = passwordAuth
		} else {
			// Identity key logic
			auth = authMethod
		}

		remote := &SSHClient{
			Name:       server.Name,
			User:       server.User,
			Host:       server.Host,
			Port:       server.Port,
			AuthMethod: auth,
		}

		if err := remote.Connect(run.Config.DisableVerifyHost, run.Config.KnownHostsFile, mu); err != nil {
			errCh <- *err
			return
		}

		clientCh <- remote
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, server := range run.Servers {
		wg.Add(1)
		go createLocalClient(server, &wg, &mu)

		if !server.Local {
			wg.Add(1)
			go createRemoteClient(server, &wg, &mu)
		}
	}
	wg.Wait()

	close(clientCh)
	close(errCh)

	localCLients := make(map[string]Client, numChannels/2)
	remoteClients := make(map[string]Client, numChannels/2)
	for client := range clientCh {
		switch client.(type) {
		case *SSHClient:
			remoteClients[client.GetHost()] = client
		case *LocalhostClient:
			localCLients[client.GetHost()] = client
		}
	}

	// Return on error
	var unreachable []ErrConnect
	for err := range errCh {
		unreachable = append(unreachable, err)
	}

	run.LocalClients = localCLients
	run.RemoteClients = remoteClients

	return unreachable, nil
}

func (run *Run) CleanupClients() {
	clients := run.RemoteClients
	var wg sync.WaitGroup

	trap := make(chan os.Signal, 1)
	signal.Notify(trap, os.Interrupt)
	go func() {
		for {
			select {
			case sig, ok := <-trap:
				if !ok {
					return
				}
				for _, c := range clients {
					err := c.Signal(sig)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v", err)
					}
				}
			}
		}
	}()
	wg.Wait()

	signal.Stop(trap)
	close(trap)

	// Close remote connections
	for _, c := range clients {
		if remote, ok := c.(*SSHClient); ok {
			remote.Close()
		}
	}
}

func (run *Run) ParseTask(configEnv []string, userArgs []string, runFlags *core.RunFlags, setRunFlags *core.SetRunFlags) error {
	// Update theme property if user flag is provided
	if runFlags.Theme != "" {
		theme, err := run.Config.GetTheme(runFlags.Theme)
		if err != nil {
			return err
		}

		run.Task.Theme = *theme
	}

	// Update output property if user flag is provided
	if runFlags.Output != "" {
		run.Task.Spec.Output = runFlags.Output
	}

	// Omit servers which provide empty output
	if setRunFlags.OmitEmpty {
		run.Task.Spec.OmitEmpty = runFlags.OmitEmpty
	}

	// If parallel flag is set to true, then update task specs
	if setRunFlags.Parallel {
		run.Task.Spec.Parallel = runFlags.Parallel
	}

	// If AnyErrorsFatal flag is set to true, then tasks execution will stop if error is encountered for all servers
	if setRunFlags.AnyErrorsFatal {
		run.Task.Spec.AnyErrorsFatal = runFlags.AnyErrorsFatal
	}

	// If IgnoreErrors flag is set to true, then tasks will run regardless of error
	if setRunFlags.IgnoreErrors {
		run.Task.Spec.IgnoreErrors = runFlags.IgnoreErrors
	}

	// If IgnoreErrors flag is set to true, then tasks will run regardless of error
	if setRunFlags.IgnoreUnreachable {
		run.Task.Spec.IgnoreUnreachable = runFlags.IgnoreUnreachable
	}

	// Update sub-commands
	for j := range run.Task.Tasks {
		// Set debug mode
		if runFlags.Debug {
			run.Task.Tasks[j].Cmd = "set -x;" + run.Task.Tasks[j].Cmd
		}

		// If local flag is set to true, then cmd will run locally, however, commands will retain their value
		if setRunFlags.Local {
			run.Task.Tasks[j].Local = runFlags.Local
		}

		envs, err := dao.ParseTaskEnv(run.Task.Tasks[j].Envs, userArgs, run.Task.Envs, configEnv)
		if err != nil {
			return err
		}
		run.Task.Tasks[j].Envs = envs
	}

	return nil
}

func (run *Run) CheckTaskNoColor() {
	for _, env := range run.Task.Envs {
		name := strings.Split(env, "=")[0]
		if name == "NO_COLOR" {
			text.DisableColors()
		}
	}
}

func SSHToServer(host string, user string, port uint8, disableVerifyHost bool, knownHostFile string) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}

	sshConnStr := fmt.Sprintf("%s@%s", user, host)
	portStr := fmt.Sprintf("-p %d", port)

	if !disableVerifyHost {
		knownHostFileStr := fmt.Sprintf("-o UserKnownHostsFile=%s", knownHostFile)
		err = unix.Exec(sshBin, []string{"ssh", "-t", sshConnStr, portStr, knownHostFileStr}, os.Environ())
	} else {
		err = unix.Exec(sshBin, []string{"ssh", "-t", sshConnStr, portStr}, os.Environ())
	}

	if err != nil {
		return err
	}

	return nil
}

func ExecTTY(host string, user string, port uint8, cmd string, envs []string) error {
	execBin, err := exec.LookPath("bash")
	if err != nil {
		return err
	}

	userEnv := append(os.Environ(), envs...)
	err = unix.Exec(execBin, []string{"bash", "-c", cmd}, userEnv)
	if err != nil {
		return err
	}

	return nil
}

func (run *Run) setKnownHostsFile(knownHostsFileFlag string) error {
	if knownHostsFileFlag != "" {
		run.Config.KnownHostsFile = knownHostsFileFlag
		return nil
	}

	value, found := os.LookupEnv("SAKE_KNOWN_HOSTS_FILE")
	if found {
		run.Config.KnownHostsFile = value
		return nil
	}

	return nil
}

func getGlobalIdentity(runFlags *core.RunFlags) (string, string) {
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

	return identityFile, password
}
