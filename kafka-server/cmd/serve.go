package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"strconv"
)

var serverPort int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
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
		r := gin.Default()
		c := cors.DefaultConfig()
		c.AllowAllOrigins = true
		r.Use(cors.New(c))
		r.GET("/movie", func(context *gin.Context) {
			movies, err := listMoviesRun(db)
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			context.JSON(http.StatusOK, movies)
		})
		r.POST("/movie", func(context *gin.Context) {
			var data struct {
				Path string `json:"path" binding:"required"`
			}
			if err := context.ShouldBindJSON(&data); err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			moviePath = data.Path
			movie, err := addMovieRun(db)
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			context.JSON(http.StatusOK, movie)
		})
		r.DELETE("/movie/:id", func(context *gin.Context) {
			movieId = context.Param("id")
			movie := movieModel{}
			key := []byte(fmt.Sprintf("movie|%s", movieId))
			if err := db.View(func(txn *badger.Txn) error {
				item, err := txn.Get(key)
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					return json.Unmarshal(val, &movie)
				})
			}); err != nil {
				if err == badger.ErrKeyNotFound {
					context.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
				} else {
					context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
				return
			}
			if err := db.Update(func(txn *badger.Txn) error {
				if err := txn.Delete(key); err != nil {
					return err
				}
				return os.Remove(movie.Path)
			}); err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			context.JSON(http.StatusOK, nil)
		})
		r.GET("/movie/:id/thumbnail", func(context *gin.Context) {
			movieId = context.Param("id")
			count, err := getThumbnailsCount(db)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					context.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
				} else {
					context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
				return
			}
			context.JSON(http.StatusOK, gin.H{"count": count})
		})
		r.GET("/movie/:id/thumbnail/:n", func(context *gin.Context) {
			movieId = context.Param("id")
			n, err := strconv.Atoi(context.Param("n"))
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			data, err := getThumbnailsAt(db, n)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					context.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
				} else {
					context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
				return
			}
			context.Data(http.StatusOK, "image/png", data)
		})
		r.POST("/movie/:id/thumbnail", func(context *gin.Context) {
			var data struct {
				ThumbnailCount int `json:"thumbnailCount" binding:"required"`
			}
			if err := context.ShouldBindJSON(&data); err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			movieId = context.Param("id")
			thumbnailCount = data.ThumbnailCount
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			movie, err := createThumbnailsRun(db)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					context.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
				} else {
					context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
				return
			}
			context.JSON(http.StatusOK, movie)
		})
		r.GET("/health", func(context *gin.Context) {
			context.Status(http.StatusOK)
		})
		if err := r.Run(fmt.Sprintf(":%d", serverPort)); err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&databasePath, "database", "", "database path")
	serveCmd.MarkFlagRequired("database")
	serveCmd.Flags().IntVar(&serverPort, "port", 8080, "Listen port")
}
