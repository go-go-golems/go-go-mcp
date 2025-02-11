package config

// Config represents the root configuration
type Config struct {
	Version        string              `yaml:"version"`
	DefaultProfile string              `yaml:"defaultProfile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
}

// Profile represents a named configuration profile
type Profile struct {
	Description string         `yaml:"description"`
	Tools       *ToolSources   `yaml:"tools"`
	Prompts     *PromptSources `yaml:"prompts"`
}

// Common source configuration for both tools and prompts
type SourceConfig struct {
	Path      string          `yaml:"path"`
	Defaults  LayerParameters `yaml:"defaults,omitempty"`
	Overrides LayerParameters `yaml:"overrides,omitempty"`
	Blacklist ParameterFilter `yaml:"blacklist,omitempty"`
	Whitelist ParameterFilter `yaml:"whitelist,omitempty"`
}

// LayerParameters maps layer names to their parameter settings
type LayerParameters map[string]map[string]interface{}

// ParameterFilter maps layer names to lists of parameter names
type ParameterFilter map[string][]string

// ToolSources configures where tools are loaded from
type ToolSources struct {
	Directories      []SourceConfig `yaml:"directories,omitempty"`
	Files            []SourceConfig `yaml:"files,omitempty"`
	ExternalCommands []struct {
		Command      string   `yaml:"command"`
		Args         []string `yaml:"args,omitempty"`
		SourceConfig `yaml:",inline"`
	} `yaml:"external_commands,omitempty"`
}

// PromptSources configures where prompts are loaded from
type PromptSources struct {
	Directories []SourceConfig `yaml:"directories,omitempty"`
	Files       []SourceConfig `yaml:"files,omitempty"`
	Pinocchio   *struct {
		Command      string   `yaml:"command"`
		Args         []string `yaml:"args,omitempty"`
		SourceConfig `yaml:",inline"`
	} `yaml:"pinocchio,omitempty"`
}
