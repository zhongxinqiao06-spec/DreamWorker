package extensions

type RunMode string

const (
	RunModeExternal RunMode = "external"
	RunModeManaged  RunMode = "managed"
)

type AppSettings struct {
	EnableNineRouterIntegration bool    `json:"enableNineRouterIntegration"`
	NineRouterRunMode           RunMode `json:"nineRouterRunMode"`
	NineRouterBaseURL           string  `json:"nineRouterBaseURL"`
	NineRouterDashboardURL      string  `json:"nineRouterDashboardURL"`
	NineRouterDefaultModel      string  `json:"nineRouterDefaultModel"`

	NineRouterAutoDetectOnStart     bool   `json:"nineRouterAutoDetectOnStart"`
	NineRouterManagedAutoStart      bool   `json:"nineRouterManagedAutoStart"`
	NineRouterManagedAutoRestart    bool   `json:"nineRouterManagedAutoRestart"`
	NineRouterManagedInstallVersion string `json:"nineRouterManagedInstallVersion"`
	NineRouterManagedPackageName    string `json:"nineRouterManagedPackageName"`
	NineRouterManagedCommand        string `json:"nineRouterManagedCommand"`
	NineRouterManagedWorkDir        string `json:"nineRouterManagedWorkDir"`
	NineRouterManagedLogDir         string `json:"nineRouterManagedLogDir"`
	NineRouterManagedTimeoutMS      int    `json:"nineRouterManagedTimeoutMs"`
	AllowNineRouterAsFreeRoute      bool   `json:"allowNineRouterAsFreeRoute"`
	AllowAgentsUseNineRouter        bool   `json:"allowAgentsUseNineRouter"`
}

type UpdateSettingsInput struct {
	EnableNineRouterIntegration *bool    `json:"enableNineRouterIntegration"`
	NineRouterRunMode           *RunMode `json:"nineRouterRunMode"`
	NineRouterBaseURL           *string  `json:"nineRouterBaseURL"`
	NineRouterDashboardURL      *string  `json:"nineRouterDashboardURL"`
	NineRouterDefaultModel      *string  `json:"nineRouterDefaultModel"`

	NineRouterAutoDetectOnStart     *bool   `json:"nineRouterAutoDetectOnStart"`
	NineRouterManagedAutoStart      *bool   `json:"nineRouterManagedAutoStart"`
	NineRouterManagedAutoRestart    *bool   `json:"nineRouterManagedAutoRestart"`
	NineRouterManagedInstallVersion *string `json:"nineRouterManagedInstallVersion"`
	NineRouterManagedPackageName    *string `json:"nineRouterManagedPackageName"`
	NineRouterManagedCommand        *string `json:"nineRouterManagedCommand"`
	NineRouterManagedWorkDir        *string `json:"nineRouterManagedWorkDir"`
	NineRouterManagedLogDir         *string `json:"nineRouterManagedLogDir"`
	NineRouterManagedTimeoutMS      *int    `json:"nineRouterManagedTimeoutMs"`
	AllowNineRouterAsFreeRoute      *bool   `json:"allowNineRouterAsFreeRoute"`
	AllowAgentsUseNineRouter        *bool   `json:"allowAgentsUseNineRouter"`
}

type ExtensionSpec struct {
	ExtensionID    string                   `json:"extensionId"`
	Name           string                   `json:"name"`
	Kind           string                   `json:"kind"`
	RuntimeKind    string                   `json:"runtimeKind"`
	Description    string                   `json:"description"`
	Install        ExtensionInstallSpec     `json:"install"`
	Process        ExtensionProcessSpec     `json:"process"`
	Health         ExtensionHealthSpec      `json:"health"`
	ProviderBridge *ExtensionProviderBridge `json:"providerBridge"`
	Capabilities   []string                 `json:"capabilities"`
	Security       ExtensionSecuritySpec    `json:"security"`
	SystemPreset   bool                     `json:"systemPreset"`
	Enabled        bool                     `json:"enabled"`
}

type ExtensionInstallSpec struct {
	PackageName    string `json:"packageName"`
	PackageVersion string `json:"packageVersion"`
	RuntimeDir     string `json:"runtimeDir"`
	LogDir         string `json:"logDir"`
	ConfigDir      string `json:"configDir"`
}

type ExtensionProcessSpec struct {
	DefaultCommand string   `json:"defaultCommand"`
	DefaultArgs    []string `json:"defaultArgs"`
	Port           int      `json:"port"`
	Env            []string `json:"env"`
}

type ExtensionHealthSpec struct {
	DashboardURL string `json:"dashboardURL"`
	BaseURL      string `json:"baseURL"`
	ModelsPath   string `json:"modelsPath"`
	ChatPath     string `json:"chatPath"`
}

type ExtensionProviderBridge struct {
	ProviderID    string `json:"providerId"`
	ProviderType  string `json:"providerType"`
	DisplayName   string `json:"displayName"`
	BaseURL       string `json:"baseURL"`
	DefaultModel  string `json:"defaultModel"`
	SortOrder     int    `json:"sortOrder"`
	SystemPreset  bool   `json:"systemPreset"`
	AllowDeletion bool   `json:"allowDeletion"`
}

