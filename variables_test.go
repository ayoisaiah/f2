package f2

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/djherbis/times.v1"
)

func randate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func TestAutoIncrementingNumber(t *testing.T) {
	// TODO: Write test
}

func TestReplaceFilenameVariables(t *testing.T) {
	testDir := setupFileSystem(t)

	for _, path := range fileSystem {
		fullPath := filepath.Join(testDir, path)
		base := filenameWithoutExtension(filepath.Base(path))
		ext := filepath.Ext(path)
		dir := filepath.Dir(path)
		ch := Change{
			BaseDir: filepath.Join(testDir, dir),
			Source:  filepath.Base(path),
		}

		op := &Operation{}
		got, err := op.handleVariables("{{p}}-{{f}}{{ext}}", ch)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		want := fmt.Sprintf(
			"%s-%s%s",
			filepath.Base(filepath.Dir(fullPath)),
			base,
			ext,
		)
		if got != want {
			t.Fatalf("Expected: %s, but got: %s", want, got)
		}
	}
}

func TestReplaceDateVariables(t *testing.T) {
	// TODO: Noticed some intermittent failures
	testDir := setupFileSystem(t)

	for _, file := range fileSystem {
		path := filepath.Join(testDir, file)

		// change the atime and mtime to a random value
		mtime, atime := randate(), randate()
		err := os.Chtimes(path, atime, mtime)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		timeInfo, err := times.Stat(path)
		if err != nil {
			t.Fatalf("Expected no errors, but got one: %v\n", err)
		}

		want := make(map[string]string)
		got := make(map[string]string)

		accessTime := timeInfo.AccessTime()
		modTime := timeInfo.ModTime()

		fileTimes := []string{"mtime", "atime", "ctime", "now", "btime"}

		for _, v := range fileTimes {
			var timeValue time.Time
			switch v {
			case "mtime":
				timeValue = modTime
			case "atime":
				timeValue = accessTime
			case "ctime":
				timeValue = modTime
				if timeInfo.HasChangeTime() {
					timeValue = timeInfo.ChangeTime()
				}
			case "btime":
				timeValue = modTime
				if timeInfo.HasBirthTime() {
					timeValue = timeInfo.BirthTime()
				}
			case "now":
				timeValue = time.Now()
			}

			for key, token := range dateTokens {
				want[v+"."+key] = timeValue.Format(token)
				out, err := replaceDateVariables(path, "{{"+v+"."+key+"}}")
				if err != nil {
					t.Fatalf("Expected no errors, but got one: %v\n", err)
				}
				got[v+"."+key] = out
			}
		}

		if !cmp.Equal(want, got) {
			t.Fatalf("Expected %v, but got %v\n", want, got)
		}
	}
}

