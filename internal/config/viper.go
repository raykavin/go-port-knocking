package config

import (
	"PROJECT_NAME/pkg/errs"
	"context"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const cfgIfTag = "Infra: Carregador de configurações"

var (
	ErrConfigFileReadFailed = errs.New(errs.ErrorTypeProcessing,
		"ERR_CONFIG_FILE_READ_FAILED", "Falha ao ler arquivo de configuração", cfgIfTag)
	ErrConfigUnmarshalFailed = errs.New(errs.ErrorTypeConversion,
		"ERR_CONFIG_UNMARSHAL_FAILED", "Falha ao desserializar configuração", cfgIfTag)
	ErrConfigValidationFailed = errs.New(errs.ErrorTypeValidation,
		"ERR_CONFIG_VALIDATION_FAILED", "Falha na validação da configuração", cfgIfTag)
	ErrConfigReloadFailed = errs.New(errs.ErrorTypeProcessing,
		"ERR_CONFIG_RELOAD_FAILED", "Falha ao recarregar configuração", cfgIfTag)
	ErrValidationReloadFailed = errs.New(errs.ErrorTypeValidation,
		"ERR_VALIDATION_RELOAD_FAILED", "Falha na validação durante recarga da configuração", cfgIfTag)
)

// LoaderOptions contains configuration options for the loader
type LoaderOptions[T any] struct {
	ConfigName          string        // Configuration file name (no extension)
	ConfigType          string        // Configuration file type (yaml, json, toml, etc.)
	ConfigPaths         []string      // Paths to search for the configuration file
	WatchConfig         bool          // Enable configuration file observation
	ReloadDebounce      time.Duration // Debounce duration for config reloads
	OnConfigChange      func(*T)      // Callback when config changes (optional)
	OnConfigChangeError func(error)   // Callback when config reload fails (optional)
}

// ConfigChangeEvent represents a configuration change event
type ConfigChangeEvent[T any] struct {
	OldConfig *T
	NewConfig *T
	Error     error
	Timestamp time.Time
}

// ConfigWatcher defines the interface for configuration change notifications
type ConfigWatcher[T any] interface {
	Subscribe() <-chan ConfigChangeEvent[T]
	Unsubscribe(<-chan ConfigChangeEvent[T])
}

// DefaultLoaderOptions returns sensible defaults for configuration loading
func DefaultLoaderOptions[T any]() *LoaderOptions[T] {
	return &LoaderOptions[T]{
		ConfigName:          "config",
		ConfigType:          "yaml",
		ConfigPaths:         []string{".", "./config", "/etc/app", "$HOME/.app"},
		WatchConfig:         false,
		ReloadDebounce:      1 * time.Second,
		OnConfigChange:      nil,
		OnConfigChangeError: nil,
	}
}

// Loader provides methods for loading configuration
type Loader[T any] struct {
	options     *LoaderOptions[T]
	viper       *viper.Viper
	current     *T
	mutex       sync.RWMutex
	subscribers []chan ConfigChangeEvent[T]
	subMutex    sync.RWMutex
	validator   func(*T) error
	ctx         context.Context
	cancel      context.CancelFunc
	debouncer   *time.Timer
}

// NewViper creates a new configuration loader with the given options
func NewViper[T any](options *LoaderOptions[T]) *Loader[T] {
	if options == nil {
		options = DefaultLoaderOptions[T]()
	}

	v := viper.New()

	// Set config file properties
	v.SetConfigName(options.ConfigName)
	v.SetConfigType(options.ConfigType)

	// Add config paths
	for _, path := range options.ConfigPaths {
		v.AddConfigPath(path)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cl := &Loader[T]{
		options:     options,
		viper:       v,
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make([]chan ConfigChangeEvent[T], 0),
	}

	return cl
}

// Load reads configuration from file
func (cl *Loader[T]) Load() (*T, error) {
	config, err := cl.loadConfig()
	if err != nil {
		return nil, err
	}

	cl.mutex.Lock()
	cl.current = config
	cl.mutex.Unlock()

	// Start watching for changes if enabled
	if cl.options.WatchConfig {
		cl.startWatching()
	}

	return config, nil
}

// loadConfig performs the actual configuration loading
func (cl *Loader[T]) loadConfig() (*T, error) {
	var config T

	// Try to read config file
	if err := cl.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, ErrConfigFileReadFailed.WithCause(err)
		}
	}

	// Unmarshal configuration
	if err := cl.viper.Unmarshal(&config); err != nil {
		return nil, ErrConfigUnmarshalFailed.WithCause(err)
	}

	return &config, nil
}

