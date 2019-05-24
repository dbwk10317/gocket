package gocket

type (
	SessionConfig struct {
		MaxChannelBufferSize int32
		MaxReadBufferSize    int32
		MaxWriteBufferSize   int32
	}

	PipelineConfig struct {
		MaxChannelBufferSize int32
	}
)

func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		MaxChannelBufferSize: 32,
		MaxReadBufferSize:    32 * 1024,
		MaxWriteBufferSize:   32 * 1024,
	}
}

func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		MaxChannelBufferSize: 32,
	}
}
