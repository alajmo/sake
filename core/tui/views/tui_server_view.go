package views

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
)

type TServer struct {
	// UI
	Page            *tview.Flex
	ServerTableView *components.TTable
	ServerTreeView  *components.TTree
	TagView         *components.TList

	// Server
	Servers           []dao.Server
	ServersFiltered   []dao.Server
	ServersSelected   map[string]bool
	serverFilterValue *string
	Headers           []string
	ShowHeaders       bool
	ServerStyle       string

	// Tags
	ServerTags           []string
	ServerTagsFiltered   []string
	ServerTagsSelected   map[string]bool
	serverTagFilterValue *string

	// Misc
	Emitter *misc.EventEmitter
}

func CreateServersData(
	servers []dao.Server,
	serverTags []string,
	headers []string,
	prefixNumber int,
	showTitle bool,
	showHeaders bool,
	selectEnabled bool,
	showTags bool,
) *TServer {
	s := &TServer{
		Servers:           servers,
		ServersFiltered:   servers,
		ServersSelected:   make(map[string]bool),
		serverFilterValue: new(string),

		ServerTags:           serverTags,
		ServerTagsFiltered:   serverTags,
		ServerTagsSelected:   make(map[string]bool),
		serverTagFilterValue: new(string),

		ShowHeaders: showHeaders,
		Headers:     headers,
		ServerStyle: "server-table",

		Emitter: misc.NewEventEmitter(),
	}

	for _, server := range s.Servers {
		s.ServersSelected[server.Name] = false
	}
	for _, tag := range s.ServerTags {
		s.ServerTagsSelected[tag] = false
	}

	title := ""
	if showTitle && prefixNumber > 0 {
		title = fmt.Sprintf("[%d] Servers (%d)", prefixNumber, len(servers))
		prefixNumber += 1
	} else if showTitle {
		title = fmt.Sprintf("Servers (%d)", len(servers))
	}

	rows := s.getTableRows()
	serverTable := s.CreateServersTable(selectEnabled, title, headers, rows)
	s.ServerTableView = serverTable

	nodes := s.getServerTreeHierarchy()
	serverTree := s.CreateServersTree(selectEnabled, title, nodes)
	s.ServerTreeView = serverTree

	if showTags {
		tagTitle := ""
		if showTitle && prefixNumber > 0 {
			tagTitle = fmt.Sprintf("[%d] Tags (%d)", prefixNumber, len(serverTags))
		} else {
			tagTitle = fmt.Sprintf("Tags (%d)", len(serverTags))
		}

		tagsList := s.CreateServersTagsList(tagTitle)
		s.TagView = tagsList
	}

	// Events
	s.Emitter.Subscribe("remove_tag_filter", func(e misc.Event) {
		if s.TagView != nil {
			s.TagView.ClearFilter()
		}
	})
	s.Emitter.Subscribe("remove_tag_selections", func(e misc.Event) {
		s.unselectAllTags()
	})
	s.Emitter.Subscribe("remove_server_filter", func(e misc.Event) {
		s.ServerTableView.ClearFilter()
		s.ServerTreeView.ClearFilter()
	})
	s.Emitter.Subscribe("remove_server_selections", func(event misc.Event) {
		s.unselectAllServers()
	})
	s.Emitter.Subscribe("filter_servers", func(e misc.Event) {
		s.filterServers()
	})

	return s
}

func (s *TServer) CreateServersTable(
	selectEnabled bool,
	title string,
	headers []string,
	rows [][]string,
) *components.TTable {
	table := &components.TTable{
		Title:         title,
		ToggleEnabled: selectEnabled,
		ShowHeaders:   s.ShowHeaders,
		FilterValue:   s.serverFilterValue,
	}
	table.Create()
	table.Update(headers, rows)

	// Methods
	table.IsRowSelected = func(name string) bool {
		return s.ServersSelected[name]
	}
	table.ToggleSelectRow = func(name string) {
		s.toggleSelectServer(name)
	}
	table.SelectAll = func() {
		s.selectAllServers()
	}
	table.UnselectAll = func() {
		s.unselectAllServers()
	}
	table.FilterRows = func() {
		s.filterServers()
	}
	table.DescribeRow = func(serverName string) {
		if serverName != "" {
			s.showServerDescModal(serverName)
		}
	}
	table.EditRow = func(serverName string) {
		if serverName != "" {
			s.editServer(serverName)
		}
	}
	table.SSHRow = func(serverName string) {
		if serverName != "" {
			s.sshServer(serverName)
		}
	}
	return table
}

