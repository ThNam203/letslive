package hack

import (
	"sen1or/lets-live/server/domain"

	"gorm.io/gorm"
)

func AutoMigrateAllTables(dbConn gorm.DB) error {
	migrator := dbConn.Migrator()

	err := migrator.AutoMigrate(&domain.User{}, &domain.RefreshToken{})
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
