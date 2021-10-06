package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"time"
)

var databasePath string
var moviePath string

type movieModel struct {
	Id               string    `json:"id"`
	Path             string    `json:"path"`
	LastModifiedTime time.Time `json:"last_modified_time"`
}

// addMovieCmd represents the addMovie command
var addMovieCmd = &cobra.Command{
	Use:   "add-movie",
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
		movie, err := addMovieRun(db)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(json.Marshal(movie))
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if databasePath, err = filepath.Abs(filepath.Clean(databasePath)); err != nil {
			return err
		}
		if moviePath, err = filepath.Abs(filepath.Clean(moviePath)); err != nil {
			return err
		}
		if _, err = os.Stat(moviePath); err != nil {
			return err
		}
		return nil
	},
}

func addMovieRun(db *badger.DB) (movieModel, error) {
	movie := movieModel{}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(moviePath)))
	if err := db.Update(func(txn *badger.Txn) error {
		stat, err := os.Stat(moviePath)
		if err != nil {
			return err
		}
		v, err := json.Marshal(movieModel{
			Id:               hash,
			Path:             moviePath,
			LastModifiedTime: stat.ModTime().UTC(),
		})
		if err != nil {
			return err
		}
		return txn.Set([]byte("movie|"+hash), v)
	}); err != nil {
		return movie, err
	}
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("movie|" + hash))
		if err != nil {
			return err
		}
		if err := item.Value(func(v []byte) error {
			return json.Unmarshal(v, &movie)
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return movie, err
	}
	return movie, nil
}

func init() {
	rootCmd.AddCommand(addMovieCmd)
	addMovieCmd.Flags().StringVar(&databasePath, "database", "", "database path")
	addMovieCmd.MarkFlagRequired("database")
	addMovieCmd.Flags().StringVar(&moviePath, "movie", "", "movie path")
	addMovieCmd.MarkFlagRequired("movie")
}