func (s *TServer) CreateServersTagsList(title string) *components.TList {
	list := &components.TList{
		Title:       title,
		FilterValue: s.serverTagFilterValue,
	}
	list.Create()
	list.Update(s.ServerTags)

	// Methods
	list.IsItemSelected = func(name string) bool {
		return s.ServerTagsSelected[name]
	}
	list.ToggleSelectItem = func(i int, tag string) {
		s.ServerTagsSelected[tag] = !s.ServerTagsSelected[tag]
		list.SetItemSelect(i, tag)
		s.filterServers()
	}
	list.SelectAll = func() {
		s.selectAllTags()
		s.filterServers()
	}
	list.UnselectAll = func() {
		s.unselectAllTags()
		s.filterServers()
	}
	list.FilterItems = func() {
		s.filterTags()
	}

	return list
}

func (s *TServer) CreateServersTree(
	selectEnabled bool,
	title string,
	nodes []components.TNode,
) *components.TTree {
	tree := &components.TTree{
		Title:         title,
		RootTitle:     "",
		SelectEnabled: selectEnabled,
		FilterValue:   s.serverFilterValue,
	}
	tree.Create()
	tree.UpdateServers(nodes)
	tree.UpdateServersStyle()

	tree.IsNodeSelected = func(name string) bool {
		return s.ServersSelected[name]
	}
	tree.ToggleSelectNode = func(name string) {
		s.toggleSelectServer(name)
	}
	tree.SelectAll = func() {
		s.selectAllServers()
	}
	tree.UnselectAll = func() {
		s.unselectAllServers()
	}
	tree.FilterNodes = func() {
		s.filterServers()
	}
	tree.DescribeNode = func(serverName string) {
		if serverName != "" {
			s.showServerDescModal(serverName)
		}
	}
	tree.EditNode = func(serverName string) {
		if serverName != "" {
			s.editServer(serverName)
		}
	}

	return tree
}

func (s *TServer) getTableRows() [][]string {
	var rows = make([][]string, len(s.ServersFiltered))
	for i, server := range s.ServersFiltered {
		rows[i] = make([]string, len(s.Headers))
		for j, header := range s.Headers {
			rows[i][j] = server.GetValue(header, 0)
		}
	}
	return rows
}

// getServerTreeHierarchy groups servers by common IP prefix or hostname domain
func (s *TServer) getServerTreeHierarchy() []components.TNode {
	// Group servers by common prefix
	groups := s.groupServersByCommonPrefix()

	var nodes []components.TNode
	// Track seen servers to prevent duplicates
	seen := make(map[string]bool)

	// Iterate groups in sorted key order; ranging the map directly reshuffles
	// the tree on every filter/redraw. "" (flat servers) sorts first.
	groupNames := make([]string, 0, len(groups))
	for groupName := range groups {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		servers := groups[groupName]
		if groupName == "" {
			// Flat servers (no grouping)
			for _, server := range servers {
				if seen[server.Name] {
					continue
				}
				seen[server.Name] = true

				displayName := server.Name
				if server.Host != "" && server.Host != server.Name {
					displayName = server.Host + " - " + server.Name
				}

				node := components.TNode{
					DisplayName: displayName,
					ID:          server.Name,
					Type:        "server",
					Children:    &[]components.TNode{},
				}
				nodes = append(nodes, node)
			}
		} else {
			// Grouped servers
			parentNode := components.TNode{
				DisplayName: groupName,
				ID:          "",
				Type:        "group",
				Children:    &[]components.TNode{},
			}
			for _, server := range servers {
				if seen[server.Name] {
					continue
				}
				seen[server.Name] = true

				// Show "host - serverName" format
				displayName := server.Host + " - " + server.Name
				child := components.TNode{
					DisplayName: displayName,
					ID:          server.Name,
					Type:        "server",
				}
				*parentNode.Children = append(*parentNode.Children, child)
			}
			nodes = append(nodes, parentNode)
		}
	}

	return nodes
}

