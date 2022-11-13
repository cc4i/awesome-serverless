package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var quit = make(chan bool)
var sf = false
var uid string
var start time.Time
var pingCount int

// Initial things
func init() {

	// Initial parameters
	uid = uuid.New().String()
	start = time.Now()
	pingCount = 0

	// Retrieve Instance ID if Cloud Run
	// TODO: Instance ID -> curl http://metadata.google.internal/computeMetadata/v1/instance/id -H "Metadata-Flavor: Google"

}

func router(ctx context.Context, r *gin.Engine) *gin.Engine {

	r.GET("/ping", ping(ctx))
	r.GET("/bite", biteCPU)
	r.GET("/bite/:timeout", biteCPU)
	r.GET("/stop", stop)
	r.POST("/write2bt/:table", write2bt)
	return r
}

// Test write fixed testing data into a table in Bigtable
func write2bt(c *gin.Context) {

	// greetings := []string{"Hello Cloud Run!", "Hello Cloud Bigtable!", "Hello golang!", "Hello GDT"}
	// columnFamilyName1 := "cf1"
	// columnFamilyName2 := "cf2"
	// columnName := "greeting"

	// ctx := c.Request.Context()

	// tbl := client.Open("mytable")
	// muts := make([]*bigtable.Mutation, len(greetings))
	// rowKeys := make([]string, len(greetings))

	// log.Printf("Writing greeting rows to table")
	// id := uuid.New()
	// for i, greeting := range greetings {
	// 	muts[i] = bigtable.NewMutation()
	// 	muts[i].Set(columnFamilyName1, columnName, bigtable.Now(), []byte(greeting))
	// 	muts[i].Set(columnFamilyName2, columnName, bigtable.Now(), []byte(greeting))

	// 	// Each row has a unique row key.
	// 	//
	// 	// Note: This example uses sequential numeric IDs for simplicity, but
	// 	// this can result in poor performance in a production application.
	// 	// Since rows are stored in sorted order by key, sequential keys can
	// 	// result in poor distribution of operations across nodes.
	// 	//
	// 	// For more information about how to design a Bigtable schema for the
	// 	// best performance, see the documentation:
	// 	//
	// 	//     https://cloud.google.com/bigtable/docs/schema-design

	// 	rowKeys[i] = fmt.Sprintf("%s_%s_%s_%d_%s", columnFamilyName1, columnFamilyName2, columnName, i, id.String())
	// }

	// rowErrs, err := tbl.ApplyBulk(ctx, rowKeys, muts)
	// if err != nil {
	// 	log.Fatalf("Could not apply bulk row mutation: %v", err)
	// }
	// if rowErrs != nil {
	// 	for _, rowErr := range rowErrs {
	// 		log.Printf("Error writing row: %v", rowErr)
	// 	}
	// 	log.Fatalf("Could not write some rows")
	// }

	// log.Printf("Getting a single greeting by row key:")
	// row, err := tbl.ReadRow(ctx, rowKeys[0], bigtable.RowFilter(bigtable.ColumnFilter(columnName)))
	// if err != nil {
	// 	log.Fatalf("Could not read row with key %s: %v", rowKeys[0], err)
	// }
	// log.Printf("\t%s = %s\n", rowKeys[0], string(row[columnFamilyName1][0].Value))
	// c.JSON(http.StatusOK, row)
}

func ping(ctx context.Context) gin.HandlerFunc {
	// str := fmt.Sprintf("Id: %s", c.Value("Id"))
	// c.String(http.StatusOK, str)
	fn := func(c *gin.Context) {
		pingCount++
		str := fmt.Sprintf("Id: %s, elapsed %s, pingCount is %d\n", ctx.Value("Id"), time.Since(start), pingCount)
		log.Println(str)
		c.String(http.StatusOK, str)
	}
	return gin.HandlerFunc(fn)
}

func stop(c *gin.Context) {
	n := runtime.NumCPU()

	for i := 0; i < n; i++ {
		quit <- true
	}
	sf = true
	fmt.Println("biteCPU() has been cancelled immediately and the routine will exit after sleep.")
	c.String(http.StatusOK, "Stopped")
}

func biteCPU(c *gin.Context) {
	timeout := c.Param("timeout")
	tnt, err := strconv.Atoi(timeout)
	if err != nil {
		tnt = 30
		fmt.Printf("Using default duration -> %ds\n", tnt)
	}
	n := runtime.NumCPU()
	fmt.Printf("runtime.NumCPU -> %d\n", n)
	runtime.GOMAXPROCS(n)

	for i := 0; i < n; i++ {
		go func() {
			for {
				select {
				case <-quit:
					return
				default:
					// fmt.Print("\033[u\033[K") // restore the cursor position and clear the line
					// fmt.Printf(".%d.", n)
				}
			}
		}()
	}

	time.Sleep(time.Duration(tnt) * time.Second)
	for i := 0; i < n && !sf; i++ {
		quit <- true
	}
	sf = false
	c.String(http.StatusOK, fmt.Sprintf("Stopped after %ds", tnt))
}

func main() {

	ctx := context.WithValue(context.Background(), "Id", uid)
	go httpSvr(ctx)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	s := <-shutdown
	log.Printf("Signal is %s, %s is terminating after %s\n", s, ctx.Value("Id"), time.Since(start))
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer handleTermination(cancel)
}

func handleTermination(cancel context.CancelFunc) {
	log.Printf("Terminated, duration was %s\n", time.Since(start))
}

func httpSvr(ctx context.Context) {
	gin.DisableConsoleColor()
	server := gin.Default()
	log.Fatal(router(ctx, server).Run("0.0.0.0:9000"))

}
