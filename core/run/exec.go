package run

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"math"

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
	tty    bool

	desc     string
	name     string
	env      []string
	workDir  string
	shell    string
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

	errConnects, err := ParseServers(run.Config.SSHConfigFile, &run.Servers, runFlags)
	if err != nil {
		return err
	}

	if len(errConnects) > 0 {
		parseOutput := dao.TableOutput{
			Headers: []string{"server", "host", "user", "port", "error"},
			Rows:    []dao.Row{},
		}

		for _, u := range errConnects {
			parseOutput.Rows = append(parseOutput.Rows, dao.Row{Columns: []string{u.Name, u.Host, u.User, strconv.Itoa(int(u.Port)), u.Reason}})
		}

		options := print.PrintTableOptions{
			Theme:                task.Theme,
			OmitEmpty:            task.Spec.OmitEmpty,
			Output:               task.Spec.Output,
			SuppressEmptyColumns: false,
			Title:                "Parse Errors",
		}
		err = print.PrintTable(parseOutput.Rows, options, parseOutput.Headers)
		if err != nil {
			return err
		}

		return &core.ExecError{Err: errors.New("Parse Error"), ExitCode: 4}
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

	errConnect, err := run.SetClients(task, runFlags, numClients, clientCh, errCh)
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

		options := print.PrintTableOptions{
			Theme:                task.Theme,
			OmitEmpty:            task.Spec.OmitEmpty,
			Output:               "table",
			SuppressEmptyColumns: false,
			Title:                "\nUnreachable Hosts\n",
		}
		err := print.PrintTable(unreachableOutput.Rows, options, unreachableOutput.Headers)
		if err != nil {
			return err
		}

		if !task.Spec.IgnoreUnreachable {
			return &core.ExecError{Err: err, ExitCode: 4}
		}
	}

	// Get reachable servers
	var reachableServers []dao.Server
	for _, server := range servers {
		if server.Local {
			reachableServers = append(reachableServers, server)
			continue
		}

		_, reachable := run.RemoteClients[server.Name]
		if reachable {
			reachableServers = append(reachableServers, server)
		}
	}
	run.Servers = reachableServers

	// Describe task
	if runFlags.Describe {
		print.PrintTaskBlock([]dao.Task{*task})
	}

	switch task.Spec.Output {
	case "table", "table-1", "table-2", "table-3", "table-4", "html", "markdown", "json", "csv":
		spinner := core.GetSpinner()
		if !runFlags.Silent {
			spinner.Start(" Running", 500)
		}

		data, derr := run.Table(runFlags.DryRun)
		options := print.PrintTableOptions{
			Theme:                task.Theme,
			OmitEmpty:            task.Spec.OmitEmpty,
			Output:               task.Spec.Output,
			SuppressEmptyColumns: false,
			Resource:             "task",
		}
		run.CleanupClients()
		if !runFlags.Silent {
			spinner.Stop()
		}
		err = print.PrintTable(data.Rows, options, data.Headers)
		if err != nil {
			return err
		}

		if derr != nil {
			return derr
		}
	default:
		err := run.Text(runFlags.DryRun)
		run.CleanupClients()

		if err != nil {
			return err
		}
	}

	if runFlags.Attach || task.Attach {
		server, err := dao.GetFirstRemoteServer(servers)
		if err != nil {
			return err
		}

		return SSHToServer(server, run.Config.DisableVerifyHost, run.Config.KnownHostsFile)
	}

	return nil
}

type Signers struct {
	agentSigners []ssh.Signer
	fingerprints map[string]ssh.Signer     // fingerprint -> signer
	identities   map[string]ssh.Signer     // identityFile -> signer
	passwords    map[string]ssh.AuthMethod // password -> signer
}

