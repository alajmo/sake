package core

// CMD Flags

type ListFlags struct {
	Output string
	Theme  string
}

type ServerFlags struct {
	Tags    []string
	Headers []string
	Edit    bool
}

type TagFlags struct {
	Headers []string
}

type TaskFlags struct {
	Headers []string
	Edit    bool
}

type RunFlags struct {
	// Flags
	Edit     bool
	DryRun   bool
	Describe bool
	Debug    bool

	// Target
	All     bool
	Servers []string
	Tags    []string
	Cwd     bool

	// Config
	KnownHostsFile string

	// Task
	Theme  string
	TTY    bool
	Attach bool
	Local  bool

	// Server
	IdentityFile string
	Password     string

	// Spec
	Parallel          bool
	AnyErrorsFatal    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
	OmitEmpty         bool
	Output            string
}

type SetRunFlags struct {
	Parallel          bool
	OmitEmpty         bool
	Local             bool
	AnyErrorsFatal    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
}
