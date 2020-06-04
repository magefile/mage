package mg

var shutdownHooks []func()

// AddShutdownHook adds a hook to mage's shutdown logic.
// Mage will execute the hooks when shutting down
// allowing for any clean up.
func AddShutdownHook(f func()) {
	shutdownHooks = append(shutdownHooks, f)
}

// RunShutdownHooks is called by mage to execute any registered shutdown hooks.
func RunShutdownHooks() {
	for _, f := range shutdownHooks {
		f()
	}
	// reset to prevent the hooks from running on subsequent invocations
	shutdownHooks = nil
}