// LoadWithValidation loads configuration and runs validation
func (cl *Loader[T]) LoadWithValidation(validator func(*T) error) (*T, error) {
	cl.validator = validator

	config, err := cl.Load()
	if err != nil {
		return nil, err
	}

	if validator != nil {
		if err := validator(config); err != nil {
			return nil, ErrConfigValidationFailed.WithCause(err)
		}
	}

	return config, nil
}

// GetCurrent returns the current configuration (thread-safe)
func (cl *Loader[T]) GetCurrent() *T {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()
	return cl.current
}

// Reload manually reloads the configuration
func (cl *Loader[T]) Reload() error {
	oldConfig := cl.GetCurrent()

	newConfig, err := cl.loadConfig()
	if err != nil {
		cl.notifyError(err)
		return ErrConfigReloadFailed.WithCause(err)
	}

	// Validate new configuration
	if cl.validator != nil {
		if err := cl.validator(newConfig); err != nil {
			validationErr := ErrValidationReloadFailed.WithCause(err)
			cl.notifyError(validationErr)
			return validationErr
		}
	}

	cl.mutex.Lock()
	cl.current = newConfig
	cl.mutex.Unlock()

	// Notify subscribers
	cl.notifyChange(oldConfig, newConfig)

	return nil
}

// startWatching starts watching for configuration file changes
func (cl *Loader[T]) startWatching() {
	cl.viper.WatchConfig()
	cl.viper.OnConfigChange(func(e fsnotify.Event) {
		cl.debouncedReload()
	})
}

// debouncedReload implements debounced configuration reloading
func (cl *Loader[T]) debouncedReload() {
	if cl.debouncer != nil {
		cl.debouncer.Stop()
	}

	cl.debouncer = time.AfterFunc(cl.options.ReloadDebounce, func() {
		_ = cl.Reload()
	})
}

// Subscribe returns a channel that receives configuration change events
func (cl *Loader[T]) Subscribe() <-chan ConfigChangeEvent[T] {
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	ch := make(chan ConfigChangeEvent[T], 10)
	cl.subscribers = append(cl.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscription channel
func (cl *Loader[T]) Unsubscribe(ch <-chan ConfigChangeEvent[T]) {
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	for i, subscriber := range cl.subscribers {
		if subscriber == ch {
			close(subscriber)
			cl.subscribers = append(cl.subscribers[:i], cl.subscribers[i+1:]...)
			break
		}
	}
}

// notifyChange notifies all subscribers about configuration changes
func (cl *Loader[T]) notifyChange(oldConfig, newConfig *T) {
	event := ConfigChangeEvent[T]{
		OldConfig: oldConfig,
		NewConfig: newConfig,
		Timestamp: time.Now(),
	}

	// Call registered callback if provided
	if cl.options.OnConfigChange != nil {
		go cl.options.OnConfigChange(newConfig)
	}

	// Notify subscribers
	cl.subMutex.RLock()
	defer cl.subMutex.RUnlock()

	for _, subscriber := range cl.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

// notifyError notifies about configuration reload errors
func (cl *Loader[T]) notifyError(err error) {
	event := ConfigChangeEvent[T]{
		Error:     err,
		Timestamp: time.Now(),
	}

	// Call registered error callback if provided
	if cl.options.OnConfigChangeError != nil {
		go cl.options.OnConfigChangeError(err)
	}

	// Notify subscribers about error
	cl.subMutex.RLock()
	defer cl.subMutex.RUnlock()

	for _, subscriber := range cl.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

// Stop stops the configuration loader and cleans up resources
func (cl *Loader[T]) Stop() {
	cl.cancel()

	if cl.debouncer != nil {
		cl.debouncer.Stop()
	}

	// Close all subscriber channels
	cl.subMutex.Lock()
	defer cl.subMutex.Unlock()

	for _, subscriber := range cl.subscribers {
		close(subscriber)
	}
	cl.subscribers = nil
}

// GetViper returns the underlying viper instance for advanced usage
func (cl *Loader[T]) GetViper() *viper.Viper {
	return cl.viper
}

// SetConfigValue sets a configuration value programmatically
func (cl *Loader[T]) SetConfigValue(key string, value any) {
	cl.viper.Set(key, value)
}

// GetConfigValue gets a configuration value
func (cl *Loader[T]) GetConfigValue(key string) any {
	return cl.viper.Get(key)
}
