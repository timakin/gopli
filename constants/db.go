package constants

const (
	MySQLSession = "mysql -u%s -p%s -h%s"

	SelectTablesCmd = "mysql -u%s -p%s -B -N -e 'SELECT * FROM %s.%s'"
	ShowTableCmd    = "mysql %s -u%s -p%s -B -N -e 'show tables'"

	DeleteTableCmd            = "mysql -u%s -p%s -B -N -e 'DELETE FROM %s.%s'"
	DeleteTableCmdWithoutPass = "mysql -u%s -B -N -e 'DELETE FROM %s.%s'"

	DefaultOffset = 1000000000

	DeleteTableQuery    = "DELETE FROM %s.%s"
	LoadInfileQuery     = "LOAD DATA LOCAL INFILE '%s' INTO TABLE %s.%s"
	DstHostMysqlConnect = "%s:%s@tcp(%s:%s)/%s"
)
