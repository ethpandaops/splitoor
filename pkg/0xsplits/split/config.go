package split

type Config struct {
	ContractAddress string  `yaml:"contractAddress,omitempty"`
	SplitAddress    *string `yaml:"splitAddress"`
}

func (c *Config) Validate() error {
	return nil
}
