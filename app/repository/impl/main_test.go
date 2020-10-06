package impl

import (
	"database/sql"
	"github.com/belito3/go-api-codebase/app/repository"

	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDrive  = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testDB *sql.DB
var testAccountImpl repository.IAccount
var testEntryImpl repository.IEntry
var testTransferImpl repository.ITransfer

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDrive, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	testAccountImpl = NewAccountImpl(testDB)
	testEntryImpl = NewEntryImpl(testDB)
	testTransferImpl = NewTransferImpl(testDB)

	os.Exit(m.Run())
}
