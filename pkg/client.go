package cadence

type (
	Client interface {
		Register(interface{}, interface{})
		RegisterNamespace(string, RegisterNamespaceOptions) error
		ExecuteWorkflow() error
	}
	clientImpl struct {
		register Register
		temporal *Temporal
	}
)

func NewClient(c ClientOptions, w WorkerOptions, queue string) (Client, error) {
	temporal, err := NewTemporalClient(c, w, queue)
	if err != nil {
		return nil, err
	}
	return &clientImpl{
		register: NewRegister(temporal),
		temporal: temporal,
	}, nil
}

func (c *clientImpl) Register(w interface{}, a interface{}) {
	c.register.RegisterWorkflow(w)
	c.register.RegisterActivity(a)
	c.register.RegisterWorker()
}

func (c *clientImpl) RegisterNamespace(namespace string, options RegisterNamespaceOptions) error {
	return c.register.RegisterNamespace(namespace, options)
}

func (c *clientImpl) ExecuteWorkflow() error {
	return nil
}
