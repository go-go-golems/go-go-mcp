package config

import (
	"fmt"

	yaml_editor "github.com/go-go-golems/clay/pkg/yaml-editor"
	"gopkg.in/yaml.v3"
)

type ConfigEditor struct {
	editor *yaml_editor.YAMLEditor
	path   string
}

func NewConfigEditor(path string) (*ConfigEditor, error) {
	editor, err := yaml_editor.NewYAMLEditorFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not create editor: %w", err)
	}

	return &ConfigEditor{
		editor: editor,
		path:   path,
	}, nil
}

func (c *ConfigEditor) Save() error {
	return c.editor.Save(c.path)
}

func (c *ConfigEditor) AddProfile(name, description string) error {
	// Create a new profile node
	profileNode, err := c.editor.CreateMap(
		"description", description,
		"tools", &yaml.Node{Kind: yaml.MappingNode},
		"prompts", &yaml.Node{Kind: yaml.MappingNode},
	)
	if err != nil {
		return fmt.Errorf("could not create profile node: %w", err)
	}

	// Add the profile to the profiles map
	err = c.editor.SetNode(profileNode, "profiles", name)
	if err != nil {
		return fmt.Errorf("could not add profile: %w", err)
	}

	return nil
}

func (c *ConfigEditor) DuplicateProfile(source, target, description string) error {
	// Get the source profile
	sourceNode, err := c.editor.GetNode("profiles", source)
	if err != nil {
		return fmt.Errorf("could not get source profile: %w", err)
	}

	// Create a deep copy of the source profile
	targetNode := c.editor.DeepCopyNode(sourceNode)

	// Update the description
	descNode, err := c.editor.GetMapNode("description", targetNode)
	if err != nil {
		return fmt.Errorf("could not get description node: %w", err)
	}
	descNode.Value = description

	// Add the new profile
	err = c.editor.SetNode(targetNode, "profiles", target)
	if err != nil {
		return fmt.Errorf("could not add target profile: %w", err)
	}

	return nil
}

func (c *ConfigEditor) AddToolDirectory(profile, path string, defaults map[string]interface{}) error {
	// Create the directory source node
	dirNode, err := c.editor.CreateMap(
		"path", path,
		"defaults", map[string]interface{}{
			"default": defaults,
		},
	)
	if err != nil {
		return fmt.Errorf("could not create directory node: %w", err)
	}

	// Get the tools node
	toolsNode, err := c.editor.GetNode("profiles", profile, "tools")
	if err != nil {
		return fmt.Errorf("could not get tools node: %w", err)
	}

	// Get or create the directories sequence
	var dirSeqNode *yaml.Node
	_, err = c.editor.GetMapNode("directories", toolsNode)
	if err != nil {
		// Create new directories sequence
		dirSeqNode = &yaml.Node{Kind: yaml.SequenceNode}
		err = c.editor.SetNode(dirSeqNode, "profiles", profile, "tools", "directories")
		if err != nil {
			return fmt.Errorf("could not create directories sequence: %w", err)
		}
	}

	// Append the new directory
	err = c.editor.AppendToSequence(dirNode, "profiles", profile, "tools", "directories")
	if err != nil {
		return fmt.Errorf("could not append directory: %w", err)
	}

	return nil
}

func (c *ConfigEditor) AddToolFile(profile, path string) error {
	// Create the file source node
	fileNode, err := c.editor.CreateMap(
		"path", path,
	)
	if err != nil {
		return fmt.Errorf("could not create file node: %w", err)
	}

	// Get the tools node
	toolsNode, err := c.editor.GetNode("profiles", profile, "tools")
	if err != nil {
		return fmt.Errorf("could not get tools node: %w", err)
	}

	// Get or create the files sequence
	var fileSeqNode *yaml.Node
	_, err = c.editor.GetMapNode("files", toolsNode)
	if err != nil {
		// Create new files sequence
		fileSeqNode = &yaml.Node{Kind: yaml.SequenceNode}
		err = c.editor.SetNode(fileSeqNode, "profiles", profile, "tools", "files")
		if err != nil {
			return fmt.Errorf("could not create files sequence: %w", err)
		}
	}

	// Append the new file
	err = c.editor.AppendToSequence(fileNode, "profiles", profile, "tools", "files")
	if err != nil {
		return fmt.Errorf("could not append file: %w", err)
	}

	return nil
}

func (c *ConfigEditor) SetDefaultProfile(profile string) error {
	// Create a scalar node with the profile name
	profileNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: profile,
	}

	// Set it as the default profile
	err := c.editor.SetNode(profileNode, "defaultProfile")
	if err != nil {
		return fmt.Errorf("could not set default profile: %w", err)
	}

	return nil
}

func (c *ConfigEditor) GetProfiles() (map[string]string, error) {
	// Get the profiles node
	profilesNode, err := c.editor.GetNode("profiles")
	if err != nil {
		return nil, fmt.Errorf("could not get profiles node: %w", err)
	}

	profiles := make(map[string]string)
	for i := 0; i < len(profilesNode.Content); i += 2 {
		name := profilesNode.Content[i].Value
		profile := profilesNode.Content[i+1]

		// Get description
		descNode, err := c.editor.GetMapNode("description", profile)
		if err != nil {
			return nil, fmt.Errorf("could not get description for profile %s: %w", name, err)
		}

		profiles[name] = descNode.Value
	}

	return profiles, nil
}

func (c *ConfigEditor) GetProfile(name string) (*yaml.Node, error) {
	return c.editor.GetNode("profiles", name)
}

func (c *ConfigEditor) GetDefaultProfile() (string, error) {
	node, err := c.editor.GetNode("defaultProfile")
	if err != nil {
		return "", fmt.Errorf("could not get default profile: %w", err)
	}

	return node.Value, nil
}