// SetClients establishes connection to server
func (run *Run) SetClients(
	task *dao.Task,
	runFlags *core.RunFlags,
	numChannels int,
	clientCh chan Client,
	errCh chan ErrConnect,
) ([]ErrConnect, error) {
	createLocalClient := func(strategy string, numTasks int, server dao.Server, wg *sync.WaitGroup, mu *sync.Mutex) {
		defer wg.Done()

		local := &LocalhostClient{
			Name: server.Name,
			User: server.User,
			Host: server.Host,
		}

		switch strategy {
		case "free":
			for i := 0; i < numTasks; i++ {
				local.Sessions = append(local.Sessions, LocalSession{})
			}
		default:
			local.Sessions = append(local.Sessions, LocalSession{})
		}

		clientCh <- local
	}

	createRemoteClient := func(
		strategy string,
		numTasks int,
		authMethod []ssh.AuthMethod,
		server dao.Server,
		wg *sync.WaitGroup,
		mu *sync.Mutex,
	) {
		defer wg.Done()

		remote := &SSHClient{
			Name:       server.Name,
			User:       server.User,
			Host:       server.Host,
			Port:       server.Port,
			AuthMethod: authMethod,
		}
		// TODO: Create sessions if free strategy
		switch strategy {
		case "free":
			for i := 0; i < numTasks; i++ {
				remote.Sessions = append(remote.Sessions, SSHSession{})
			}
		default:
			remote.Sessions = append(remote.Sessions, SSHSession{})
		}

		var bastion *SSHClient
		if server.BastionHost != "" {
			bastion = &SSHClient{
				Host:       server.BastionHost,
				User:       server.BastionUser,
				Port:       server.BastionPort,
				AuthMethod: authMethod,
			}
			// Connect to bastion
			if err := bastion.Connect(ssh.Dial, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, mu); err != nil {
				errCh <- *err
				return
			}

			// Connect to server through bastion
			if err := remote.Connect(bastion.DialThrough, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, mu); err != nil {
				errCh <- *err
				return
			}
		} else {
			if err := remote.Connect(ssh.Dial, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, mu); err != nil {
				errCh <- *err
				return
			}
		}

		clientCh <- remote
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	agentSigners, err := GetSSHAgentSigners()
	if err != nil {
		return []ErrConnect{}, err
	}

	signers := Signers{
		agentSigners: agentSigners,
		fingerprints: make(map[string]ssh.Signer),
		identities:   make(map[string]ssh.Signer),
		passwords:    make(map[string]ssh.AuthMethod),
	}

	// Generate fingerprint (public key) for each agent key
	for _, s := range signers.agentSigners {
		fp := ssh.FingerprintSHA256(s.PublicKey())
		signers.fingerprints[fp] = s
	}

	for _, server := range run.Servers {
		err := populateSigners(server, &signers)
		if err != nil {
			return []ErrConnect{}, err
		}
	}

	// TODO: Dont create remote clients if task is set to local
	for _, server := range run.Servers {
		wg.Add(1)
		go createLocalClient(task.Spec.Strategy, len(task.Tasks), server, &wg, &mu)
		if !server.Local {
			wg.Add(1)
			authMethods := getAuthMethod(server, &signers)
			go createRemoteClient(task.Spec.Strategy, len(task.Tasks), authMethods, server, &wg, &mu)
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
			remoteClients[client.GetName()] = client
		case *LocalhostClient:
			localCLients[client.GetName()] = client
		}
	}

	run.LocalClients = localCLients
	run.RemoteClients = remoteClients

	var unreachable []ErrConnect
	for err := range errCh {
		unreachable = append(unreachable, err)
	}

	return unreachable, nil
}

func (run *Run) CleanupClients() {
	clients := run.RemoteClients
	var wg sync.WaitGroup

	trap := make(chan os.Signal, 1)
	signal.Notify(trap, os.Interrupt)
	go func() {
		for {
			sig, ok := <-trap
			if !ok {
				return
			}
			for _, c := range clients {
				switch c := c.(type) {
				case *SSHClient:
					for i := range c.Sessions {
						err := c.Signal(i, sig)
						if err != nil {
							fmt.Fprintf(os.Stderr, "%v", err)
						}
					}
				case *LocalhostClient:
					for i := range c.Sessions {
						err := c.Signal(i, sig)
						if err != nil {
							fmt.Fprintf(os.Stderr, "%v", err)
						}
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
			for i := range c.(*SSHClient).Sessions {
				remote.Close(i)
			}
		}
	}
}

// ParseServers resolves host, port, proxyjump in users ssh config
func ParseServers(sshConfigFile *string, servers *[]dao.Server, runFlags *core.RunFlags) ([]ErrConnect, error) {
	if runFlags.IdentityFile != "" {
		for i := range *servers {
			(*servers)[i].IdentityFile = &runFlags.IdentityFile
		}
	}

	if runFlags.Password != "" {
		for i := range *servers {
			(*servers)[i].Password = &runFlags.Password
		}
	}

	if sshConfigFile == nil {
		return nil, nil
	}

	cfg, err := core.ParseSSHConfig(*sshConfigFile)
	if err != nil {
		return nil, err
	}

	var errConnects []ErrConnect
	for i := range *servers {
		serv := cfg[(*servers)[i].Host]

		// Bastion resolve:
		//  1. proxyjump alias
		//	2. proxyjump
		//	3. bastion alias
		if proxyJump := serv.ProxyJump; proxyJump != "" {
			if hostName := cfg[proxyJump].HostName; hostName != "" {
				// 1. proxyjump alias
				(*servers)[i].BastionHost = hostName

				port := cfg[proxyJump].Port
				if port != "" {
					p, err := strconv.ParseUint(port, 10, 16)
					if err != nil {
						errConnect := &ErrConnect{
							Name:   (*servers)[i].Name,
							User:   (*servers)[i].User,
							Host:   (*servers)[i].Host,
							Port:   (*servers)[i].Port,
							Reason: err.Error(),
						}
						errConnects = append(errConnects, *errConnect)
						continue
					}
					(*servers)[i].BastionPort = uint16(p)
				} else {
					(*servers)[i].BastionPort = (*servers)[i].Port
				}

				user := cfg[proxyJump].User
				if user != "" {
					(*servers)[i].BastionUser = user
				} else {
					(*servers)[i].BastionUser = (*servers)[i].User
				}
			} else {
				// 2. proxyjump
				user, host, port, err := core.ParseHostName(proxyJump, (*servers)[i].User, (*servers)[i].Port)
				if err != nil {
					errConnect := &ErrConnect{
						Name:   (*servers)[i].Name,
						User:   (*servers)[i].User,
						Host:   (*servers)[i].Host,
						Port:   (*servers)[i].Port,
						Reason: err.Error(),
					}
					errConnects = append(errConnects, *errConnect)
					continue
				}

				(*servers)[i].BastionHost = host
				(*servers)[i].BastionPort = port
				(*servers)[i].BastionUser = user
			}
		} else if bastionHost := cfg[(*servers)[i].BastionHost].HostName; bastionHost != "" {
			// 3. bastion alias
			(*servers)[i].BastionHost = bastionHost

			port := cfg[(*servers)[i].BastionHost].Port
			if port != "" {
				p, err := strconv.ParseUint(port, 10, 16)
				if err != nil {
					errConnect := &ErrConnect{
						Name:   (*servers)[i].Name,
						User:   (*servers)[i].User,
						Host:   (*servers)[i].Host,
						Port:   (*servers)[i].Port,
						Reason: err.Error(),
					}
					errConnects = append(errConnects, *errConnect)
					continue
				}
				(*servers)[i].BastionPort = uint16(p)
			} else {
				(*servers)[i].BastionPort = (*servers)[i].Port
			}

			user := cfg[(*servers)[i].BastionHost].User
			if user != "" {
				(*servers)[i].BastionUser = user
			} else {
				(*servers)[i].BastionUser = (*servers)[i].User
			}
		}

		// IdentityFile
		if len(serv.IdentityFiles) > 0 {
			(*servers)[i].IdentityFile = &serv.IdentityFiles[0]
		}

		// HostName
		host := serv.HostName
		if host != "" {
			(*servers)[i].Host = host
		}

		// User
		user := serv.User
		if user != "" {
			(*servers)[i].User = user
		}

		// Port
		port := serv.Port
		if port != "" {
			p, err := strconv.ParseUint(port, 10, 16)
			if err != nil {
				errConnect := &ErrConnect{
					Name:   (*servers)[i].Name,
					User:   (*servers)[i].User,
					Host:   (*servers)[i].Host,
					Port:   (*servers)[i].Port,
					Reason: err.Error(),
				}
				errConnects = append(errConnects, *errConnect)
				continue
			}
			(*servers)[i].Port = uint16(p)
		}
	}

	return errConnects, err
}

func (run *Run) ParseTask(
	configEnv []string,
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) error {
	// Update theme property if user flag is provided
	if runFlags.Theme != "" {
		theme, err := run.Config.GetTheme(runFlags.Theme)
		if err != nil {
			return err
		}

		run.Task.Theme = *theme
	}

	if runFlags.Spec != "" {
		spec, err := run.Config.GetSpec(runFlags.Spec)
		if err != nil {
			return err
		}

		run.Task.Spec = *spec
	}

	if run.Task.Spec.Forks == 0 {
		run.Task.Spec.Forks = 10000
	}

	// Batch or BatchP must be > 0
	if run.Task.Spec.Batch == 0 && run.Task.Spec.BatchP == 0 {
		run.Task.Spec.Batch = uint32(len(run.Servers))
	} else if run.Task.Spec.BatchP > 0 {
		tot := float64(len(run.Servers))
		percentage := float64(run.Task.Spec.BatchP) / float64(100)
		batch := uint32(math.Floor(percentage * tot))

		if batch > 0 {
			run.Task.Spec.Batch = batch
		} else {
			run.Task.Spec.Batch = 1
		}
	}

	// Update strategy property if user flag is provided
	if runFlags.Strategy != "" {
		run.Task.Spec.Strategy = runFlags.Strategy
	}

	// Update output property if user flag is provided
	if runFlags.Output != "" {
		run.Task.Spec.Output = runFlags.Output
	}

	// Omit servers which provide empty output
	if setRunFlags.OmitEmpty {
		run.Task.Spec.OmitEmpty = runFlags.OmitEmpty
	}

	// If AnyErrorsFatal flag is set to true, then tasks execution will stop if error is encountered for all servers
	if setRunFlags.AnyErrorsFatal {
		run.Task.Spec.AnyErrorsFatal = runFlags.AnyErrorsFatal
	}

	if run.Task.Spec.AnyErrorsFatal {
		run.Task.Spec.MaxFailPercentage = 0
	}

	// If IgnoreErrors flag is set to true, then tasks will run regardless of error
	if setRunFlags.IgnoreErrors {
		run.Task.Spec.IgnoreErrors = runFlags.IgnoreErrors
	}

	// If IgnoreErrors flag is set to true, then tasks will run regardless of error
	if setRunFlags.IgnoreUnreachable {
		run.Task.Spec.IgnoreUnreachable = runFlags.IgnoreUnreachable
	}

	// If tty flag is set to true, then update task
	if setRunFlags.TTY {
		run.Task.TTY = runFlags.TTY
	}

	// Update sub-commands
	for j := range run.Task.Tasks {

		// If command name is not set, set one
		if run.Task.Tasks[j].Name == "" {
			run.Task.Tasks[j].Name = fmt.Sprintf("output-%d", j)
		}

		// If local flag is set to true, then cmd will run locally instead of on remote server
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

func getWorkDir(cmd dao.TaskCmd, server dao.Server) string {
	if cmd.Local || server.Local {
		rootDir := os.ExpandEnv(cmd.RootDir)
		if cmd.WorkDir != "" {
			workDir := os.ExpandEnv(cmd.WorkDir)
			if filepath.IsAbs(workDir) {
				return workDir
			} else {
				return filepath.Join(rootDir, workDir)
			}
		} else if server.WorkDir != "" {
			workDir := os.ExpandEnv(server.WorkDir)
			if filepath.IsAbs(workDir) {
				return workDir
			} else {
				return filepath.Join(rootDir, workDir)
			}
		} else {
			return rootDir
		}
	} else if cmd.WorkDir != "" {
		// task work_dir
		return cmd.WorkDir
	} else if server.WorkDir != "" {
		// server work_dir
		return server.WorkDir
	}

	return ""
}

func populateSigners(server dao.Server, signers *Signers) error {
	// If no identity or password provided, return
	if server.IdentityFile == nil && server.Password == nil {
		return nil
	}

	if server.IdentityFile != nil {
		// Check if identity_file already exists
		_, found := signers.identities[*server.IdentityFile]
		if found {
			return nil
		}
	} else if server.Password != nil {
		// Check if password already exists
		_, found := signers.passwords[*server.Password]
		if found {
			return nil
		}
	}

	// Check if exists in ssh-agent
	if server.PubFile != nil {
		fp, err := GetFingerprintPubKey(server)
		if err != nil {
			return err
		}

		v, found := signers.fingerprints[fp]
		if found {
			signers.identities[*server.IdentityFile] = v
			return nil
		}
	}

	// If only password provided -> assume password login
	if server.IdentityFile == nil && server.Password != nil {
		passAuthMethod, err := GetPasswordAuth(server)
		if err != nil {
			return err
		}
		signers.passwords[*server.Password] = passAuthMethod
		return nil
	} else if server.Password != nil {
		// If identity key + password -> assume password protected, populate and return
		signer, err := GetPasswordIdentitySigner(server)
		if err != nil {
			return err
		}
		signers.identities[*server.IdentityFile] = signer
		return nil
	} else {
		// If identity key -> try first without passphrase, if passphrase required prompt password, return
		signer, err := GetSigner(server)
		if err != nil {
			return err
		}
		signers.identities[*server.IdentityFile] = signer
		return nil
	}
}

func getAuthMethod(server dao.Server, signers *Signers) []ssh.AuthMethod {
	var authMethods []ssh.AuthMethod

	if server.IdentityFile != nil {
		v, found := signers.identities[*server.IdentityFile]
		if found {
			authMethods = append(authMethods, ssh.PublicKeys(v))
			return authMethods
		}
	}

	if server.Password != nil {
		v, found := signers.passwords[*server.Password]
		if found {
			authMethods = append(authMethods, v)
		}
	}

	// No signers found, use agent signers
	if len(signers.agentSigners) > 0 {
		authMethods = append(authMethods, ssh.PublicKeys(signers.agentSigners...))
	}

	return authMethods
}

func CalcFreeForks(batch int, tasks int, forks uint32) int {
	tot := batch * tasks
	if tot < int(forks) {
		return tot
	}
	return int(forks)
}

func CalcForks(batch int, forks uint32) int {
	if batch < int(forks) {
		return batch
	}
	return int(forks)
}
