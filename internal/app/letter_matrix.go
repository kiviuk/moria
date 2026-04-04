package app

type LetterTuple struct {
	Letter         string
	LetterPosition int
	LetterGroup    int
}

type MagicSpell struct {
	Spell string
}

func LetterGroup(letter string) int {
	if len(letter) == 0 {
		return 0
	}
	r := rune(letter[0])
	selected := rune(0)
	if r >= 'A' && r <= 'Z' {
		selected = 'A'
	} else if r >= 'a' && r <= 'z' {
		selected = 'a'
	}
	if selected == 0 {
		return 0
	}
	return int(r-selected)/GroupSize + 1
}

func (m MagicSpell) LetterTuples() []LetterTuple {
	tuples := make([]LetterTuple, len(m.Spell))
	for i, r := range m.Spell {
		tuples[i] = LetterTuple{Letter: string(r), LetterPosition: i, LetterGroup: LetterGroup(string(r))}
	}
	return tuples
}

func ModN(value int, n int) int {
	return value % n
}

func (m LetterTuple) MapModN() LetterTuple {
	return LetterTuple{
		Letter:         m.Letter,
		LetterPosition: ModN(m.LetterPosition, MatrixN),
		LetterGroup:    m.LetterGroup,
	}
}
