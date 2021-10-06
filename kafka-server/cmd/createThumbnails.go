package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/floostack/transcoder/ffmpeg"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
)

var thumbnailCount int

// createThumbnailsCmd represents the createThumbnails command
var createThumbnailsCmd = &cobra.Command{
	Use:   "create-thumbnails",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := badger.Open(badger.DefaultOptions(databasePath))
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()
		movie, err := createThumbnailsRun(db)
		if err != nil {
			log.Fatalln(err)
		}
		j, err := json.Marshal(movie)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(j))
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if databasePath, err = filepath.Abs(filepath.Clean(databasePath)); err != nil {
			return err
		}
		if moviePath, err = filepath.Abs(filepath.Clean(moviePath)); err != nil {
			return err
		}
		if _, err = os.Stat(databasePath); err != nil {
			return err
		}
		if _, err = os.Stat(moviePath); err != nil {
			return err
		}
		return nil
	},
}

func createThumbnailsRun(db *badger.DB) (movieModel, error) {
	movie := movieModel{}
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("movie|%s", movieId)))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &movie)
		})
	}); err != nil {
		return movie, err
	}

	exe, err := os.Executable()
	if err != nil {
		return movie, err
	}
	ffConf := ffmpeg.Config{
		FfmpegBinPath:   filepath.Join(filepath.Dir(exe), "ffmpeg-4.4-full_build", "bin", "ffmpeg.exe"),
		FfprobeBinPath:  filepath.Join(filepath.Dir(exe), "ffmpeg-4.4-full_build", "bin", "ffprobe.exe"),
		ProgressEnabled: false,
		Verbose:         false,
	}
	meta, err := ffmpeg.New(&ffConf).Input(movie.Path).GetMetadata()
	if err != nil {
		return movie, err
	}
	duration, err := strconv.ParseFloat(meta.GetFormat().GetDuration(), 32)
	if err != nil {
		return movie, err
	}
	startPositionPercent := 5
	endPositionPercent := 95
	addPercent := (endPositionPercent - startPositionPercent) / (thumbnailCount - 1)
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return movie, err
	}
	defer os.RemoveAll(tmpDir)
	for i := 0; i < thumbnailCount; i++ {
		outPNG := path.Join(tmpDir, fmt.Sprintf("%d.png", i))
		opt := []string{
			"-ss", fmt.Sprintf("%d", int(duration)*(startPositionPercent+addPercent*i)/100),
			"-i", fmt.Sprintf("%s", movie.Path),
			"-vframes", "1",
			"-vf", "scale=300:-1",
			"-y",
			outPNG,
		}
		out, err := exec.Command(ffConf.FfmpegBinPath, opt...).CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return movie, err
		}
		b, err := ioutil.ReadFile(outPNG)
		if err != nil {
			return movie, err
		}
		if err := db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("thumbnail|%s|%d", movie.Id, i)), b)
		}); err != nil {
			return movie, err
		}
	}
	return movie, nil
}

func init() {
	rootCmd.AddCommand(createThumbnailsCmd)
	createThumbnailsCmd.Flags().StringVar(&databasePath, "database", "", "database path")
	createThumbnailsCmd.MarkFlagRequired("database")
	createThumbnailsCmd.Flags().StringVar(&movieId, "movieid", "", "movieModel ID")
	createThumbnailsCmd.MarkFlagRequired("movieid")
	createThumbnailsCmd.Flags().IntVar(&thumbnailCount, "count", 0, "count")
	createThumbnailsCmd.MarkFlagRequired("count")
}
