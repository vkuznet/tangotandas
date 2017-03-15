package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// colors
const BLACK = "0;30m"
const RED = "0;31m"
const GREEN = "0;32m"
const BROWN = "0;33m"
const BLUE = "0;34m"
const PURPLE = "0;35m"
const CYAN = "0;36m"
const LIGHT_PURPLE = "1;35m"
const LIGHT_CYAN = "1;36m"

const BOLD = "\x1b[1m"
const PLAIN = "\x1b[0m"

const MUSIC_NOTE = "\u266B"
const MUSIC_COUPLE = "\u8E40"
const MUSIC_BREAK2 = "\u1F377"
const MUSIC_BREAK = "\u2615"

// helper function to convert color constant into html color string
func colorHtml(col string) string {
	cmap := make(map[string]string)
	cmap[BLACK] = "black"
	cmap[RED] = "red"
	cmap[GREEN] = "green"
	cmap[BROWN] = "brown"
	cmap[BLUE] = "blue"
	cmap[PURPLE] = "purple"
	cmap[CYAN] = "cyan"
	cmap[LIGHT_PURPLE] = "purple"
	cmap[LIGHT_CYAN] = "cyan"
	c, ok := cmap[col]
	if ok {
		return c
	}
	return "black"
}

// helper function to generate colored text for terminal
func color(text, col string) string {
	return BOLD + "\x1b[" + col + text + PLAIN
}

// Execute command
func exe(cmd []string) string {
	out, err := exec.Command("osascript", cmd...).CombinedOutput()
	if err != nil {
		log.Println("output", string(out))
		log.Fatal(err)
	}
	return string(out)
}

// Fetch all tracks from current playlist
func tracks() []string {
	cmd := []string{"-e", "tell application \"iTunes\"", "-e", "set trackNames to {}", "-e", "repeat with aName in name of tracks of current playlist", "-e", "set trackNames to trackNames & \"SEPARATOR\" & aName", "-e", "end repeat", "-e", "end tell"}
	res := exe(cmd)
	var out []string
	for i, val := range strings.Split(res, "SEPARATOR,") {
		if i == 0 {
			continue
		}
		// use string w/o last character which is a common separator from AppleScript
		v := strings.Trim(val[0:len(val)-2], " ")
		out = append(out, v)
	}
	return out
	// keep this as an example of AppleScript command, the problem is that
	// it does not distinguish comma in AppleScript output and track name
	//     cmd := []string{"-e", "tell application \"iTunes\" to get name of every track in current playlist"}
	//     return strings.Split(exe(cmd), ", ")
}

// Fetch all artists from current playlist
func artists() []string {
	cmd := []string{"-e", "tell application \"iTunes\" to get artist of every track in current playlist"}
	return strings.Split(exe(cmd), ", ")
}

// Fetch all genres from current playlist
func genres() []string {
	cmd := []string{"-e", "tell application \"iTunes\" to get genre of every track in current playlist"}
	return strings.Split(exe(cmd), ", ")
}

type Song struct {
	Artist string
	Track  string
	Genre  string
}

func (s *Song) String() string {
	return fmt.Sprintf("Name: %s, Artist: %s, Genre: %s", s.Track, s.Artist, s.Genre)
}

// Fetch current track"
func track() Song {
	cmd := []string{"-e", "tell application \"iTunes\"", "-e", "tell current track to artist & tab & name & tab & genre", "-e", "end tell"}
	trk := exe(cmd)
	arr := strings.Split(trk, "\t")
	artist := strings.Replace(arr[0], "\n", "", -1)
	track := strings.Replace(arr[1], "\n", "", -1)
	genre := strings.Replace(arr[2], "\n", "", -1)
	return Song{Artist: artist, Track: track, Genre: genre}
}

// Human readable output for given song
func getSong(song Song, col, pad, output string) string {
	var track, msg string
	if song.Genre == "Cortina" {
		track = strings.Replace(song.Track, "z_", "", 0)
		msg = fmt.Sprintf("%sCortina: %s, %s", pad, song.Artist, track)
	} else {
		track = song.Track
		msg = fmt.Sprintf("%s%s, %s", pad, song.Artist, track)
	}
	if output == "html" {
		m := fmt.Sprintf("<span style=\"font-weight: bold;color:%s\">%s</span></br>", colorHtml(col), msg)
		return m
	}
	if col != "" {
		return color(msg, col)
	}
	return msg
}

// Human readable format for tanda
func getTanda(tanda []Song, trk Song, prefix, output string) string {
	var out []string
	if output == "html" {
		m := fmt.Sprintf("<div id=\"tanda\" name=\"tanda\" class=\"tanda\">\n")
		out = append(out, m)
	}
	var title string
	col := BLACK
	for _, song := range tanda {
		genre := strings.ToLower(song.Genre)
		pad := "  "
		if title == "" {
			title = fmt.Sprintf("%s: %s\n", prefix, strings.Title(genre))
			if output == "html" {
				title = fmt.Sprintf("<h3>%s</h3>", title)
			}
			//             fmt.Println(title)
			out = append(out, title)
		}
		if song == trk {
			pad = MUSIC_NOTE + " "
			if genre == "tango" {
				col = RED
			} else if genre == "vals" {
				col = GREEN
			} else if genre == "milonga" || genre == "tango foxtrot" {
				col = PURPLE
			} else {
				col = LIGHT_CYAN
				pad = MUSIC_BREAK + " "
			}
		} else {
			col = BLACK
		}
		s := getSong(song, col, pad, output)
		out = append(out, s)
	}
	if output == "html" {
		out = append(out, "\n</div>")
	}
	return strings.Join(out, "\n")
}

