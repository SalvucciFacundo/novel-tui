package domain

// Spellchecker defines the interface for checking word validity and generating spelling suggestions.
type Spellchecker interface {
	Check(word string) bool
	Suggestions(word string) []string
}

// DictionaryManager defines the interface for loading dictionaries and adding custom words.
type DictionaryManager interface {
	LoadDictionary(affPath, dicPath string) error
	AddCustomWord(word string) error
	AvailableDictionaries() []string
}
