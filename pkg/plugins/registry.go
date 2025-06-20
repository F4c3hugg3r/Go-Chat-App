package plugins

type Plugin interface {
	Name() string
	Execute() string
}

type PluginRegistry struct {
	plugins map[string]*Plugin
}

func (p *PluginRegistry) RunPlugin(name string) (result string, err error) {
	return
}
