package pkg


type Chart struct {
	APIVersion   string       `yaml:"apiVersion" json:"apiVersion"`
	Name         string       `yaml:"name" json:"name"`
	Version      string       `yaml:"version" json:"version"`
	Description  string       `yaml:"description" json:"description"`
	Type         string       `yaml:"type" json:"type"`
	Dependencies []Dependency `yaml:"dependencies" json:"dependencies"`
}

type Dependency struct {
	Name       string `yaml:"name" json:"name"`
	Version    string `yaml:"version" json:"version"`
	Repository string `yaml:"repository" json:"repository"`
	Condition  string `yaml:"condition,omitempty" json:"condition,omitempty"`
}

type Values struct {
	Image  ImageConfig  `yaml:"image" json:"image"`
	Canary CanaryConfig `yaml:"canary" json:"canary"`
}

type ImageConfig struct {
	Tag string `yaml:"tag" json:"tag"`
}

type CanaryConfig struct {
	Tag string `yaml:"tag" json:"tag"`
}

type ManifestMetadata struct {
	ImageTag       string   `json:"imageTag"`
	CanaryTag      string   `json:"canaryTag"`
	ContainerImages []string `json:"containerImages"`
	IngressPaths   []string `json:"ingressPaths"`
	ServicePorts   []string `json:"servicePorts"`
}

type ChartInfo struct {
	Chart            Chart             `json:"chart"`
	ImageTag         string            `json:"imageTag"`
	CanaryTag        string            `json:"canaryTag"`
	ManifestMetadata *ManifestMetadata `json:"manifestMetadata,omitempty"`
}

type DockerConfig struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Registry string `json:"Registry"`
}

type ChartRequest struct {
	ChartURL    string   `json:"chartUrl"`
	ValuesPath  string   `json:"valuesPath"`
	SetValues   []string `json:"setValues"`
	UseHostNetwork bool  `json:"useHostNetwork"`
}

type RegistryConfig struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	RegistryURL string `json:"registry_url" db:"registry_url"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"password" db:"password"`
	IsDefault   bool   `json:"is_default" db:"is_default"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	UpdatedAt   string `json:"updated_at" db:"updated_at"`
}


