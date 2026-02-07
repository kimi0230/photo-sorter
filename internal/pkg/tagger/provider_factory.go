package tagger

// ProviderFactory creates Tagger provider functions.
type ProviderFactory interface {
	NewProvider() func() (Tagger, error)
}

// DefaultProviderFactory is the default Tagger provider factory.
type DefaultProviderFactory struct{}

func (f *DefaultProviderFactory) NewProvider() func() (Tagger, error) {
	var cachedTagger Tagger
	var cachedTaggerErr error
	return func() (Tagger, error) {
		if cachedTagger != nil || cachedTaggerErr != nil {
			return cachedTagger, cachedTaggerErr
		}
		cachedTagger, cachedTaggerErr = NewTagger()
		return cachedTagger, cachedTaggerErr
	}
}

// NewProviderFactory returns the default ProviderFactory.
func NewProviderFactory() ProviderFactory {
	return &DefaultProviderFactory{}
}