// Clear terminal screen
func clear(output string) {
	if output != "html" {
		fmt.Println("\x1b[2J")
	}

	//     if output == "html" {
	//             fmt.Println(strings.Repeat("\n</br>", 10))
	//         fmt.Println("\n<div style=\"padding:100px\"></div>")
	//     } else {
	//         fmt.Println("\x1b[2J")
	//     }
}

// Time reminder
func timeReminder(dj string, startTime, timeOffset int64, output string) string {
	dt1 := startTime - timeOffset
	dt2 := time.Now().Unix()
	tdiff := dt2 - dt1
	var hours, minutes, seconds int64
	hour := int64(60 * 60)
	minute := int64(60)
	for {
		if tdiff > hour {
			hours += 1
			tdiff -= hour
		} else if tdiff > minute {
			minutes += 1
			tdiff -= minute
		} else if tdiff > 0 {
			seconds += 1
			tdiff -= 1
		} else {
			break
		}
	}
	lst := fmt.Sprintf("%d hours, %d minutes and %d seconds", hours, minutes, seconds)
	msg := fmt.Sprintf("\nDJ %s: %s", dj, lst)
	if output == "html" {
		return fmt.Sprintf("<h5>%s</h5>", msg)
	}
	return color(msg, LIGHT_PURPLE)
}

// compare two tandas
func identicalTandas(t1, t2 []Song) bool {
	if len(t1) != len(t2) {
		return false
	}
	out := true
	for i := 0; i < len(t1); i++ {
		if t1[i] != t2[i] {
			return false
		}
	}
	return out
}

// Fetch playlist from itunes
func playlist(dj string, startTime, timeOffset int64, output, style string) string {
	itracks := tracks()
	iartists := artists()
	igenres := genres()
	itrack := track()
	var ptracks []Song
	if len(itracks) != len(iartists) || len(itracks) != len(igenres) || len(iartists) != len(igenres) {
		msg := fmt.Sprintf("Please check your playlist, # tracks %d, # artists %d, # genres %d", len(itracks), len(iartists), len(igenres))
		for i, s := range itracks {
			fmt.Println(i, s)
		}
		log.Fatal(msg)
	}
	for idx := 0; idx < len(itracks); idx++ {
		track := strings.Replace(itracks[idx], "\n", "", -1)
		artist := strings.Replace(iartists[idx], "\n", "", -1)
		genre := strings.Replace(igenres[idx], "\n", "", -1)
		ptracks = append(ptracks, Song{Track: track, Artist: artist, Genre: genre})
	}
	var tanda, oldTanda []Song
	var emptySong Song
	exit := false
	for _, song := range ptracks {
		if song.Track == itrack.Track {
			exit = true
		}
		if song.Genre == "Cortina" && exit {
			tanda = append(tanda, song)
			break
		}
		if song.Genre == "Cortina" {
			oldTanda = tanda
			tanda = []Song{}
			continue
		}
		tanda = append(tanda, song)
	}
	st := getHeader(style, output)
	if exit && !identicalTandas(oldTanda, tanda) {
		clear(output)
		st += getTanda(tanda, itrack, "CURRENT TANDA", output)
		// find next tanda
		var nextTanda []Song
		for idx := 0; idx < len(ptracks); idx++ {
			song := ptracks[idx]
			if song == tanda[len(tanda)-1] {
				for jdx := 0; jdx < len(ptracks); jdx++ {
					if jdx <= idx {
						continue
					}
					song = ptracks[jdx]
					nextTanda = append(nextTanda, song)
					if song.Genre == "Cortina" {
						break
					}
				}
			}
		}
		st += getTanda(nextTanda, emptySong, "\n\nNEXT TANDA", output)
		st += fmt.Sprintf("\n%s\n", timeReminder(dj, startTime, timeOffset, output))
	}
	st += getFooter(output)
	return st
}

// helper function to read/write style
func getStyle() string {
	var style, fname, hdir string
	for _, item := range os.Environ() {
		val := strings.Split(item, "=")
		if val[0] == "HOME" {
			hdir = fmt.Sprintf("%s/.tangotandas", val[1])
			fname = fmt.Sprintf("%s/styles.css", hdir)
		}
	}
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		style = "body { background-color: #F8E0A9; padding: 10px; font-size: 20px;}"
		if _, err := os.Stat(hdir); os.IsNotExist(err) {
			err := os.Mkdir(hdir, 0744)
			if err != nil {
				log.Fatal(err)
			}
		}
		err := ioutil.WriteFile(fname, []byte(style), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	data, err := ioutil.ReadFile(fname)
	if err == nil {
		style = string(data)
	}
	return style
}

func getHeader(style, output string) string {
	var out string
	if output == "html" {
		head := fmt.Sprintf("<style type=\"text/css\">%s</style>\n", style)
		out = fmt.Sprintf("<html><head>\n%s\n</head><body>\n", head)
	}
	return out
}

func getFooter(output string) string {
	var out string
	if output == "html" {
		out = fmt.Sprintf("</body></html>")
	}
	return out
}

func main() {
	// server options
	var dj string
	flag.StringVar(&dj, "dj", "", "dj name")
	var tOffset int64
	flag.Int64Var(&tOffset, "tOffset", 0, "time offset")
	var output string
	flag.StringVar(&output, "output", "text", "output type: text or html")
	flag.Parse()

	startTime := time.Now().Unix()
	if dj == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		dj = u.Name
	}
	style := getStyle()
	for {
		tandas := playlist(dj, startTime, tOffset, output, style)
		fmt.Println(tandas)
		time.Sleep(time.Duration(1000) * time.Millisecond)
	}
}
