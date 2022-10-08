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
	Regex   string
	Invert  bool
}

type TargetFlags struct {
	Headers []string
	Edit    bool
}

type SpecFlags struct {
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
	Silent   bool

	// Target
	All     bool
	Regex   string
	Servers []string
	Tags    []string
	Cwd     bool
	Invert  bool
	Limit   uint32
	LimitP  uint8

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
	All               bool
	Invert            bool
	Parallel          bool
	OmitEmpty         bool
	Local             bool
	TTY               bool
	AnyErrorsFatal    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
}
