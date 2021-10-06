package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// listMoviesCmd represents the listMovies command
var listMoviesCmd = &cobra.Command{
	Use:   "list-movies",
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
		movies, err := listMoviesRun(db)
		if err != nil {
			log.Fatalln(err)
		}
		for _, m := range movies {
			fmt.Println(json.Marshal(m))
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

func listMoviesRun(db *badger.DB) ([]movieModel, error) {
	movies := make([]movieModel, 0)
	if err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("movie|")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := it.Item().Value(func(v []byte) error {
				m := movieModel{}
				if err := json.Unmarshal(v, &m); err != nil {
					return err
				}
				movies = append(movies, m)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return movies, nil
}

func init() {
	rootCmd.AddCommand(listMoviesCmd)
	listMoviesCmd.Flags().StringVar(&databasePath, "database", "", "database path")
	listMoviesCmd.MarkFlagRequired("database")
}
