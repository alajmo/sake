package core

// CMD Flags

type ListFlags struct {
	Output string
	Theme  string
}

type ServerFlags struct {
	Tags       []string
	Headers    []string
	Edit       bool
	Regex      string
	Invert     bool
	AllHeaders bool
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
	Headers    []string
	Edit       bool
	AllHeaders bool
}

type RunFlags struct {
	// Flags
	Edit     bool
	DryRun   bool
	Describe bool
	Silent   bool

	// Reports
	Report []string

	// Target
	All     bool
	Regex   string
	Servers []string
	Tags    []string
	Cwd     bool
	Invert  bool
	Limit   uint32
	LimitP  uint8
	Target  string

	// Config
	KnownHostsFile string

	// Task
	Theme  string
	TTY    bool
	Attach bool
	Local  bool

	// Server
	IdentityFile string
	User		 string
	Password     string

	// Spec
	Spec              string
	AnyErrorsFatal    bool
	MaxFailPercentage uint8
	IgnoreErrors      bool
	IgnoreUnreachable bool
	OmitEmpty         bool
	Forks             uint32
	Batch             uint32
	BatchP            uint8
	Output            string
	Strategy          string
}

type SetRunFlags struct {
	Silent            bool
	Describe          bool
	Attach            bool
	All               bool
	Invert            bool
	OmitEmpty         bool
	Local             bool
	TTY               bool
	AnyErrorsFatal    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
	Report			  bool
}
