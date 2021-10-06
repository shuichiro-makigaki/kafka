package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var movieId string

// getThumbnailsCmd represents the getThumbnails command
var getThumbnailsCmd = &cobra.Command{
	Use:   "get-thumbnails",
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
		images, err := getThumbnailsRun(db)
		if err != nil {
			log.Fatalln(err)
		}
		for _, v := range images {
			b := base64.StdEncoding.EncodeToString(v)
			fmt.Println(b)
		}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if databasePath, err = filepath.Abs(filepath.Clean(databasePath)); err != nil {
			return err
		}
		if _, err = os.Stat(databasePath); err != nil {
			return err
		}
		return nil
	},
}

func getThumbnailsRun(db *badger.DB) ([][]byte, error) {
	images := make([][]byte, 0)
	if err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("thumbnail|%s|", movieId))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				images = append(images, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return images, nil
}

func getThumbnailsAt(db *badger.DB, n int) ([]byte, error) {
	image := make([]byte, 0)
	if err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("thumbnail|%s|%d", movieId, n)))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			image = val
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return image, nil
}

func getThumbnailsCount(db *badger.DB) (int, error) {
	count := 0
	if err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("thumbnail|%s|", movieId))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	}); err != nil {
		return -1, err
	}
	return count, nil
}

func init() {
	rootCmd.AddCommand(getThumbnailsCmd)
	getThumbnailsCmd.Flags().StringVar(&databasePath, "database", "", "database path")
	getThumbnailsCmd.MarkFlagRequired("database")
	getThumbnailsCmd.Flags().StringVar(&movieId, "movieid", "", "movieModel ID")
	getThumbnailsCmd.MarkFlagRequired("movieid")
}
