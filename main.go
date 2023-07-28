package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/gorilla/mux"
)

var SOURCES = map[string]source{}
var DB *sqlx.DB

func main() {
	readConfs()

	readSources()
	log.Printf("found %d sources", len(SOURCES))

	initDbPool()

	startListen()
}

func initDbPool() {
	dburl := viper.GetString("dburl")
	pgConfig, err := pgxpool.ParseConfig(dburl)
	if err != nil {
		log.Fatalf("connected to %s failed: %s", dburl, err)
	}
	DB, err = sqlx.Connect("postgres", viper.GetString("dburl"))
	if err != nil {
		log.Fatal(err)
	}
	dbName := pgConfig.ConnConfig.Config.Database
	dbUser := pgConfig.ConnConfig.Config.User
	dbHost := pgConfig.ConnConfig.Config.Host
	log.Printf("Connected as '%s' to '%s' @ '%s'", dbUser, dbName, dbHost)
}

func startListen() {
	r := mux.NewRouter()

	r.HandleFunc("/list", handleList)
	r.HandleFunc("/{id}/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}", handleZXY)
	r.HandleFunc("/{id}", handleTilejson)
	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port")),
		Handler: r,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Printf("listening now on: \"%s:%d\"", viper.GetString("host"), viper.GetInt("port"))
	// wait here for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	// Interrupt signal received:  Start shutting down
	log.Printf("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 1000)
	defer cancel()
	err := s.Shutdown(ctx)

	if DB != nil {
		log.Printf("Closing DB connections")
		err := DB.Close()
		if err != nil {
			//ignore the server would
			log.Printf("postgres connection close failed, %s", err)
		}
	}

	if err != nil {
		log.Fatalf("Server stopped unexpectedly: %s", err)
	}
	log.Printf("Server stopped.")
}

func handleTilejson(writer http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]
	s := SOURCES[id]
	if s == nil {
		http.NotFound(writer, request)
	}
	tj := s.Tilejson()
	url := request.URL
	tj.tiles = []string{
		fmt.Sprintf("%s/%s/{z}/{x}/{y}", url, id),
	}
	json.NewEncoder(writer).Encode(s.Tilejson())
}

func handleZXY(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]
	s := SOURCES[id]
	if s == nil {
		http.NotFound(w, r)
	}

	z, _ := strconv.Atoi(vars["z"])
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])

	tile, err := s.Tile(z, x, y)
	if err != nil {
		log.Printf("getting tile failed: zoom %d x %d y %d @ %s,error: %s", z, x, y, id, err)
	}
	w.Header().Add("Content-Type", "application/vnd.mapbox-vector-tile")
	_, err1 := w.Write(tile)
	if err1 != nil {
		log.Printf("getting tile failed: zoom %d x %d y %d @ %s,error: %s\n", z, x, y, id, err)
	}
}

func handleList(writer http.ResponseWriter, _ *http.Request) {
	log.Printf("requested \"/\"")
	var list []string
	for k := range SOURCES {
		list = append(list, k)
	}
	err := json.NewEncoder(writer).Encode(list)
	// how could I return a 500?
	if err != nil {
		http.Error(writer, "", 500)
	}
}

func readSources() {
	recipeDir := viper.GetString("recipeDir")
	log.Printf("tring to find recipes in %s", recipeDir)
	_, err := os.Stat(recipeDir)
	if err != nil {
		log.Printf("%s is not exist,skipped", recipeDir)
		return
	}

	filepath.Walk(recipeDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(info.Name()) != ".yml" {
			return nil
		}
		if info.Size() == 0 {
			log.Printf("%s is a empty file,skipped", info.Name())
		}
		absName := filepath.Join(recipeDir, info.Name())
		f, err := ioutil.ReadFile(absName)
		if err != nil {
			log.Printf("open and read %s failed, skipped", err)
			return err
		}
		tempPG := pg{}
		err = yaml.Unmarshal(f, &tempPG)
		if err != nil {
			log.Printf("try read contents from %s as a pg tile recipe failed with error %v, skipped", absName, err)
		}

		_, ok := SOURCES[tempPG.Name]
		if ok {
			log.Printf("duplicated source id, skipped: %s", info.Name())
			return nil
		}

		SOURCES[tempPG.Name] = source(tempPG)
		log.Printf("found source \"%s\": id: %s minzoom: %d maxzoom: %d", tempPG.Name, tempPG.Name, tempPG.MinZoom, tempPG.MaxZoom)
		return nil
	})
	return
}

func readConfs() {
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if strings.HasPrefix(arg, "postgres://") {
				viper.Set("dburl", arg)
				break
			}
		}
	}

	// set defaults
	viper.SetDefault("conf", "./conf.yml")
	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", 3000)
	viper.SetDefault("recipeDir", "./recipes")

	// env
	viper.SetEnvPrefix("got")
	viper.AutomaticEnv()

	// command line args
	pflag.String("conf", "", "path of config file")
	pflag.String("host", "0.0.0.0", "host which will be bind")
	pflag.Int("port", 3000, "host which would be bind")
	pflag.String("dburl", "", "postgres database url")
	pflag.String("recipedir", "./recipe", "recipe for scanning postgis tiles recipes")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		//ignore this shouldn't  happend
		log.Printf("tring to bind commandline flags failed,this shouldn't happened: %s", err)
	}

	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	err = viper.ReadInConfig()
	if err != nil {
		log.Print("init without config file")
	}
}
