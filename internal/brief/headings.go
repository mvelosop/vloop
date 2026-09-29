package brief

// HeadingSet is the language-specific part of the rules: the section headings
// and the phrasings of the task-count and exit-code rules. Patterns are
// regular expressions matched case-insensitively (accents still matter)
// against the whole brief. vloop's own messages are never translated.
type HeadingSet struct {
	// Sections is the full heading table, in the order a brief writes them.
	Sections      []string
	BindingRefs   string
	WorkedExample string
	OutOfScope    string
	Constraints   string
	TaskCount     string
	ExitCodes     string
}

// English is the heading set for language "en".
var English = &HeadingSet{
	Sections: []string{
		"What it is", "Why this shape, and what was rejected", "Binding references",
		"Behaviour contract", "Worked example", "Out of scope", "Constraints", "Shape",
	},
	BindingRefs:   "Binding references",
	WorkedExample: "Worked example",
	OutOfScope:    "Out of scope",
	Constraints:   "Constraints",
	TaskCount:     `[0-9]+ (to|–|-) ?[0-9]* ?tasks|[a-z]+ to [a-z]+ tasks|[0-9]+ tasks`,
	ExitCodes:     `exit (code|[0-9])|\b(200|201|302|400|404|409|500)\b|exits? [0-9]`,
}

// Spanish is the heading set for language "es".
var Spanish = &HeadingSet{
	Sections: []string{
		"Qué es", "Por qué esta forma, y qué se descartó", "Referencias vinculantes",
		"Contrato de comportamiento", "Ejemplo trabajado", "Fuera de alcance", "Restricciones", "Forma",
	},
	BindingRefs:   "Referencias vinculantes",
	WorkedExample: "Ejemplo trabajado",
	OutOfScope:    "Fuera de alcance",
	Constraints:   "Restricciones",
	TaskCount:     `[0-9]+ (a|–|-) ?[0-9]* ?tareas|[0-9]+ tareas`,
	ExitCodes:     `código de salida [0-9]|\b(200|201|302|400|404|409|500)\b|sale con [0-9]`,
}

// sets maps a config language to its heading set.
var sets = map[string]*HeadingSet{"en": English, "es": Spanish}

// SetFor returns the heading set of a language. There is no cross-language
// fallback: config only admits languages in sets, so English is the default
// for an unset value alone.
func SetFor(language string) *HeadingSet {
	if s, ok := sets[language]; ok {
		return s
	}
	return English
}
