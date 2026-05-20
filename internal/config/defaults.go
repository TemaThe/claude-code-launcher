package config

func DefaultConfigYAML() string {
	return `defaults:
  claude_binary: claude
  temp_root: /tmp/ccx
  redact_patterns:
    - key
    - token
    - secret
    - password
plugins:
  serena@claude-plugins-official:
    default: inherit
  gopls-lsp@claude-plugins-official:
    default: inherit
hooks:
  none:
    hooks: {}
mcp:
  none:
    servers: {}
  serena:
    servers:
      serena:
        command: serena
        args:
          - serve
  gitlab:
    servers: {}
  grafana-dashboard:
    servers: {}
  k8s-readonly:
    servers: {}
profiles:
  min:
    hook_profile: none
    mcp_profile: none
    plugins:
      serena@claude-plugins-official: false
    flags:
      strict_mcp_config: true
      exclude_dynamic_system_prompt_sections: true
  serena:
    hook_profile: none
    mcp_profile: serena
    plugins:
      serena@claude-plugins-official: false
      gopls-lsp@claude-plugins-official: true
    flags:
      strict_mcp_config: true
      exclude_dynamic_system_prompt_sections: true
`
}