func TestReplaceExifVariables(t *testing.T) {
	rootDir := filepath.Join("testdata", "images")

	type FileExif struct {
		Year         string `json:"year"`
		Make         string `json:"make"`
		Model        string `json:"model"`
		ISO          int    `json:"iso"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		Dimensions   string `json:"dimensions"`
		ExposureTime string `json:"exposure_time"`
		FocalLength  string `json:"focal_length"`
		Aperture     string `json:"aperture"`
	}

	cases := []testCase{
		{
			name: "Use EXIF data to rename CR2 file",
			want: []Change{
				{
					Source:  "tractor-raw.cr2",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"tractor-raw.cr2",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.w}}_h{{exif.h}}_{{exif.wh}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
		{
			name: "Use EXIF data to rename JPEG file",
			want: []Change{
				{
					Source:  "bike.jpeg",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"bike.jpeg",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.w}}_h{{exif.h}}_{{exif.wh}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
		{
			name: "Use EXIF data to rename DNG file",
			want: []Change{
				{
					Source:  "proraw.dng",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"proraw.dng",
				"-r",
				"{{exif.dt.YYYY}}_{{exif.make}}_{{exif.model}}_{{exif.iso}}_w{{exif.h}}_h{{exif.w}}_{{exif.h}}x{{exif.w}}_{{exif.et}}_{{exif.fl}}_{{exif.fnum}}{{ext}}",
				rootDir,
			},
		},
	}

	for _, c := range cases {
		f := filenameWithoutExtension(c.want[0].Source)
		ext := filepath.Ext(c.want[0].Source)

		jsonFile, err := os.ReadFile(filepath.Join(rootDir, f+".json"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var exif FileExif
		err = json.Unmarshal(jsonFile, &exif)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		target := fmt.Sprintf(
			"%s_%s_%s_ISO%d_w%d_h%d_%s_%ss_%smm_f%s%s",
			exif.Year,
			exif.Make,
			exif.Model,
			exif.ISO,
			exif.Width,
			exif.Height,
			exif.Dimensions,
			exif.ExposureTime,
			exif.FocalLength,
			exif.Aperture,
			ext,
		)

		c.want[0].Target = target
	}

	runFindReplace(t, cases)
}

func TestReplaceID3Variables(t *testing.T) {
	rootDir := filepath.Join("testdata", "audio")

	type FileID3 struct {
		Format      string `json:"format"`
		FileType    string `json:"file_type"`
		Title       string `json:"title"`
		Album       string `json:"album"`
		Artist      string `json:"artist"`
		AlbumArtist string `json:"album_artist"`
		Genre       string `json:"genre"`
		Year        string `json:"year"`
		Track       string `json:"track"`
		TotalTracks string `json:"total_tracks"`
		Disc        string `json:"disc"`
		TotalDiscs  string `json:"total_discs"`
	}

	cases := []testCase{
		{
			name: "Use ID3 tags to rename an mp3 file",
			want: []Change{
				{
					Source:  "sample_mp3.mp3",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"sample_mp3.mp3",
				"-r",
				"{{id3.title}}_{{id3.artist}}_{{id3.format}}_{{id3.type}}_{{id3.album}}_{{id3.album_artist}}_{{id3.track}}_{{id3.total_tracks}}_{{id3.disc}}_{{id3.total_discs}}_{{id3.year}}",
				rootDir,
			},
		},
		{
			name: "Use ID3 tags to rename an ogg file",
			want: []Change{
				{
					Source:  "sample_ogg.ogg",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"sample_ogg.ogg",
				"-r",
				"{{id3.title}}_{{id3.artist}}_{{id3.format}}_{{id3.type}}_{{id3.album}}_{{id3.album_artist}}_{{id3.track}}_{{id3.total_tracks}}_{{id3.disc}}_{{id3.total_discs}}_{{id3.year}}",
				rootDir,
			},
		},
		{
			name: "Use ID3 tags to rename a flac file",
			want: []Change{
				{
					Source:  "sample_flac.flac",
					BaseDir: rootDir,
				},
			},
			args: []string{
				"-f",
				"sample_flac.flac",
				"-r",
				"{{id3.title}}_{{id3.artist}}_{{id3.format}}_{{id3.type}}_{{id3.album}}_{{id3.album_artist}}_{{id3.track}}_{{id3.total_tracks}}_{{id3.disc}}_{{id3.total_discs}}_{{id3.year}}",
				rootDir,
			},
		},
	}

	for _, c := range cases {
		f := filenameWithoutExtension(c.want[0].Source)

		jsonFile, err := os.ReadFile(filepath.Join(rootDir, f+".json"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		var id3 FileID3
		err = json.Unmarshal(jsonFile, &id3)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		target := fmt.Sprintf(
			"%s_%s_%s_%s_%s_%s_%s_%s_%s_%s_%s",
			id3.Title,
			id3.Artist,
			id3.Format,
			id3.FileType,
			id3.Album,
			id3.AlbumArtist,
			id3.Track,
			id3.TotalTracks,
			id3.Disc,
			id3.TotalDiscs,
			id3.Year,
		)

		c.want[0].Target = target
	}

	runFindReplace(t, cases)
}

func TestFileHash(t *testing.T) {
	testDir := filepath.Join("testdata", "images")

	cases := []testCase{
		{
			name: "Replace md5 and sha1 hash",
			want: []Change{
				{
					Source:  "bike.jpeg",
					BaseDir: testDir,
					Target:  "6801e3de5f584028b8cd4292c6eca7ba_5b97fd595c700277315742bc91ac0ae67e5eb7a3",
				},
			},
			args: []string{
				"-f",
				"bike.jpeg",
				"-r",
				"{{hash.md5}}_{{hash.sha1}}",
				testDir,
			},
		},
		{
			name: "Replace sha256 and sha512 hash",
			want: []Change{
				{
					Source:  "proraw.dng",
					BaseDir: testDir,
					Target:  "55195ff447785e9af9dea2b0e4f3dc1e991f19dc224413f7a3e5718efb980d99_d53831330e6a70899ad36cbde793284d2cd0332ef090cf20dae86299ec9b8f5b50e06becd8bfadb65fce001d3fedb811d02d751cd9a8279cbaf88b46d25b6408",
				},
			},
			args: []string{
				"-f",
				"proraw.dng",
				"-r",
				"{{hash.sha256}}_{{hash.sha512}}",
				testDir,
			},
		},
	}

	runFindReplace(t, cases)
}

func TestRepalceRandomVariable(t *testing.T) {
	slice := []string{
		`{{10r_l}}`,
		`{{8r_d}}`,
		`{{9r_l}}`,
		`{{5r_ld}}`,
		`{{r}}`,
	}

	for _, v := range slice {
		submatches := randomRegex.FindAllStringSubmatch(v, -1)
		strLen := submatches[0][1]
		length := 10
		var err error
		if strLen != "" {
			length, err = strconv.Atoi(strLen)
			if err != nil {
				t.Fatalf("Test (%s) — Unexpected error: %v", v, err)
			}
		}

		str, err := replaceRandomVariable(v)
		if err != nil {
			t.Fatalf("Test (%s) — Unexpected error: %v", v, err)
		}

		if len(str) != length {
			t.Fatalf(
				"Test (%s) — Expected length of random string to be %d, got: %d",
				v,
				length,
				len(str),
			)
		}
	}
}
