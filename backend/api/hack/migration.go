package hack

import (
	"sen1or/lets-live/api/domains"

	"gorm.io/gorm"
)

func AutoMigrateAllTables(dbConn gorm.DB) error {
	migrator := dbConn.Migrator()

	err := migrator.AutoMigrate(&domains.User{}, &domains.RefreshToken{}, &domains.VerifyToken{})
	if err != nil {
		return err
	}

	return nil
}

//func (mm *MyMigrator) RecreateDatabase() {
//	migrator := mm.dbConn.Migrator()
//
//	err := migrator.DropTable(&domain.User{})
//	if err != nil {
//		return err
//	}
//
//	return nil
//
//}
