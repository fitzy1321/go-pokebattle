package main

import (
	"fmt"
	"os"

	"pogomon/consts"
	"pogomon/mvu"
	"pogomon/setup"
	"pogomon/utils"

	tea "charm.land/bubbletea/v2"
	"gorm.io/gorm"
)

func printErrExit(errs ...error) {
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "Error:: %+v\n", e)
	}
	os.Exit(1)
}

func main() {
	// TODO * fix dbFilePath for XDG and OS specific locations later
	// dataDirPath, dErr := utils.GetDataDirPath()
	// if dErr != nil {
	// 	printErrExit(dErr)
	// }

	// dbFilePath := filepath.Join(dataDirPath, consts.DBFILEPATH)
	_, dErr := utils.GetDataDirPath()
	if dErr != nil {
		printErrExit(dErr)
	}
	dbFilePath := consts.DBFILEPATH
	var gdb *gorm.DB = nil

	if !utils.FileExists(dbFilePath) {
		db, errs := setup.FetchDataAndCreateDB(dbFilePath)
		if errs != nil || len(errs) > 0 {
			printErrExit(errs...)
		}
		gdb = db
	} else {
		db, err := setup.GetGormSqliteDB(dbFilePath)
		if err != nil {
			printErrExit(fmt.Errorf("Something failed connecting to pokemon db: %v\n", err))
		}
		gdb = db
	}

	model, err := mvu.NewAppModel(gdb)
	if err != nil {
		printErrExit(err)
	}

	// fmt.Print("> ")
	// fmt.Scanln()

	p := tea.NewProgram(*model)
	if _, err := p.Run(); err != nil {
		printErrExit(err)
	}
}
