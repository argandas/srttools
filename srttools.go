package srttools

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var (
	regexpID   = regexp.MustCompile("^[0-9]+$")
	regexpTime = regexp.MustCompile("^([0-9]+):([0-9]+):([0-9]+).*")
)

const (
	timeFormat   string = "HH:MM:SS,MIL"
	timeSep      string = " --> "
	timePreffMin string = "HH:"
	timePreffSec string = "HH:MM:"
	timePreffMil string = "HH:MM:SS,"
)

type srtTime struct {
	start int
	end   int
	extra string
}

func (t *srtTime) String() string {
	return timeMilli2Str(t.start) + timeSep + timeMilli2Str(t.end) + t.extra
}
func (t *srtTime) Offset(offset int) {
	t.start += offset
	t.end += offset
}

func Concat(output string, files ...string) error {
	// Create file
	file, err := os.Create(output)
	defer func() {
		file.Sync()
		file.Close()
	}()
	if err != nil {
		return err
	}
	// Create writter
	w := bufio.NewWriter(file)
	defer w.Flush()
	// At least 2 files are needed
	if len(files) < 2 {
		return nil
	}
	// Set offsets
	timeOff := 0
	idOff := 0
	// Parse each file
	for index, file := range files {
		if index == 0 {
			// First file, written as it is
			copySRTLines(w, file)
			continue
		}
		// Second file and beyond...
		t, id, err := getSRTlimits(files[index-1])
		if err != nil {
			return err
		}
		// Increment offset
		timeOff += t
		idOff += id
		// Apply offset
		addSRToffset(w, file, timeOff, idOff)
	}
	return nil
}

func copySRTLines(w *bufio.Writer, path string) error {
	// Open SRT file
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	// Create a reader
	r := bufio.NewReader(file)
	for {
		// Read new line
		line, _, err := r.ReadLine()
		// Write to file
		if _, err := w.WriteString(string(line) + "\n"); err != nil {
			return err
		}
		// Check errors
		if err != nil {
			break
		}
	}
	return nil
}

func getSRTlimits(srt string) (int, int, error) {
	// Last readed time
	lasTime := srtTime{}
	var lastID int
	// Open SRT file
	file, err := os.Open(srt)
	defer file.Close()
	if err != nil {
		return 0, 0, err
	}
	// Create a reader
	reader := bufio.NewReader(file)
	for {
		// Read new line
		line, _, err := reader.ReadLine()
		if regexpID.MatchString(string(line)) {
			// Parse ID
			lastID, _ = parseID(string(line))
		}
		// Match time lines
		if regexpTime.MatchString(string(line)) {
			// Parse Time line
			lasTime = parseTime(string(line))
		}
		// Check errors
		if err != nil {
			break
		}
	}
	return lasTime.end, lastID, nil
}

func addSRToffset(w *bufio.Writer, srt string, timeOff, idOff int) error {
	// Open SRT file
	file, err := os.Open(srt)
	defer file.Close()
	if err != nil {
		return err
	}
	// Create a reader
	reader := bufio.NewReader(file)
	for {
		// Read new line
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		switch {
		case regexpID.MatchString(string(line)):
			// Parse ID
			id, _ := parseID(string(line))
			line = []uint8(fmt.Sprintf("%d", id+idOff))
		case regexpTime.MatchString(string(line)):
			// Parse Times
			time := parseTime(string(line))
			time.Offset(timeOff)
			line = []uint8(time.String())
		default:
			// Do nothing
		}
		// Write to file
		if _, err := w.WriteString(string(line) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func parseID(line string) (int, error) {
	return strconv.Atoi(line)
}

func parseTime(line string) srtTime {
	// Trim times
	init := line[:len(timeFormat)]
	finish := line[len(timeFormat)+len(timeSep):][:len(timeFormat)]
	// Convert to milliseconds
	return srtTime{
		start: timeStr2Milli(init),
		end:   timeStr2Milli(finish),
		extra: line[2*len(timeFormat)+len(timeSep):],
	}
}

func timeStr2Milli(time string) int {
	millis := 0
	// Convert hours
	if hours, err := strconv.Atoi(time[:2]); err != nil {
		panic(err)
	} else {
		millis += hours * 60 * 60 * 1000
	}
	// Convert minutes
	if minutes, err := strconv.Atoi(time[len(timePreffMin):][:2]); err != nil {
		panic(err)
	} else {
		millis += minutes * 60 * 1000
	}
	// Convert seconds
	if seconds, err := strconv.Atoi(time[len(timePreffSec):][:2]); err != nil {
		panic(err)
	} else {
		millis += seconds * 1000
	}
	// Convert seconds
	if milli, err := strconv.Atoi(time[len(timePreffMil):]); err != nil {
		panic(err)
	} else {
		millis += milli
	}
	return millis
}

func timeMilli2Str(millis int) string {
	sec := uint(millis / 1000)
	min := uint(sec / 60)
	sec = sec % 60
	hrs := uint(min / 60)
	min = min % 60
	return fmt.Sprintf("%s:%s:%s,%s", time2Str(hrs), time2Str(min), time2Str(sec), millis2Str(millis%1000))
}

func time2Str(t uint) string {
	if t < 10 {
		return fmt.Sprintf("0%d", t)
	}
	return fmt.Sprintf("%d", t)
}

func millis2Str(t int) string {
	if t < 10 {
		return fmt.Sprintf("00%d", t)
	}
	if t < 100 {
		return fmt.Sprintf("0%d", t)
	}
	return fmt.Sprintf("%d", t)
}
