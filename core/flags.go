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
	Edit      bool
	DryRun    bool
	Describe  bool
	ListHosts bool
	Silent    bool
	Confirm   bool
	Step      bool
	Verbose   bool

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
	Order   string

	// Config
	KnownHostsFile string

	// Task
	Theme  string
	TTY    bool
	Attach bool
	Local  bool

	// Server
	IdentityFile string
	User         string
	Password     string

	// Spec
	Spec              string
	AnyErrorsFatal    bool
	MaxFailPercentage uint8
	IgnoreErrors      bool
	IgnoreUnreachable bool
	OmitEmptyRows     bool
	OmitEmptyColumns  bool
	Forks             uint32
	Batch             uint32
	BatchP            uint8
	Output            string
	Print             string
	Strategy          string
}

type SetRunFlags struct {
	Silent            bool
	Describe          bool
	ListHosts         bool
	Attach            bool
	All               bool
	Invert            bool
	OmitEmptyRows     bool
	OmitEmptyColumns  bool
	Local             bool
	TTY               bool
	AnyErrorsFatal    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
	Order             bool
	Report            bool
	Forks             bool
	Batch             bool
	BatchP            bool
	Servers           bool
	Tags              bool
	Regex             bool
	Limit             bool
	LimitP            bool
	Verbose           bool
	Confirm           bool
	Step              bool
	MaxFailPercentage bool
}
