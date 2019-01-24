package qiniu

type Option func(s *S3)

func AccessKey(key string) Option {
	return func(s *S3) {
		s.accessKey = key
	}
}

func SecretKey(key string) Option {
	return func(s *S3) {
		s.secretKey = key
	}
}

type Config struct {
	useHTTPS bool
}

func NewConfig(opts ...RequestOption) *Config {
	c := &Config{}
	for _, o := range opts {
		o(c)
	}
	return c
}

type RequestOption func(c *Config)

func UseHTTPS(v bool) RequestOption {
	return func(c *Config) {
		c.useHTTPS = v
	}
}