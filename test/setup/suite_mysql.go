package setup

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/suite"

	"github.com/evgeniy-klemin/todo/db/schema"
)

type MySQLSuite struct {
	suite.Suite
	DB         *sql.DB
	FTSEnabled bool
}

func (s *MySQLSuite) SetupSuite() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "todo:todo@tcp(localhost:3306)/todotest?parseTime=true"
	}

	db, err := sql.Open(schema.DriverMySQL, dsn)
	if err != nil {
		s.T().Skipf("MySQL not available: %v", err)
	}

	if err := db.Ping(); err != nil {
		s.T().Skipf("MySQL not available: %v", err)
	}

	s.DB = db
	ftsEnabled, err := schema.ApplyAll(db, schema.DriverMySQL)
	s.Require().NoError(err)
	s.FTSEnabled = ftsEnabled
}

func (s *MySQLSuite) TearDownSuite() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *MySQLSuite) SetupTest() {
	if s.DB != nil {
		_, err := s.DB.Exec("DELETE FROM item")
		s.Require().NoError(err)
	}
}
