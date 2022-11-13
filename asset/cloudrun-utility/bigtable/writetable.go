package bigtable

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/bigtable"
	"github.com/google/uuid"
)

var client *bigtable.Client

func InitClient(projectId string, instance string) {

	if client == nil {
		if clt, err := bigtable.NewClient(context.TODO(), projectId, instance); err != nil {
			log.Fatalf("Could not create admin client: %v", err)
		} else {
			client = clt
		}

	}

}

func Write2Bt() {

	greetings := []string{"Hello Cloud Run!", "Hello Cloud Bigtable!", "Hello golang!", "Hello GDT"}
	columnFamilyName1 := "cf1"
	columnFamilyName2 := "cf2"
	columnName := "greeting"

	tbl := client.Open("mytable")
	muts := make([]*bigtable.Mutation, len(greetings))
	rowKeys := make([]string, len(greetings))

	log.Printf("Writing greeting rows to table")
	id := uuid.New()
	for i, greeting := range greetings {
		muts[i] = bigtable.NewMutation()
		muts[i].Set(columnFamilyName1, columnName, bigtable.Now(), []byte(greeting))
		muts[i].Set(columnFamilyName2, columnName, bigtable.Now(), []byte(greeting))

		// Each row has a unique row key.
		//
		// Note: This example uses sequential numeric IDs for simplicity, but
		// this can result in poor performance in a production application.
		// Since rows are stored in sorted order by key, sequential keys can
		// result in poor distribution of operations across nodes.
		//
		// For more information about how to design a Bigtable schema for the
		// best performance, see the documentation:
		//
		//     https://cloud.google.com/bigtable/docs/schema-design

		rowKeys[i] = fmt.Sprintf("%s_%s_%s_%d_%s", columnFamilyName1, columnFamilyName2, columnName, i, id.String())
	}

	rowErrs, err := tbl.ApplyBulk(context.TODO(), rowKeys, muts)
	if err != nil {
		log.Fatalf("Could not apply bulk row mutation: %v", err)
	}
	if rowErrs != nil {
		for _, rowErr := range rowErrs {
			log.Printf("Error writing row: %v", rowErr)
		}
		log.Fatalf("Could not write some rows")
	}

	log.Printf("Getting a single greeting by row key:")
	row, err := tbl.ReadRow(ctx, rowKeys[0], bigtable.RowFilter(bigtable.ColumnFilter(columnName)))
	if err != nil {
		log.Fatalf("Could not read row with key %s: %v", rowKeys[0], err)
	}
	log.Printf("\t%s = %s\n", rowKeys[0], string(row[columnFamilyName1][0].Value))
	c.JSON(http.StatusOK, row)
}
