package main

import "github.com/argandas/srttools"

func main() {
	srttools.Concat(
		"Star Wars - Episode II.srt",
		"Star Wars - Episode II - CD1.srt",
		"Star Wars - Episode II - CD2.srt",
		"Star Wars - Episode II - CD3.srt",
	)
}
