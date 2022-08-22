package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"
)

func printRemainTime(remain int) {
	hh := int(remain / 60 / 60)
	mm := int((remain - hh*60*60) / 60)
	ss := int((remain - hh*60*60 - mm*60))
	fmt.Printf("\033[2K")
	fmt.Printf("\r%02d:%02d:%02d", hh, mm, ss)
}

func main() {
	var remain int

	if len(os.Args) != 2 {
		log.Fatalln("stopwatch: Wrong number of arguments. Use only one argument.", os.Args[1:])
	}

	v := strings.Split(os.Args[1], ":")
	for i := len(v) - 1; i >= 0; i-- {
		r1, err := strconv.Atoi(v[i])
		if err != nil {
			log.Fatalln(err)
		}

		r2 := r1 * int(math.Pow(60, float64(len(v)-1-i)))
		remain += int(r2)
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	pause := make(chan bool)
	restart := make(chan bool)
	done := make(chan bool)
	right := make(chan bool)
	left := make(chan bool)
	up := make(chan bool)
	down := make(chan bool)

	fmt.Printf("\033[2J")
	fmt.Printf("\033[2;0H")
	fmt.Printf("q:\t quit program\n")
	fmt.Printf("p:\t pause stopwatch\n")
	fmt.Printf("space:\t start stopwatch\n")
	fmt.Printf("↑↓:\t set stopwatch time \n")
	fmt.Printf("←→:\t select hours, minutes, and seconds\n")
	fmt.Printf("\033[6A")

	defer fmt.Printf("\033[2J")
	defer fmt.Printf("\033[0;0H")

	go func() {
		printRemainTime(remain)
		select_hms := 1
		for {
			select {
			case <-done:
				return
			case <-pause:
				ticker.Stop()
			case <-restart:
				ticker.Reset(1000 * time.Millisecond)
			case <-left:
				if select_hms == 3600 {
					select_hms = 1
				} else {
					select_hms *= 60
				}
			case <-right:
				if select_hms == 1 {
					select_hms = 3600
				} else {
					select_hms /= 60
				}
			case <-up:
				ticker.Stop()
				remain = remain + select_hms
			case <-down:
				ticker.Stop()
				if remain < select_hms {
					remain = 0
				} else {
					remain = remain - select_hms
				}

			case <-ticker.C:
				remain = remain - 1
				if remain < 1 {
					ticker.Stop()
				}
			}
			printRemainTime(remain)
			switch select_hms {
			case 1:
				fmt.Printf("\033[0D")
			case 60:
				fmt.Printf("\033[4D")
			case 3600:
				fmt.Printf("\033[7D")
			}
		}
	}()

	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		select {
		case <-done:
			return
		default:
		}

		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}

		switch string(r) {
		case "q":
			done <- true
			return
		case "p":
			pause <- true
		case " ":
			restart <- true
		case "u":
			restart <- true
		case "\033":
			r2, err := tty.ReadRune()
			if err != nil {
				log.Fatal(err)
			}
			if string(r2) != "[" {
				break
			}

			r3, err := tty.ReadRune()
			if err != nil {
				log.Fatal(err)
			}
			switch string(r3) {
			case "A":
				up <- true
			case "B":
				down <- true
			case "C":
				right <- true
			case "D":
				left <- true
			}
		}

	}
}