// groupServersByCommonPrefix groups servers by IP prefix or hostname domain
func (s *TServer) groupServersByCommonPrefix() map[string][]dao.Server {
	if len(s.ServersFiltered) == 0 {
		return map[string][]dao.Server{}
	}

	// Try to find common IP prefixes (e.g., 192.168.1.x)
	ipGroups := make(map[string][]dao.Server)
	hostnameGroups := make(map[string][]dao.Server)
	ungrouped := []dao.Server{}

	// Track seen servers to prevent duplicates
	seen := make(map[string]bool)

	for _, server := range s.ServersFiltered {
		// Skip duplicates
		if seen[server.Name] {
			continue
		}
		seen[server.Name] = true

		host := server.Host
		// Local servers or servers without host go to ungrouped
		if host == "" || server.Local {
			ungrouped = append(ungrouped, server)
			continue
		}

		// Check if it's an IP address
		if isIPAddress(host) {
			prefix := getIPPrefix(host)
			if prefix != "" {
				ipGroups[prefix] = append(ipGroups[prefix], server)
			} else {
				ungrouped = append(ungrouped, server)
			}
		} else {
			// It's a hostname - group by domain
			domain := getHostnameDomain(host)
			if domain != "" {
				hostnameGroups[domain] = append(hostnameGroups[domain], server)
			} else {
				ungrouped = append(ungrouped, server)
			}
		}
	}

	// Decide which grouping to use based on which has more groups with multiple servers
	result := make(map[string][]dao.Server)

	// Count meaningful groups (groups with more than 1 server)
	ipMeaningful := 0
	for _, servers := range ipGroups {
		if len(servers) > 1 {
			ipMeaningful++
		}
	}
	hostnameMeaningful := 0
	for _, servers := range hostnameGroups {
		if len(servers) > 1 {
			hostnameMeaningful++
		}
	}

	// Use the grouping that has more meaningful groups. Iterate the group maps
	// in sorted key order so demoted single-server groups land in `ungrouped`
	// deterministically (ranging the map directly reshuffles the flat section).
	if ipMeaningful > 0 && ipMeaningful >= hostnameMeaningful {
		// Use IP grouping
		prefixes := make([]string, 0, len(ipGroups))
		for prefix := range ipGroups {
			prefixes = append(prefixes, prefix)
		}
		sort.Strings(prefixes)
		for _, prefix := range prefixes {
			servers := ipGroups[prefix]
			if len(servers) > 1 {
				result[prefix+".*"] = servers
			} else {
				ungrouped = append(ungrouped, servers...)
			}
		}
	} else if hostnameMeaningful > 0 {
		// Use hostname grouping
		domains := make([]string, 0, len(hostnameGroups))
		for domain := range hostnameGroups {
			domains = append(domains, domain)
		}
		sort.Strings(domains)
		for _, domain := range domains {
			servers := hostnameGroups[domain]
			if len(servers) > 1 {
				result["*."+domain] = servers
			} else {
				ungrouped = append(ungrouped, servers...)
			}
		}
	}

	// Add ungrouped servers as flat
	if len(ungrouped) > 0 {
		result[""] = ungrouped
	}

	// If no meaningful grouping found, return all as flat
	if len(result) == 0 || (len(result) == 1 && result[""] != nil) {
		return map[string][]dao.Server{"": s.ServersFiltered}
	}

	return result
}

// isIPAddress checks if a string looks like an IP address
func isIPAddress(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if _, err := fmt.Sscanf(part, "%d", new(int)); err != nil {
			return false
		}
	}
	return true
}

// getIPPrefix returns the first 3 octets of an IP address
func getIPPrefix(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], ".")
	}
	return ""
}

// getHostnameDomain returns the domain part of a hostname (last 2 parts)
func getHostnameDomain(hostname string) string {
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return ""
}

func (s *TServer) toggleSelectServer(name string) {
	s.ServersSelected[name] = !s.ServersSelected[name]
	s.ServerTableView.ToggleSelectCurrentRow(name)
	s.ServerTreeView.ToggleSelectCurrentNode(name)
}

func (s *TServer) filterServers() {
	serverTags := []string{}
	for key, filtered := range s.ServerTagsSelected {
		if filtered {
			serverTags = append(serverTags, key)
		}
	}

	if len(serverTags) > 0 {
		var filtered []dao.Server
		for _, server := range s.Servers {
			// Match on the first selected tag the server carries, then move to
			// the next server. A plain break would only exit the inner loop, so
			// a server with multiple selected tags would be appended once per
			// match -> duplicate rows and duplicated execution targets.
		matchTags:
			for _, tag := range serverTags {
				for _, serverTag := range server.Tags {
					if serverTag == tag {
						filtered = append(filtered, server)
						break matchTags
					}
				}
			}
		}
		s.ServersFiltered = filtered
	} else {
		s.ServersFiltered = s.Servers
	}

	var finalServers []dao.Server
	for _, server := range s.ServersFiltered {
		if strings.Contains(strings.ToLower(server.Name), strings.ToLower(*s.serverFilterValue)) {
			finalServers = append(finalServers, server)
		}
	}
	s.ServersFiltered = finalServers

	// Table
	rows := s.getTableRows()
	s.ServerTableView.Update(s.Headers, rows)
	s.ServerTableView.Table.ScrollToBeginning()
	s.ServerTableView.Table.Select(1, 0)

	// Tree
	serverTree := s.getServerTreeHierarchy()
	s.ServerTreeView.UpdateServers(serverTree)
	s.ServerTreeView.UpdateServersStyle()
	s.ServerTreeView.FocusFirst()
}

