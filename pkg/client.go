package cadence

var (
	register Register
)

func init() {
	register = NewRegister()
}

func Register(w interface{}, a interface{}, options RegisterNamespaceOptions) {
	register.RegisterWorkflow(w)
	register.RegisterActivity(a)
	register.RegisterWorker()
}

func RegisterNamespace(namespace string, options RegisterNamespaceOptions) error {
	return register.RegisterNamespace(namespace, options)
}

func ExecuteWorkflow() error {
	return nil
}
