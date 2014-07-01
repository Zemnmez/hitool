package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

const format = time.Kitchen

type status int

const (
	ok   status = 0
	fail status = 1
)

type RGB [3]uint8

func (r RGB) To256() uint8 {
	return 16 + (inr(r[0]) * 36) + (inr(r[1]) * 6) + inr(r[2])
}

func inr(a uint8) uint8 {
	return uint8((float64(a) / 255) * 5)
}

func (r RGB) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%#X%#X%#X", r[0], r[1], r[2])), nil
}

func (r RGB) String() string {
	t, _ := r.MarshalText()
	return "#" + string(t)
}

func (r *RGB) UnmarshalText(text []byte) (err error) {
	n, err := fmt.Fscanf(bytes.NewReader(text), "%2x%2x%2x", &r[0], &r[1], &r[2])
	if err != nil {
		return
	}

	if n < 3 {
		err = fmt.Errorf("string %+q is an invalid RGB hex code", text)
	}
	return
}

var (
	colour          string
	backgroundcolor string
	tag             string
	terminal        bool
)

func main() {
	var s status
	defer os.Exit(int(s))

	//Flags
	flag.StringVar(
		&colour,
		"colour",
		"",
		"(foreground)color: returned in 256 colours if no -bgcolour provided, "+
			" otherwise, the foreground color in the :hi statement.",
	)

	flag.StringVar(
		&backgroundcolor,
		"bgcolour",
		"",
		"backgroundcolor for :hi statements. Providing this flag will return a vim :hi statement"+
			", unless -terminal is also set.",
	)

	flag.StringVar(
		&tag,
		"tag",
		"Example",
		"Hi statement tag.",
	)

	flag.BoolVar(
		&terminal,
		"term",
		false,
		"If set, writes an Xterm 256 color colur escape code to stdout",
	)

	flag.Parse()
	for inf := range run() {
		var err error
		switch v := inf.(type) {
		case status:
			s = v

		case error:
			_, err = fmt.Fprintf(os.Stderr, "Error: %s\n", v.Error())

		case string:
			_, err = os.Stdout.Write([]byte(v + "\n"))
		}

		if err != nil {
			panic(fmt.Sprintf("Fatal error: %s\n", err.Error()))
		}
	}

}

func run() (chn <-chan interface{}) {
	c := make(chan interface{})
	chn = c
	go func() {
		defer close(c)

		if colour == "" {
			c <- errors.New("No colour provided (-colour flag)")
			return
		}

		var fg RGB

		if err := fg.UnmarshalText([]byte(colour)); err != nil {
			c <- "Colour parsing failed."
			c <- err
			return
		}

		if backgroundcolor == "" {
			if terminal {
				c <- "\x1b[38;5;" + strconv.Itoa(int(fg.To256())) + "m" + "\x1b[0m"
				return
			}

			c <- strconv.Itoa(int(fg.To256()))
			return
		}

		if backgroundcolor == "transparent" {
			backgroundcolor = ""
		}

		if backgroundcolor == "" {

			c <- fmt.Sprintf("hi %s guifg=%s ctermfg=%s", tag, colour, strconv.Itoa(int(fg.To256())))
			return
		}

		var bg RGB
		if err := bg.UnmarshalText([]byte(backgroundcolor)); err != nil {
			c <- "Background colour parsing failed."
			c <- err
			return
		}

		c <- fmt.Sprintf(
			"hi %s guifg=%s guibg=%s ctermfg=%s ctermbg=%s",
			tag,
			colour,
			strconv.Itoa(int(fg.To256())),
			backgroundcolor,
			strconv.Itoa(int(bg.To256())),
		)
	}()
	return
}
