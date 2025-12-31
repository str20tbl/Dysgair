package controllers

import (
	"database/sql"
	"fmt"
	"strings"

	"app/app/models"
	"app/appJobs"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql" // Required for db connectivity
	"github.com/str20tbl/modules/jobs/app/jobs"
	"github.com/str20tbl/revel"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	revel.OnAppStart(InitDB)
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*AuthController).Before, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)
	revel.TemplateFuncs["get"] = func(a map[string]int, b string) int {
		return a[b]
	}

	revel.TemplateFuncs["format"] = func(input interface{}) (out string) {
		out = fmt.Sprintf("%+v", input)
		return
	}

	revel.TemplateFuncs["recorded"] = func(id int64, uid string) bool {
		count, err := Dbm.SelectInt("SELECT COUNT(*) FROM Recording WHERE custom_prompt_id = ? && uid = ?", id, uid)
		if err != nil {
			revel.AppLog.Error("Unable to check recording", err)
		}
		return count > 0
	}

	revel.TemplateFuncs["add"] = func(a, b int) int {
		return a + b
	}

	revel.TemplateFuncs["sub"] = func(a, b int) int {
		return a - b
	}

	revel.TemplateFuncs["ef"] = func(a, b interface{}) bool {
		if _, ok := a.(string); !ok {
			return true
		}
		return strings.EqualFold(a.(string), b.(string))
	}

	revel.TemplateFuncs["nef"] = func(a, b interface{}) bool {
		if _, ok := a.(string); !ok {
			return false
		}
		return !strings.EqualFold(a.(string), b.(string))
	}

	revel.TemplateFuncs["ia"] = func(a interface{}) bool {
		if _, ok := a.(int); !ok {
			return true
		}
		return a.(int) > 0
	}

	revel.TemplateFuncs["percent"] = func(input, total int64) string {
		sum := 0.0
		if total > 0 {
			sum = float64(input) / float64(total) * 100
		}
		return fmt.Sprintf("%.0f", sum)
	}
}

var tables = map[string]interface{}{
	"User":  models.User{},
	"Entry": models.Entry{},
	"Word":  models.Word{},
}

// InitDB Init the db object
func InitDB() {
	connectString := "dysgair:cjJaLbuCXGf9XJn94h9S3bes@tcp(db:3306)/dysgair?charset=utf8mb4&parseTime=True"
	db, err := sql.Open("mysql", connectString)
	if err == nil {
		Dbm = &gorp.DbMap{
			Db:      db,
			Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"},
		}
		for _, table := range tables {
			Dbm.AddTable(table).SetKeys(true, "id")
		}
		err = Dbm.CreateTablesIfNotExists()
		if err != nil {
			revel.AppLog.Error(err.Error())
		}
		revel.AppLog.Info("migrateTables(Dbm)")
		migrateTables(Dbm)
		revel.AppLog.Info("makeAdmin(Dbm)")
		makeAdmin(Dbm)
		//revel.AppLog.Info("LoadWords(Dbm)")
		//err = models.LoadWords(Dbm)
		//if err != nil {
		//	revel.AppLog.Error(err.Error())
		//}
		revel.AppLog.Info("Pre-loading TTS audio files...")
		jobs.Now(appJobs.TTSPreloader{Dbm: Dbm})
		revel.AppLog.Info("App Init Complete")
	}
}

func makeAdmin(dbm *gorp.DbMap) {
	query := "SELECT COUNT(*) FROM User WHERE Username = 'Admin'"
	hasAdmin, err := dbm.SelectInt(query)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	if hasAdmin == 0 {
		var hashedPassword []byte
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte("cjJaLbuCXGf9XJn94h9S3bes"), bcrypt.DefaultCost)
		if err != nil {
			revel.AppLog.Error(err.Error())
		}
		err = dbm.Insert(&models.User{
			Username:       "Admin",
			FirstName:      "Admin",
			LastName:       "User",
			HashedPassword: hashedPassword,
			UserType:       1,
		})
		if err != nil {
			revel.AppLog.Error(err.Error())
		}
	}
}

func migrateTables(dbm *gorp.DbMap) {
	tableData := make(map[string][]interface{})
	migrationTables := revel.Config.StringDefault("migration.tables", "")
	for name, table := range tables {
		if strings.Contains(migrationTables, name) {
			tableData[name] = migrateTable(dbm, table, name)
		}
		dbm.AddTable(table).SetKeys(true, "id")
	}
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	reimportData(dbm, tableData)
}

func reimportData(dbm *gorp.DbMap, data map[string][]interface{}) {
	migrationTables := revel.Config.StringDefault("migration.tables", "")
	for key, dat := range data {
		if len(dat) > 0 && strings.Contains(migrationTables, key) {
			populateData(dbm, key, dat)
		}
	}
	revel.AppLog.Infof("============>> %s", migrationTables)
}

func populateData(dbm *gorp.DbMap, key string, dat []interface{}) {
	switch key {
	case "User":
		for _, interfaceData := range dat {
			_ = dbm.Insert(interfaceData.(*models.User))
		}
	case "Entry":
		for _, interfaceData := range dat {
			_ = dbm.Insert(interfaceData.(*models.Entry))
		}
	case "Word":
		for _, interfaceData := range dat {
			_ = dbm.Insert(interfaceData.(*models.Word))
		}
	}
}

func migrateTable(dbm *gorp.DbMap, table interface{}, tableName string) []interface{} {
	queryFormat := "SELECT * FROM %s"
	query := fmt.Sprintf(queryFormat, tableName)
	tableData, err := dbm.Select(table, query)
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	_, err = dbm.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		revel.AppLog.Error(err.Error())
	}
	return tableData
}
