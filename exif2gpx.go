package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/twpayne/go-gpx"
)

func main() {
	flag.Parse()

	points := []*gpx.WptType{}

	err := filepath.Walk(flag.Arg(0), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)

		file, err := os.Open(path)
		if err != nil {
			return nil
		}

		x, err := exif.Decode(file)
		if err != nil {
			return nil
		}

		t, _ := x.DateTime()
		lat, lng, err := x.LatLong()
		if err != nil {
			return nil
		}
		alt, err := x.Get(exif.FieldName("GPSAltitude"))
		var ele float64
		if err != nil {
			ele = 0
		} else {
			num, den, _ := alt.Rat2(0)
			ele = float64(num) / float64(den)
		}

		wpt := gpx.WptType{Lat: lat, Lon: lng, Ele: ele, Time: t, Name: name}
		points = append(points, &wpt)

		return nil
	})

	if err != nil {
		panic(err)
	}

	// pointsのソート
	sort.Slice(points, func(i, j int) bool {
		if points[i].Time.Before(points[j].Time) {
			return true
		} else if points[i].Time == points[j].Time {
			if points[i].Name < points[j].Name {
				return true
			}
		}
		return false
	})

	trkseg := []*gpx.TrkSegType{{TrkPt: points}}
	trk := []*gpx.TrkType{{Name: "Track 1", Desc: "Create by exif2gpx", TrkSeg: trkseg}}

	g := &gpx.GPX{
		Version: "1.1",
		Creator: "go-gpx - https://github.com/twpayne/go-gpx",
		Trk:     trk,
	}
	if _, err := os.Stdout.WriteString(xml.Header); err != nil {
		fmt.Printf("err == %v", err)
	}
	if err := g.WriteIndent(os.Stdout, "", "  "); err != nil {
		fmt.Printf("err == %v", err)
	}
}