func (s *TServer) filterTags() {
	var finalTags []string
	for _, tag := range s.ServerTags {
		if strings.Contains(tag, *s.serverTagFilterValue) {
			finalTags = append(finalTags, tag)
		}
	}
	s.ServerTagsFiltered = finalTags
	s.TagView.Update(s.ServerTagsFiltered)
}

func (s *TServer) selectAllServers() {
	for _, server := range s.ServersFiltered {
		s.ServersSelected[server.Name] = true
	}
	s.ServerTableView.UpdateRowStyle()
	s.ServerTreeView.UpdateServersStyle()
}

func (s *TServer) selectAllTags() {
	for _, tag := range s.ServerTagsFiltered {
		s.ServerTagsSelected[tag] = true
	}
	s.TagView.Update(s.ServerTagsFiltered)
}

func (s *TServer) unselectAllServers() {
	for _, server := range s.ServersFiltered {
		s.ServersSelected[server.Name] = false
	}
	s.ServerTableView.UpdateRowStyle()
	s.ServerTreeView.UpdateServersStyle()
}

func (s *TServer) unselectAllTags() {
	for _, tag := range s.ServerTagsFiltered {
		s.ServerTagsSelected[tag] = false
	}
	if s.TagView != nil {
		s.TagView.Update(s.ServerTagsFiltered)
	}
}

func (s *TServer) showServerDescModal(name string) {
	server, err := misc.Config.GetServer(name)
	if err != nil {
		return
	}

	// Get bastion hosts as strings
	var bastions []string
	for _, bastion := range server.Bastions {
		bastionStr := bastion.Host
		if bastion.User != "" {
			bastionStr = bastion.User + "@" + bastionStr
		}
		bastions = append(bastions, bastionStr)
	}

	// Get identity file
	identityFile := ""
	if server.IdentityFile != nil {
		identityFile = *server.IdentityFile
	}

	description := misc.FormatServerBlock(
		server.Name,
		server.Desc,
		server.Host,
		server.User,
		server.Port,
		server.Local,
		server.Tags,
		bastions,
		identityFile,
		server.WorkDir,
	)
	components.OpenTextModal("server-description-modal", description, server.Name)
}

func (s *TServer) editServer(serverName string) {
	misc.App.Suspend(func() {
		err := misc.Config.EditServer(serverName)
		if err != nil {
			return
		}
	})
}

func (s *TServer) sshServer(serverName string) {
	server, err := misc.Config.GetServer(serverName)
	if err != nil {
		return
	}

	// Don't SSH to local servers
	if server.Local {
		return
	}

	misc.App.Suspend(func() {
		// Build SSH command args
		args := []string{}

		// Add identity file if specified
		if server.IdentityFile != nil && *server.IdentityFile != "" {
			args = append(args, "-i", *server.IdentityFile)
		}

		// Add port if not default
		if server.Port != 0 && server.Port != 22 {
			args = append(args, "-p", fmt.Sprintf("%d", server.Port))
		}

		// Add bastion/jump hosts if specified
		if len(server.Bastions) > 0 {
			var jumpHosts []string
			for _, bastion := range server.Bastions {
				jumpHost := bastion.Host
				if bastion.User != "" {
					jumpHost = bastion.User + "@" + jumpHost
				}
				if bastion.Port != 0 && bastion.Port != 22 {
					jumpHost = fmt.Sprintf("%s:%d", jumpHost, bastion.Port)
				}
				jumpHosts = append(jumpHosts, jumpHost)
			}
			args = append(args, "-J", strings.Join(jumpHosts, ","))
		}

		// Add user@host
		target := server.Host
		if server.User != "" {
			target = server.User + "@" + server.Host
		}
		args = append(args, target)

		// Run SSH command
		cmd := exec.Command("ssh", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	})
}

// GetSelectedServers returns names of selected servers
func (s *TServer) GetSelectedServers() []string {
	var selected []string
	for name, isSelected := range s.ServersSelected {
		if isSelected {
			selected = append(selected, name)
		}
	}
	return selected
}

// GetSelectedServerObjects returns the selected server objects
func (s *TServer) GetSelectedServerObjects() []dao.Server {
	var selected []dao.Server
	for _, server := range s.Servers {
		if s.ServersSelected[server.Name] {
			selected = append(selected, server)
		}
	}
	return selected
}

// GetFilteredServers returns the filtered server objects
func (s *TServer) GetFilteredServers() []dao.Server {
	return s.ServersFiltered
}
