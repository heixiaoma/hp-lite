package config

type Config struct {
	Admin  AdminConfig  `yaml:"admin"`
	Cmd    CmdConfig    `yaml:"cmd"`
	Tunnel TunnelConfig `yaml:"tunnel"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

type CmdConfig struct {
	Port int `yaml:"port"`
}

type TunnelConfig struct {
	IP         string `yaml:"ip"`
	Port       int    `yaml:"port"`
	OpenDomain bool   `yaml:"open-domain"`
}
