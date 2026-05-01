package audio

import (
	"strings"
)

// StreamChunker primește fragmente de text (tokeni de la un LLM) și emite propoziții întregi
type StreamChunker struct {
	Input  chan string
	Output chan string
	buffer strings.Builder
}

// NewStreamChunker instanțiază și pornește procesul de grupare a textului
func NewStreamChunker() *StreamChunker {
	c := &StreamChunker{
		Input:  make(chan string, 100),
		Output: make(chan string, 100),
	}
	go c.run()
	return c
}

// run procesează tokenii pe măsură ce sosesc
func (c *StreamChunker) run() {
	defer close(c.Output)

	for token := range c.Input {
		c.buffer.WriteString(token)
		
		// Dacă am acumulat text, verificăm ultima literă pentru punctuație
		currentText := c.buffer.String()
		if len(currentText) > 0 {
			lastRune := rune(currentText[len(currentText)-1])
			// Tăiem la punct, semnul exclamării, semnul întrebării sau newline
			if lastRune == '.' || lastRune == '!' || lastRune == '?' || lastRune == '\n' {
				// Trimitem propoziția formată și curățăm bufferul
				c.Output <- strings.TrimSpace(currentText)
				c.buffer.Reset()
			}
		}
	}

	// Când canalul de intrare e închis, golim orice a mai rămas în buffer
	remaining := strings.TrimSpace(c.buffer.String())
	if remaining != "" {
		c.Output <- remaining
	}
}

// Close închide fluxul de intrare
func (c *StreamChunker) Close() {
	close(c.Input)
}

// Helper func pentru teste: împarte direct un string mare
func SplitTextIntoSentences(text string) []string {
	chunker := NewStreamChunker()
	
	// Simulăm stream-ul introducând literă cu literă sau cuvânt cu cuvânt
	go func() {
		for _, r := range text {
			chunker.Input <- string(r)
		}
		chunker.Close()
	}()

	var results []string
	for sentence := range chunker.Output {
		if strings.TrimSpace(sentence) != "" {
			results = append(results, sentence)
		}
	}
	return results
}
