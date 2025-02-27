package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type Reactor struct {
	reaction map[string]string
	reader   io.Reader
	writer   io.Writer
}

func NewReactor(reader io.Reader, writer io.Writer) *Reactor {
	return &Reactor{
		reaction: make(map[string]string),
		reader:   reader,
		writer:   writer,
	}
}

func (r *Reactor) AddReaction(command, response string) {
	r.reaction[command] = response
}

func SplitRedisCommand(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	idx := -1
loop:
	for i, b := range data {
		switch b {
		case '\n', '?', ')':
			idx = i
			break loop
		}
	}

	if idx >= 0 {
		token := data[0:idx]
		advance = idx + 1
		return advance, token, nil
	}

	if atEOF {
		if len(data) > 0 {
			return len(data), data, bufio.ErrFinalToken
		} else {
			return 0, nil, nil
		}
	}

	return 0, nil, nil
}

func (r *Reactor) React(cancel func()) {
	scanner := bufio.NewScanner(r.reader)
	scanner.Split(SplitRedisCommand)

	scanCh := make(chan string, 10)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for scanner.Scan() {
			scanCh <- scanner.Text()
		}
		close(scanCh)
	}()

	count := atomic.Int64{}

loop:
	for {
		select {
		case <-ticker.C:
			if v := count.Swap(0); v == 0 {
				log.Debug().Int64("value", v).Msg("no command")
				cancel()
				break loop
			}
		case command, ok := <-scanCh:
			if !ok {
				cancel()
				break loop
			}
			count.Add(1)
			command = strings.TrimSpace(command)
			for k := range r.reaction {
				if strings.Contains(command, k) {
					log.Debug().Msgf("reactor: %s", r.reaction[k])
					fmt.Fprintln(r.writer, r.reaction[k])
					continue loop
				}
			}
		}
	}
}
