package run

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

type Run struct {
	LocalClients       map[string]Client
	RemoteClients      map[string]Client
	Servers            []dao.Server
	UnreachableServers []dao.Server
	Task               *dao.Task
	Config             dao.Config
}

type TaskContext struct {
	rIndex int
	cIndex int
	client Client
	dryRun bool
	tty    bool
	print  string

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

	err = run.ParseTask(configEnv, userArgs, runFlags, setRunFlags)
	if err != nil {
		return err
	}
	run.CheckTaskNoColor()

	errConnects, err := ParseServers(run.Config.SSHConfigFile, &run.Servers, runFlags, run.Task.Spec.Order)
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
			Theme:            task.Theme,
			OmitEmptyRows:    task.Spec.OmitEmptyRows,
			OmitEmptyColumns: false,
			Output:           task.Spec.Output,
			Title:            "Parse Errors",
		}
		err = print.PrintTable(parseOutput.Rows, options, parseOutput.Headers, []string{}, true, true)
		if err != nil {
			return err
		}

		return &core.ExecError{Err: errors.New("parse Error"), ExitCode: 4}
	}

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
			Theme:            task.Theme,
			OmitEmptyRows:    task.Spec.OmitEmptyRows,
			OmitEmptyColumns: false,
			Output:           "table",
			Title:            "\nUnreachable Hosts\n",
		}
		err := print.PrintTable(unreachableOutput.Rows, options, unreachableOutput.Headers, []string{}, true, true)
		if err != nil {
			return err
		}

		if !task.Spec.IgnoreUnreachable {
			return &core.ExecError{Err: err, ExitCode: 4}
		}
	}

	// Get reachable servers
	var reachableServers []dao.Server
	var unreachableServers []dao.Server
	for _, server := range servers {
		if server.Local {
			reachableServers = append(reachableServers, server)
			continue
		}

		_, reachable := run.RemoteClients[server.Name]
		if reachable {
			reachableServers = append(reachableServers, server)
		} else {
			unreachableServers = append(unreachableServers, server)
		}
	}
	run.Servers = reachableServers
	run.UnreachableServers = unreachableServers

	// Describe task
	if task.Spec.Describe {
		PrintHeader("TASK DESCRIPTION ", run.Task.Theme.Text, false)
		print.PrintTaskBlock([]dao.Task{*task})
	}

	// Describe Servers
	if task.Spec.ListHosts {
		PrintHeader("HOSTS ", run.Task.Theme.Text, false)
		err := print.PrintServerList(servers)
		if err != nil {
			return err
		}
	}

	if runFlags.Confirm && !confirmExecute(run.Task.Name) {
		return nil
	}

	switch task.Spec.Output {
	case "table", "table-1", "table-2", "table-3", "table-4", "html", "markdown", "json", "csv", "none":
		spinner := core.GetSpinner()
		if !task.Spec.Silent && !task.Spec.Step && !task.Spec.Confirm {
			spinner.Start(" Running", 500)
		}

		data, reportData, derr := run.Table(runFlags.DryRun)
		options := print.PrintTableOptions{
			Theme:            task.Theme,
			OmitEmptyRows:    task.Spec.OmitEmptyRows,
			OmitEmptyColumns: task.Spec.OmitEmptyColumns,
			Output:           task.Spec.Output,
			Resource:         "task",
		}
		run.CleanupClients()
		if !task.Spec.Silent && !task.Spec.Step && !task.Spec.Confirm {
			spinner.Stop()
		}

		if len(run.Servers) > 0 && task.Spec.Output != "none" {
			if strings.Contains(task.Spec.Output, "table") {
				PrintHeader("TASKS ", run.Task.Theme.Text, true)
			}

			err = print.PrintTable(data.Rows, options, data.Headers, []string{}, false, false)
			if err != nil {
				return err
			}
		}

		err := print.PrintReport(&run.Task.Theme, reportData, task.Spec)
		if err != nil {
			return err
		}

		if derr != nil {
			return derr
		}
	default:
		if (len(run.Servers) > 0 && len(run.Task.Tasks) > 1) || run.Task.Spec.Strategy != "linear" {
			PrintHeader("TASKS ", run.Task.Theme.Text, true)
		} else {
			fmt.Println()
		}

		reportData, derr := run.Text(runFlags.DryRun)

		run.CleanupClients()

		err = print.PrintReport(&run.Task.Theme, reportData, task.Spec)
		if err != nil {
			return err
		}

		if derr != nil {
			return derr
		}
	}

	if task.Attach {
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
	createLocalClient := func(strategy string, numTasks int, server dao.Server, wg *sync.WaitGroup) {
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
		switch strategy {
		case "free":
			for i := 0; i < numTasks; i++ {
				remote.Sessions = append(remote.Sessions, SSHSession{})
			}
		default:
			remote.Sessions = append(remote.Sessions, SSHSession{})
		}

		if len(server.Bastions) > 0 {
			var bastion *SSHClient
			for _, bastionServer := range server.Bastions {
				if bastion == nil {
					bastion = &SSHClient{
						Name:       "Bastion",
						Host:       bastionServer.Host,
						User:       bastionServer.User,
						Port:       bastionServer.Port,
						AuthMethod: authMethod,
					}
					if err := bastion.Connect(ssh.Dial, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, run.Config.DefaultTimeout, mu); err != nil {
						errCh <- *err
						return
					}
				} else {
					bastt := &SSHClient{
						Name:       "Bastion",
						Host:       bastionServer.Host,
						User:       bastionServer.User,
						Port:       bastionServer.Port,
						AuthMethod: authMethod,
					}

					if err := bastt.Connect(bastion.DialThrough, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, run.Config.DefaultTimeout, mu); err != nil {
						errCh <- *err
						return
					}
					bastion = bastt
				}
			}
			if err := remote.Connect(bastion.DialThrough, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, run.Config.DefaultTimeout, mu); err != nil {
				errCh <- *err
				return
			}
		} else {
			if err := remote.Connect(ssh.Dial, run.Config.DisableVerifyHost, run.Config.KnownHostsFile, run.Config.DefaultTimeout, mu); err != nil {
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
		go createLocalClient(task.Spec.Strategy, len(task.Tasks), server, &wg)
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

// ParseServers resolves host, port, proxyjump in user ssh config
func ParseServers(
	sshConfigFile *string,
	servers *[]dao.Server,
	runFlags *core.RunFlags,
	order string,
) ([]ErrConnect, error) {
	dao.SortServers(order, servers)

	if runFlags.IdentityFile != "" {
		for i := range *servers {
			(*servers)[i].IdentityFile = &runFlags.IdentityFile
		}
	}

	if runFlags.User != "" {
		for i := range *servers {
			(*servers)[i].User = runFlags.User
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
		// Bastion resolve, for instance, host: server-1 has an entry in ssh config
		// that has ProxyJump, ProxyJump alias or if in sake it has a bastion: server-1
		//  1. proxyjump alias
		//	2. proxyjump
		//	3. bastion alias, in this case we need to handle multiple bastions
		// In-case sake has bastions defined, then skip resolving
		// TODO: Refactor this part
		if proxyJump := serv.ProxyJump; proxyJump != "" && len((*servers)[i].Bastions) == 0 {
			if hostName := cfg[proxyJump].HostName; hostName != "" {
				// 1. proxyjump alias
				bastionHost := hostName
				bastionPort := (*servers)[i].Port
				bastionUser := (*servers)[i].User
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
					bastionPort = uint16(p)
				}

				user := cfg[proxyJump].User
				if user != "" {
					bastionUser = user
				}

				bastion := dao.Bastion{
					Host: bastionHost,
					User: bastionUser,
					Port: bastionPort,
				}

				(*servers)[i].Bastions = append((*servers)[i].Bastions, bastion)
			} else {
				// 2. proxyjump

				for _, proxy := range strings.Split(proxyJump, ",") {
					user, host, port, err := core.ParseHostName(proxy, (*servers)[i].User, (*servers)[i].Port)
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

					bastion := dao.Bastion{
						Host: host,
						User: user,
						Port: port,
					}
					(*servers)[i].Bastions = append((*servers)[i].Bastions, bastion)
				}
			}
		} else {
			// 3. bastion alias
			for j, bastion := range (*servers)[i].Bastions {
				if bastionHost := cfg[bastion.Host].HostName; bastionHost != "" {
					bastionPort := (*servers)[i].Port
					bastionUser := (*servers)[i].User
					if cfg[bastion.Host].Port != "" {
						p, err := strconv.ParseUint(cfg[bastion.Host].Port, 10, 16)
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
						bastionPort = uint16(p)
					}

					if cfg[bastion.Host].User != "" {
						bastionUser = cfg[bastion.Host].User
					}

					(*servers)[i].Bastions[j].Host = bastionHost
					(*servers)[i].Bastions[j].User = bastionUser
					(*servers)[i].Bastions[j].Port = bastionPort
				}
			}
		}

		// IdentityFile
		if len(serv.IdentityFiles) > 0 {
			iFile, err := core.ExpandPath(serv.IdentityFiles[0])
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

			(*servers)[i].IdentityFile = &iFile

			// TODO: Update PubFile as well
			if _, err := os.Stat(*(*servers)[i].IdentityFile); errors.Is(err, os.ErrNotExist) {
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

			pubFile := *(*servers)[i].IdentityFile + ".pub"
			if _, err := os.Stat(pubFile); errors.Is(err, os.ErrNotExist) {
				errConnect := &ErrConnect{
					Name:   (*servers)[i].Name,
					User:   (*servers)[i].User,
					Host:   (*servers)[i].Host,
					Port:   (*servers)[i].Port,
					Reason: err.Error(),
				}
				errConnects = append(errConnects, *errConnect)
				continue
			} else {
				*(*servers)[i].PubFile = pubFile
			}
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

	if setRunFlags.Forks {
		run.Task.Spec.Forks = runFlags.Forks
	}

	if setRunFlags.Batch { // Flag
		run.Task.Spec.Batch = runFlags.Batch
	} else if setRunFlags.BatchP { // Flag
		tot := float64(len(run.Servers))
		percentage := float64(runFlags.BatchP) / float64(100)
		batch := uint32(math.Floor(percentage * tot))

		if batch > 0 {
			run.Task.Spec.Batch = batch
		} else {
			run.Task.Spec.Batch = 1
		}
	} else { // Spec
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
	}

	if setRunFlags.Order {
		run.Task.Spec.Order = runFlags.Order
	}

	// Report
	if setRunFlags.Report {
		run.Task.Spec.Report = runFlags.Report
	}

	// Update describe property if user flag is provided
	if setRunFlags.Describe {
		run.Task.Spec.Describe = runFlags.Describe
	}

	// Update describe property if user flag is provided
	if setRunFlags.ListHosts {
		run.Task.Spec.ListHosts = runFlags.ListHosts
	}

	// Update describe property if user flag is provided
	if setRunFlags.Silent {
		run.Task.Spec.Silent = runFlags.Silent
	}

	// Update describe property if user flag is provided
	if setRunFlags.Attach {
		run.Task.Attach = runFlags.Attach
	}

	// Update strategy property if user flag is provided
	if runFlags.Strategy != "" {
		run.Task.Spec.Strategy = runFlags.Strategy
	}

	// Update output property if user flag is provided
	if runFlags.Output != "" {
		run.Task.Spec.Output = runFlags.Output
	}

	if runFlags.Print != "" {
		run.Task.Spec.Print = runFlags.Print
	}

	// Omit empty row
	if setRunFlags.OmitEmptyRows {
		run.Task.Spec.OmitEmptyRows = runFlags.OmitEmptyRows
	}

	// Omit empty column
	if setRunFlags.OmitEmptyColumns {
		run.Task.Spec.OmitEmptyColumns = runFlags.OmitEmptyColumns
	}

	// If AnyErrorsFatal flag is set to true, then tasks execution will stop if error is encountered for all servers
	if setRunFlags.AnyErrorsFatal {
		run.Task.Spec.AnyErrorsFatal = runFlags.AnyErrorsFatal

		if run.Task.Spec.AnyErrorsFatal {
			run.Task.Spec.MaxFailPercentage = 0
		} else {
			run.Task.Spec.MaxFailPercentage = 100
		}
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

	// Confirm
	if setRunFlags.Confirm {
		run.Task.Spec.Confirm = runFlags.Confirm
	}

	if setRunFlags.Step {
		run.Task.Spec.Step = runFlags.Step
	}

	// Update sub-commands
	for j := range run.Task.Tasks {

		// If command name is not set, set one
		if run.Task.Tasks[j].Name == "" {
			run.Task.Tasks[j].Name = fmt.Sprintf("task-%d", j)
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

	run.ParseTaskTarget(runFlags, setRunFlags)

	if setRunFlags.Verbose || run.Task.Spec.Verbose {
		run.Task.Spec.Describe = true
		run.Task.Spec.ListHosts = true
		run.Task.Spec.Report = []string{"all"}
	}

	return nil
}

func (run *Run) ParseTaskTarget(
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) {
	if setRunFlags.All {
		run.Task.Target.All = runFlags.All
	}
	if setRunFlags.Servers {
		run.Task.Target.Servers = runFlags.Servers
	}
	if setRunFlags.Tags {
		run.Task.Target.Tags = runFlags.Tags
	}
	if setRunFlags.Regex {
		run.Task.Target.Regex = runFlags.Regex
	}
	if setRunFlags.Invert {
		run.Task.Target.Invert = runFlags.Invert
	}
	if setRunFlags.Limit {
		run.Task.Target.Limit = runFlags.Limit
	}
	if setRunFlags.LimitP {
		run.Task.Target.LimitP = runFlags.LimitP
	}
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

func getWorkDir(
	cmdLocal bool,
	serverLocal bool,
	cmdWD string,
	serverWD string,
	cmdDir string,
	serverDir string,
) string {
	var cmdWDTrue bool
	if cmdWD != "" {
		cmdWDTrue = true
	}

	var serverWDTrue bool
	if serverWD != "" {
		serverWDTrue = true
	}

	// Remote

	if !cmdLocal && !serverLocal {
		if !cmdWDTrue && !serverWDTrue {
			return ""
		}

		if cmdWDTrue && !serverWDTrue {
			return cmdWD
		}

		if !cmdWDTrue && serverWDTrue {
			return serverWD
		}

		// cmdWD relative to serverWD
		if cmdWDTrue && serverWDTrue {
			if filepath.IsAbs(cmdWD) {
				return cmdWD
			}
			return filepath.Join(serverWD, cmdWD)
		}
	}

	// Local

	// cmd context
	if (cmdLocal && serverLocal && !cmdWDTrue && !serverWDTrue) ||
		(cmdLocal && !serverLocal && !cmdWDTrue && !serverWDTrue) ||
		(cmdLocal && !serverLocal && !cmdWDTrue && serverWDTrue) ||
		(!cmdLocal && serverLocal && !cmdWDTrue && !serverWDTrue) {
		return cmdDir
	}

	// cmdWD relative to serverWD and serverDir
	if (cmdLocal && serverLocal && cmdWDTrue && serverWDTrue) ||
		(!cmdLocal && serverLocal && cmdWDTrue && serverWDTrue) {
		if filepath.IsAbs(cmdWD) {
			return cmdWD
		}

		if filepath.IsAbs(serverWD) {
			return filepath.Join(serverWD, cmdWD)
		}

		return filepath.Join(serverDir, serverWD, cmdWD)
	}

	// cmdWD relative to cmd context
	if (cmdLocal && !serverLocal && cmdWDTrue && !serverWDTrue) ||
		(cmdLocal && !serverLocal && cmdWDTrue && serverWDTrue) ||
		(cmdLocal && serverLocal && cmdWDTrue && !serverWDTrue) ||
		(!cmdLocal && serverLocal && cmdWDTrue && !serverWDTrue) {
		if filepath.IsAbs(cmdWD) {
			return cmdWD
		}
		return filepath.Join(cmdDir, cmdWD)
	}

	// serverWD relative to server context
	if (!cmdLocal && serverLocal && !cmdWDTrue && serverWDTrue) ||
		(cmdLocal && serverLocal && !cmdWDTrue && serverWDTrue) {
		if filepath.IsAbs(serverWD) {
			return serverWD
		}
		return filepath.Join(serverDir, serverWD)
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
		signer, err := GetSigner(*server.IdentityFile)
		if err != nil {
			return err
		}
		signers.identities[*server.IdentityFile] = signer
		return nil
	}
}

func getAuthMethod(server dao.Server, signers *Signers) []ssh.AuthMethod {
	var authMethods []ssh.AuthMethod
	var publicKeys []ssh.Signer

	if len(signers.agentSigners) > 0 {
		publicKeys = append(publicKeys, signers.agentSigners...)
	}

	if server.IdentityFile != nil {
		pubKey, found := signers.identities[*server.IdentityFile]
		if found {
			publicKeys = append(publicKeys, pubKey)
		}
	}

	if len(publicKeys) > 0 {
		authMethods = append(authMethods, ssh.PublicKeys(publicKeys...))
	}

	if server.Password != nil {
		pwSigner, found := signers.passwords[*server.Password]
		if found {
			authMethods = append(authMethods, pwSigner)
		}
	}

	return authMethods
}

func CalcForks(batch int, forks uint32) int {
	if batch < int(forks) {
		return batch
	}
	return int(forks)
}

func confirmExecute(taskName string) bool {
	var mu sync.Mutex

	mu.Lock()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nPerform task `%s`: (y)es/(n)o: ", taskName)

	a, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	mu.Unlock()

	return strings.ToLower(strings.TrimSpace(a)) == "yes" || strings.ToLower(strings.TrimSpace(a)) == "y"
}

// TODO: Prompt again when invalid answer
func StepTaskExecute(task string, host string, mu *sync.Mutex) (TaskOption, error) {
	mu.Lock()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Perform task `%s` on host `%s`: (y)es/(n)o/(c)ontinue: ", task, host)

	a, err := reader.ReadString('\n')
	if err != nil {
		return Yes, err
	}

	option := strings.ToLower(strings.TrimSpace(a))
	var value TaskOption

	switch option {
	case "yes", "y":
		value = Yes
	case "no", "n":
		value = No
	case "continue", "c":
		value = Continue
	default:
		value = No
	}

	mu.Unlock()

	return value, nil
}

type TaskOption int

const (
	No = iota
	Yes
	Continue
)