type ExtensionSecuritySpec struct {
	RiskLevel       string   `json:"riskLevel"`
	AllowedHosts    []string `json:"allowedHosts"`
	SecretKeys      []string `json:"secretKeys"`
	EnvAllowList    []string `json:"envAllowList"`
	ManagedRequires bool     `json:"managedRequiresExplicitEnable"`
}

type ExtensionStatus struct {
	ExtensionID          string          `json:"extensionId"`
	Installed            bool            `json:"installed"`
	InstallSource        string          `json:"installSource"`
	NodeAvailable        bool            `json:"nodeAvailable"`
	NPMAvailable         bool            `json:"npmAvailable"`
	NodeVersion          string          `json:"nodeVersion,omitempty"`
	NPMVersion           string          `json:"npmVersion,omitempty"`
	Command              string          `json:"command,omitempty"`
	RunMode              RunMode         `json:"runMode"`
	ProcessState         string          `json:"processState"`
	PID                  int             `json:"pid,omitempty"`
	StartedByDreamWorker bool            `json:"startedByDreamWorker"`
	BaseURL              string          `json:"baseURL"`
	DashboardURL         string          `json:"dashboardURL"`
	HealthStatus         string          `json:"healthStatus"`
	ModelCount           int             `json:"modelCount"`
	Models               []string        `json:"models"`
	StreamingVerified    bool            `json:"streamingVerified"`
	HasAPIKey            bool            `json:"hasApiKey"`
	MaskedKey            string          `json:"maskedKey,omitempty"`
	LogDir               string          `json:"logDir"`
	WorkDir              string          `json:"workDir"`
	LastStartedAt        string          `json:"lastStartedAt,omitempty"`
	LastStoppedAt        string          `json:"lastStoppedAt,omitempty"`
	LastCheckedAt        string          `json:"lastCheckedAt,omitempty"`
	LastErrorCode        string          `json:"lastErrorCode,omitempty"`
	LastErrorMessage     string          `json:"lastErrorMessage,omitempty"`
	Runtime              NodeRuntimeInfo `json:"runtime"`
}

type NodeRuntimeInfo struct {
	NodeAvailable    bool   `json:"nodeAvailable"`
	NPMAvailable     bool   `json:"npmAvailable"`
	NodeVersion      string `json:"nodeVersion,omitempty"`
	NPMVersion       string `json:"npmVersion,omitempty"`
	CommandAvailable bool   `json:"commandAvailable"`
	Command          string `json:"command,omitempty"`
	InstallSource    string `json:"installSource"`
	ManagedLocalBin  string `json:"managedLocalBin,omitempty"`
	CustomCommand    string `json:"customCommand,omitempty"`
	LastErrorCode    string `json:"lastErrorCode,omitempty"`
	LastErrorMessage string `json:"lastErrorMessage,omitempty"`
}

type ManagedProcess struct {
	ProcessID            string   `json:"processId"`
	ExtensionID          string   `json:"extensionId"`
	PID                  int      `json:"pid"`
	Command              string   `json:"command"`
	Args                 []string `json:"args"`
	WorkDir              string   `json:"workDir"`
	LogFile              string   `json:"logFile"`
	StartedAt            string   `json:"startedAt"`
	StoppedAt            string   `json:"stoppedAt,omitempty"`
	ExitCode             int      `json:"exitCode,omitempty"`
	StartedByDreamWorker bool     `json:"startedByDreamWorker"`
	State                string   `json:"state"`
}

type ExtensionLogLine struct {
	ExtensionID string `json:"extensionId"`
	Timestamp   string `json:"timestamp"`
	Stream      string `json:"stream"`
	Line        string `json:"line"`
}

type InstallExtensionInput struct {
	ExtensionID string `json:"extensionId"`
	Version     string `json:"version"`
}

type ExtensionIDRequest struct {
	ExtensionID string `json:"extensionId"`
}

type TailLogsInput struct {
	ExtensionID string `json:"extensionId"`
	Limit       int    `json:"limit"`
}

type ExtensionActionResult struct {
	OK          bool            `json:"ok"`
	ExtensionID string          `json:"extensionId"`
	Message     string          `json:"message"`
	Status      ExtensionStatus `json:"status"`
}

type ExtensionModelRefreshResult struct {
	OK          bool            `json:"ok"`
	ExtensionID string          `json:"extensionId"`
	Models      []string        `json:"models"`
	Status      ExtensionStatus `json:"status"`
}

type ExtensionStreamingResult struct {
	OK          bool            `json:"ok"`
	ExtensionID string          `json:"extensionId"`
	Message     string          `json:"message"`
	LatencyMS   int             `json:"latencyMs"`
	Status      ExtensionStatus `json:"status"`
}

type Error struct {
	Code       string
	Message    string
	UserAction string
}

func (e *Error) Error() string {
	return e.Message
}
