package sqlxtraced

import (
	"log"
	"testing"

	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/DataDog/dd-trace-go/tracer/contrib/sqltraced/sqlutils"
	"github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {
	trc, transport := tracer.GetTestTracer()
	dbx, err := OpenTraced(&mysql.MySQLDriver{}, "ubuntu@tcp(127.0.0.1:3306)/circle_test", "mysql-test", trc)
	if err != nil {
		log.Fatal(err)
	}
	defer dbx.Close()

	testDB := &sqlutils.DB{
		DB:         dbx.DB,
		Tracer:     trc,
		Transport:  transport,
		DriverName: "mysql",
	}

	expectedSpan := &tracer.Span{
		Name:    "mysql.query",
		Service: "mysql-test",
		Type:    "sql",
	}
	expectedSpan.Meta = map[string]string{
		"db.user":  "ubuntu",
		"out.host": "127.0.0.1",
		"out.port": "3306",
		"db.name":  "circle_test",
	}

	sqlutils.AllSQLTests(t, testDB, expectedSpan)
}
